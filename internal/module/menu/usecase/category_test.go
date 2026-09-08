package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

func TestCategoryUsecase_AddCategory(t *testing.T) {
	repo := NewFakeCategoryRepo()
	uc := NewCategoryUsecase(repo)

	restaurantID := uuid.New()
	cat, err := uc.AddCategory(context.Background(), restaurantID, "Drinks", 1)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, cat.ID)
	assert.Equal(t, restaurantID, cat.RestaurantID)
	assert.Equal(t, "Drinks", cat.Name)
	assert.Equal(t, 1, cat.SortOrder)

	// verify saved
	saved, err := repo.GetByID(context.Background(), cat.ID)
	require.NoError(t, err)
	assert.Equal(t, cat, saved)
}

func TestCategoryUsecase_DeleteCategory(t *testing.T) {
	repo := NewFakeCategoryRepo()
	uc := NewCategoryUsecase(repo)

	restaurantID := uuid.New()
	cat := &domain.Category{
		ID:           uuid.New(),
		RestaurantID: restaurantID,
		Name:         "Drinks",
	}
	repo.Categories[cat.ID] = cat

	err := uc.DeleteCategory(context.Background(), restaurantID, cat.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), cat.ID)
	assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
}

func TestCategoryUsecase_DeleteCategory_NotEmpty(t *testing.T) {
	repo := NewFakeCategoryRepo()
	uc := NewCategoryUsecase(repo)

	repo.ErrDelete = domain.ErrCategoryNotEmpty

	restaurantID := uuid.New()
	id := uuid.New()

	err := uc.DeleteCategory(context.Background(), restaurantID, id)
	assert.ErrorIs(t, err, domain.ErrCategoryNotEmpty)
}
