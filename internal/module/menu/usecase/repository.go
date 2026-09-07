package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

// CategoryRepository defines the persistence operations for menu categories.
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) (*domain.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetByIDAndRestaurantID(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID) (*domain.Category, error)
	ListByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]domain.Category, error)
	Delete(ctx context.Context, restaurantID uuid.UUID, id uuid.UUID) error
}

// DishRepository defines the persistence operations for dishes.
type DishRepository interface {
	Create(ctx context.Context, dish *domain.Dish) (*domain.Dish, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Dish, error)
	GetByIDAndRestaurantID(ctx context.Context, id uuid.UUID, restaurantID uuid.UUID) (*domain.Dish, error)
	Update(ctx context.Context, dish *domain.Dish) (*domain.Dish, error)
	SoftDelete(ctx context.Context, restaurantID uuid.UUID, id uuid.UUID, at time.Time) error
	Restore(ctx context.Context, restaurantID uuid.UUID, id uuid.UUID, at time.Time) (*domain.Dish, error)
	SetAvailability(ctx context.Context, restaurantID uuid.UUID, id uuid.UUID, available bool, at time.Time) (*domain.Dish, error)
	ListByRestaurantID(ctx context.Context, restaurantID uuid.UUID, includeDeleted bool) ([]domain.Dish, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Dish, error)
	GetMenuWithDishes(ctx context.Context, restaurantID uuid.UUID) ([]domain.CategoryWithDishes, error)
}
