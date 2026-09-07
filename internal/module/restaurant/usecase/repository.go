package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

// ListFilter represents filtering and pagination options for listing restaurants.
type ListFilter struct {
	Status *domain.RestaurantStatus
	Limit  int
	Offset int
}

// RestaurantRepository defines the persistence operations for Restaurant aggregates.
type RestaurantRepository interface {
	Create(ctx context.Context, restaurant *domain.Restaurant) (*domain.Restaurant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RestaurantStatus) (*domain.Restaurant, error)
	List(ctx context.Context, filter ListFilter) ([]domain.Restaurant, int, error)
}
