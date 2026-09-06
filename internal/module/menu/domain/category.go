package domain

import (
	"github.com/google/uuid"
)

// Category represents a menu category in a restaurant.
type Category struct {
	ID           uuid.UUID
	RestaurantID uuid.UUID
	Name         string
	SortOrder    int
}

// CategoryWithDishes represents a menu category with its associated dishes.
type CategoryWithDishes struct {
	Category Category
	Dishes   []Dish
}
