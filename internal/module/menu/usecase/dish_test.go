package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
	"github.com/stretchr/testify/assert"
)

func TestDishUsecase_AddDish(t *testing.T) {
	dishRepo := NewFakeDishRepo()
	catRepo := NewFakeCategoryRepo()
	uc := NewDishUsecase(dishRepo, catRepo)

	restaurantID := uuid.New()
	categoryID := uuid.New()

	// Category doesn't exist
	_, err := uc.AddDish(context.Background(), restaurantID, categoryID, "Pizza", "Margarita", 10.5, true)
	assert.ErrorIs(t, err, domain.ErrCategoryNotFound)

	// Category exists
	catRepo.Categories[categoryID] = &domain.Category{ID: categoryID, RestaurantID: restaurantID}
	dish, err := uc.AddDish(context.Background(), restaurantID, categoryID, "Pizza", "Margarita", 10.5, true)
	
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, dish.ID)
	assert.Equal(t, restaurantID, dish.RestaurantID)
	assert.Equal(t, categoryID, dish.CategoryID)
	assert.Equal(t, "Pizza", dish.Name)
	assert.Equal(t, 10.5, dish.Price)
	assert.True(t, dish.Available)
	assert.False(t, dish.CreatedAt.IsZero())

	// Verify in repo
	saved, err := dishRepo.GetByID(context.Background(), dish.ID)
	assert.NoError(t, err)
	assert.Equal(t, dish, saved)
}

func TestDishUsecase_PatchDish(t *testing.T) {
	dishRepo := NewFakeDishRepo()
	catRepo := NewFakeCategoryRepo()
	uc := NewDishUsecase(dishRepo, catRepo)

	restaurantID := uuid.New()
	categoryID := uuid.New()
	catRepo.Categories[categoryID] = &domain.Category{ID: categoryID, RestaurantID: restaurantID}

	dish := &domain.Dish{
		ID:           uuid.New(),
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         "Pizza",
		Price:        10.5,
	}
	dishRepo.Dishes[dish.ID] = dish

	newName := "Pizza 2"
	newPrice := 12.0
	patched, err := uc.PatchDish(context.Background(), restaurantID, dish.ID, PatchDishParams{
		Name:  &newName,
		Price: &newPrice,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Pizza 2", patched.Name)
	assert.Equal(t, 12.0, patched.Price)
	assert.Equal(t, categoryID, patched.CategoryID)

	// Change category
	newCatID := uuid.New()
	catRepo.Categories[newCatID] = &domain.Category{ID: newCatID, RestaurantID: restaurantID}
	
	patched, err = uc.PatchDish(context.Background(), restaurantID, dish.ID, PatchDishParams{
		CategoryID: &newCatID,
	})
	assert.NoError(t, err)
	assert.Equal(t, newCatID, patched.CategoryID)
}

func TestDishUsecase_DeleteAndRestore(t *testing.T) {
	dishRepo := NewFakeDishRepo()
	uc := NewDishUsecase(dishRepo, nil)

	restaurantID := uuid.New()
	dish := &domain.Dish{
		ID:           uuid.New(),
		RestaurantID: restaurantID,
	}
	dishRepo.Dishes[dish.ID] = dish

	err := uc.DeleteDish(context.Background(), restaurantID, dish.ID)
	assert.NoError(t, err)
	assert.NotNil(t, dish.DeletedAt)

	_, err = uc.RestoreDish(context.Background(), restaurantID, dish.ID)
	assert.NoError(t, err)
	assert.Nil(t, dish.DeletedAt)
}

func TestDishUsecase_CheckAvailability(t *testing.T) {
	dishRepo := NewFakeDishRepo()
	uc := NewDishUsecase(dishRepo, nil)

	restaurantID := uuid.New()
	
	d1 := &domain.Dish{ID: uuid.New(), RestaurantID: restaurantID, Available: true}
	d2 := &domain.Dish{ID: uuid.New(), RestaurantID: restaurantID, Available: false} // unavailable
	d3 := &domain.Dish{ID: uuid.New(), RestaurantID: restaurantID, Available: true}
	d3.SoftDelete(d3.CreatedAt) // deleted
	
	dishRepo.Dishes[d1.ID] = d1
	dishRepo.Dishes[d2.ID] = d2
	dishRepo.Dishes[d3.ID] = d3
	
	d4ID := uuid.New() // not found

	unavailable, err := uc.CheckAvailability(context.Background(), restaurantID, []uuid.UUID{d1.ID, d2.ID, d3.ID, d4ID})
	assert.NoError(t, err)
	
	assert.Len(t, unavailable, 3)
	
	reasons := make(map[uuid.UUID]UnavailableDishReason)
	for _, u := range unavailable {
		reasons[u.DishID] = u.Reason
	}
	
	assert.Equal(t, ReasonUnavailable, reasons[d2.ID])
	assert.Equal(t, ReasonDeleted, reasons[d3.ID])
	assert.Equal(t, ReasonNotFound, reasons[d4ID])
}
