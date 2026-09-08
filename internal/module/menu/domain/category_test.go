package domain_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

func TestCategory(t *testing.T) {
	catID := uuid.New()
	restID := uuid.New()

	cat := domain.Category{
		ID:           catID,
		RestaurantID: restID,
		Name:         "Soups",
		SortOrder:    1,
	}

	if cat.ID != catID {
		t.Fatalf("expected ID %v, got %v", catID, cat.ID)
	}
	if cat.RestaurantID != restID {
		t.Fatalf("expected RestaurantID %v, got %v", restID, cat.RestaurantID)
	}
	if cat.Name != "Soups" {
		t.Fatalf("expected Name 'Soups', got '%s'", cat.Name)
	}
	if cat.SortOrder != 1 {
		t.Fatalf("expected SortOrder 1, got %d", cat.SortOrder)
	}
}

func TestCategoryWithDishes(t *testing.T) {
	catID := uuid.New()
	restID := uuid.New()

	cat := domain.Category{
		ID:           catID,
		RestaurantID: restID,
		Name:         "Drinks",
		SortOrder:    2,
	}

	dish1 := domain.Dish{
		ID:           uuid.New(),
		RestaurantID: restID,
		CategoryID:   catID,
		Name:         "Lemonade",
		Price:        3.5,
		Available:    true,
	}

	dish2 := domain.Dish{
		ID:           uuid.New(),
		RestaurantID: restID,
		CategoryID:   catID,
		Name:         "Tea",
		Price:        2.0,
		Available:    true,
	}

	cwd := domain.CategoryWithDishes{
		Category: cat,
		Dishes:   []domain.Dish{dish1, dish2},
	}

	if cwd.Category.Name != "Drinks" {
		t.Fatalf("expected Category Name 'Drinks', got '%s'", cwd.Category.Name)
	}
	if len(cwd.Dishes) != 2 {
		t.Fatalf("expected 2 dishes, got %d", len(cwd.Dishes))
	}
	if cwd.Dishes[0].Name != "Lemonade" || cwd.Dishes[1].Name != "Tea" {
		t.Fatalf("unexpected dishes content: %+v", cwd.Dishes)
	}
}

func TestDomainErrors(t *testing.T) {
	errorsList := []error{
		domain.ErrDishNotFound,
		domain.ErrCategoryNotFound,
		domain.ErrCategoryNotEmpty,
		domain.ErrDishAlreadyDeleted,
		domain.ErrDishNotDeleted,
	}

	for _, err := range errorsList {
		if err == nil {
			t.Fatalf("expected error not to be nil")
		}
		if err.Error() == "" {
			t.Fatalf("expected error message not to be empty")
		}
	}
}
