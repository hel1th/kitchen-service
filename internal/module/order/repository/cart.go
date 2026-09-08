package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
	"github.com/hel1th/kitchen-service/internal/module/order/usecase"
	"github.com/hel1th/kitchen-service/internal/shared/database"
)

var _ usecase.CartRepository = (*CartRepo)(nil)

type CartRepo struct {
	db database.DBTX
}

func NewCartRepo(db database.DBTX) *CartRepo {
	return &CartRepo{db: db}
}

const resetQ = "UPDATE carts SET restaurant_id = NULL WHERE id = $1"

func (r *CartRepo) GetByUserID(ctx context.Context, userID int64) (*domain.Cart, error) {
	query := `SELECT id, user_id, restaurant_id FROM carts WHERE user_id = $1`

	var cart domain.Cart
	err := r.db.QueryRow(ctx, query, userID).Scan(&cart.ID, &cart.UserID, &cart.RestaurantID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get cart by user id: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		cart.ID = uuid.New()
		cart.UserID = userID
		insertQ := `INSERT INTO carts (id, user_id) VALUES ($1, $2)`
		_, err = r.db.Exec(ctx, insertQ, cart.ID, cart.UserID)
		if err != nil {
			return nil, fmt.Errorf("create cart: %w", err)
		}
	}

	itemsQ := `SELECT id, dish_id, quantity FROM cart_items WHERE cart_id = $1`
	rows, err := r.db.Query(ctx, itemsQ, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("query cart items: %w", err)
	}
	defer rows.Close()

	cart.Items = make([]domain.CartItem, 0)
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(&item.ID, &item.DishID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("scan cart item: %w", err)
		}
		cart.Items = append(cart.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cart items: %w", err)
	}

	return &cart, nil
}

func (r *CartRepo) AddItem(ctx context.Context, cartID, dishID uuid.UUID, quantity int) error {
	query := `
		INSERT INTO cart_items (id, cart_id, dish_id, quantity)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (cart_id, dish_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	`
	_, err := r.db.Exec(ctx, query, uuid.New(), cartID, dishID, quantity)
	if err != nil {
		return fmt.Errorf("add item to cart: %w", err)
	}
	return nil
}

func (r *CartRepo) SetRestaurant(ctx context.Context, cartID, restaurantID uuid.UUID) error {
	query := `UPDATE carts SET restaurant_id = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, cartID, restaurantID)
	if err != nil {
		return fmt.Errorf("set cart restaurant: %w", err)
	}
	return nil
}

func (r *CartRepo) RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1 AND id = $2`
	tag, err := r.db.Exec(ctx, query, cartID, itemID)
	if err != nil {
		return fmt.Errorf("remove cart item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCartItemNotFound
	}

	checkQ := `SELECT COUNT(*) FROM cart_items WHERE cart_id = $1`
	var count int
	if err := r.db.QueryRow(ctx, checkQ, cartID).Scan(&count); err == nil && count == 0 {
		_, _ = r.db.Exec(ctx, resetQ, cartID)
	}

	return nil
}

func (r *CartRepo) UpdateItemQuantity(ctx context.Context, cartID, itemID uuid.UUID, quantity int) error {
	query := `UPDATE cart_items SET quantity = $3 WHERE cart_id = $1 AND id = $2`
	tag, err := r.db.Exec(ctx, query, cartID, itemID, quantity)
	if err != nil {
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCartItemNotFound
	}
	return nil
}

func (r *CartRepo) Clear(ctx context.Context, cartID uuid.UUID) error {
	clearQ := `DELETE FROM cart_items WHERE cart_id = $1`
	if _, err := r.db.Exec(ctx, clearQ, cartID); err != nil {
		return fmt.Errorf("clear cart items: %w", err)
	}

	if _, err := r.db.Exec(ctx, resetQ, cartID); err != nil {
		return fmt.Errorf("reset cart restaurant: %w", err)
	}
	return nil
}
