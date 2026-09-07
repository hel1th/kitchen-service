package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

// RestaurantProvider defines the cross-module contract for accessing restaurant state.
// Implemented by restaurant/usecase and consumed by order/usecase during cart operations and checkout.
type RestaurantProvider interface {
	GetRestaurant(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error)
	IsOpen(ctx context.Context, id uuid.UUID) (bool, error)
}
