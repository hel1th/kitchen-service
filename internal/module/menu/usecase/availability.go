package usecase

import (
	"context"

	"github.com/google/uuid"
)

// UnavailableDishReason describes why a dish is unavailable for ordering.
type UnavailableDishReason string

const (
	ReasonUnavailable UnavailableDishReason = "unavailable"
	ReasonDeleted     UnavailableDishReason = "deleted"
	ReasonNotFound    UnavailableDishReason = "not_found"
)

// UnavailableDish represents a dish that cannot be ordered.
type UnavailableDish struct {
	DishID uuid.UUID
	Reason UnavailableDishReason
}

// DishInfo contains live menu information for cart display and order snapshots.
type DishInfo struct {
	ID           uuid.UUID
	RestaurantID uuid.UUID
	CategoryID   uuid.UUID
	Name         string
	Price        float64
	Available    bool
	DeletedAt    *bool
}

// AvailabilityChecker defines the cross-module contract for checking dish availability.
// Implemented by menu/usecase and consumed by order/usecase during checkout.
// It returns the full list of unavailable dishes rather than failing on the first one.
type AvailabilityChecker interface {
	CheckAvailability(ctx context.Context, restaurantID uuid.UUID, dishIDs []uuid.UUID) ([]UnavailableDish, error)
}

// MenuDishProvider provides dish details needed across module boundaries
type MenuDishProvider interface {
	GetDish(ctx context.Context, dishID uuid.UUID) (*DishInfo, error)
	GetDishesByIDs(ctx context.Context, dishIDs []uuid.UUID) (map[uuid.UUID]DishInfo, error)
}
