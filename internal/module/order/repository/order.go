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

func (r *OrderRepo) CreateOrder(ctx context.Context, order *domain.Order, cartID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

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
	_, err = tx.Exec(ctx, orderQ, order.ID, order.UserID, order.RestaurantID, order.Status, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	itemQ := `
		INSERT INTO order_items (id, order_id, dish_id, dish_name_snapshot, price_snapshot, quantity)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, itemQ, uuid.New(), order.ID, item.DishID, item.DishNameSnapshot, item.PriceSnapshot, item.Quantity)
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
