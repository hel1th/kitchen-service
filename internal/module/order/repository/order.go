package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
	"github.com/hel1th/kitchen-service/internal/module/order/usecase"
)

var _ usecase.OrderRepository = (*OrderRepo)(nil)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) CreateOrder(
	ctx context.Context,
	order *domain.Order,
	cartID uuid.UUID,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = order.CreatedAt
	}

	orderQ := `
		INSERT INTO orders (id, user_id, restaurant_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(
		ctx,
		orderQ,
		order.ID,
		order.UserID,
		order.RestaurantID,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	itemQ := `
		INSERT INTO order_items (id, order_id, dish_id, dish_name_snapshot, price_snapshot, quantity)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, item := range order.Items {
		_, err = tx.Exec(
			ctx,
			itemQ,
			uuid.New(),
			order.ID,
			item.DishID,
			item.DishNameSnapshot,
			item.PriceSnapshot,
			item.Quantity,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	clearCartQ := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err = tx.Exec(ctx, clearCartQ, cartID)
	if err != nil {
		return fmt.Errorf("clear cart items: %w", err)
	}

	resetCartQ := `UPDATE carts SET restaurant_id = NULL WHERE id = $1`
	_, err = tx.Exec(ctx, resetCartQ, cartID)
	if err != nil {
		return fmt.Errorf("reset cart restaurant: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT id, user_id, restaurant_id, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	var o domain.Order
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}

	itemsQ := `
		SELECT dish_id, dish_name_snapshot, price_snapshot, quantity
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.pool.Query(ctx, itemsQ, o.ID)
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	o.Items = make([]domain.OrderItem, 0)
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.DishID, &item.DishNameSnapshot, &item.PriceSnapshot, &item.Quantity); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		o.Items = append(o.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}

	return &o, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $2, updated_at = $3
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, status, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepo) ListByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.restaurant_id, o.status, o.created_at, o.updated_at,
		       i.dish_id, i.dish_name_snapshot, i.price_snapshot, i.quantity
		FROM orders o
		LEFT JOIN order_items i ON o.id = i.order_id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC, i.dish_name_snapshot ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list orders by user: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

func (r *OrderRepo) ListByRestaurantID(
	ctx context.Context,
	restaurantID uuid.UUID,
	status *domain.OrderStatus,
) ([]domain.Order, error) {
	var statusStr *string
	if status != nil {
		s := string(*status)
		statusStr = &s
	}

	query := `
		SELECT o.id, o.user_id, o.restaurant_id, o.status, o.created_at, o.updated_at,
		       i.dish_id, i.dish_name_snapshot, i.price_snapshot, i.quantity
		FROM orders o
		LEFT JOIN order_items i ON o.id = i.order_id
		WHERE o.restaurant_id = $1 AND ($2::text IS NULL OR o.status = $2)
		ORDER BY o.created_at DESC, i.dish_name_snapshot ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID, statusStr)
	if err != nil {
		return nil, fmt.Errorf("list orders by restaurant: %w", err)
	}
	defer rows.Close()

	return r.scanOrders(rows)
}

func (r *OrderRepo) scanOrders(rows pgx.Rows) ([]domain.Order, error) {
	orderMap := make(map[uuid.UUID]*domain.Order)
	var orderIDs []uuid.UUID

	for rows.Next() {
		var o domain.Order
		var dishID *uuid.UUID
		var dishName *string
		var price *float64
		var quantity *int

		err := rows.Scan(
			&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.CreatedAt, &o.UpdatedAt,
			&dishID, &dishName, &price, &quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order row: %w", err)
		}

		if _, exists := orderMap[o.ID]; !exists {
			orderMap[o.ID] = &o
			orderMap[o.ID].Items = make([]domain.OrderItem, 0)
			orderIDs = append(orderIDs, o.ID)
		}

		if dishID != nil {
			orderMap[o.ID].Items = append(orderMap[o.ID].Items, domain.OrderItem{
				DishID:           *dishID,
				DishNameSnapshot: *dishName,
				PriceSnapshot:    *price,
				Quantity:         *quantity,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order rows: %w", err)
	}

	result := make([]domain.Order, 0, len(orderIDs))
	for _, id := range orderIDs {
		result = append(result, *orderMap[id])
	}
	return result, nil
}
