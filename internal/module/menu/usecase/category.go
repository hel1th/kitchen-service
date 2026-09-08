package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

type CategoryUsecase struct {
	repo CategoryRepository
}

func NewCategoryUsecase(repo CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{repo: repo}
}

func (uc *CategoryUsecase) AddCategory(
	ctx context.Context,
	restaurantID uuid.UUID,
	name string,
	sortOrder int,
) (*domain.Category, error) {
	cat := &domain.Category{
		ID:           uuid.New(),
		RestaurantID: restaurantID,
		Name:         name,
		SortOrder:    sortOrder,
	}
	return uc.repo.Create(ctx, cat)
}

func (uc *CategoryUsecase) DeleteCategory(ctx context.Context, restaurantID, id uuid.UUID) error {
	// The repository returns domain.ErrCategoryNotEmpty if there are dishes.
	return uc.repo.Delete(ctx, restaurantID, id)
}
