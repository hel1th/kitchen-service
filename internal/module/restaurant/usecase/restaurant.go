package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

type RestaurantUsecase struct {
	repo RestaurantRepository
}

func NewRestaurantUsecase(repo RestaurantRepository) *RestaurantUsecase {
	return &RestaurantUsecase{repo: repo}
}

func (uc *RestaurantUsecase) SetStatus(ctx context.Context, id uuid.UUID, status domain.RestaurantStatus) (*domain.Restaurant, error) {
	if !status.IsValid() {
		return nil, domain.ErrInvalidRestaurantStatus
	}
	return uc.repo.UpdateStatus(ctx, id, status)
}

func (uc *RestaurantUsecase) List(ctx context.Context, filter ListFilter) ([]domain.Restaurant, int, error) {
	return uc.repo.List(ctx, filter)
}

func (uc *RestaurantUsecase) Get(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetRestaurant implements RestaurantProvider
func (uc *RestaurantUsecase) GetRestaurant(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	return uc.Get(ctx, id)
}

// IsOpen implements RestaurantProvider
func (uc *RestaurantUsecase) IsOpen(ctx context.Context, id uuid.UUID) (bool, error) {
	r, err := uc.Get(ctx, id)
	if err != nil {
		return false, err
	}
	return r.IsOpen(), nil
}
