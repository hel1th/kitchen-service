package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
)

// CartRepository defines persistence operations for Carts.
type CartRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*domain.Cart, error)
	AddItem(ctx context.Context, cartID, dishID uuid.UUID, quantity int) error
	SetRestaurant(ctx context.Context, cartID, restaurantID uuid.UUID) error
	RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error
	UpdateItemQuantity(ctx context.Context, cartID, itemID uuid.UUID, quantity int) error
	Clear(ctx context.Context, cartID uuid.UUID) error
}

// OrderRepository defines persistence operations for Orders.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order, cartID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	ListByUserID(ctx context.Context, userID int64) ([]domain.Order, error)
	ListByRestaurantID(ctx context.Context, restaurantID uuid.UUID, status *domain.OrderStatus) ([]domain.Order, error)
}

// UserRepository defines persistence operations for Users.
type UserRepository interface {
	Upsert(ctx context.Context, id int64) error
}
