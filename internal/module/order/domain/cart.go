package domain

import (
	"github.com/google/uuid"
)

// CartItem represents an item in a user's cart.
// Price and dish availability are not stored in the cart domain model,
// as they are resolved dynamically from the menu module.
type CartItem struct {
	ID       uuid.UUID
	DishID   uuid.UUID
	Quantity int
}

// Cart represents a user's active shopping cart.
type Cart struct {
	ID           uuid.UUID
	UserID       int64
	RestaurantID *uuid.UUID
	Items        []CartItem
}

// NewCart creates a new empty Cart instance for a given user.
func NewCart(id uuid.UUID, userID int64) *Cart {
	return &Cart{
		ID:           id,
		UserID:       userID,
		RestaurantID: nil,
		Items:        make([]CartItem, 0),
	}
}

// IsEmpty returns true if the cart has no items.
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// TotalItems returns the sum of quantities of all items in the cart.
func (c *Cart) TotalItems() int {
	total := 0
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}

// DishIDs returns a slice of all unique dish IDs currently in the cart.
func (c *Cart) DishIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(c.Items))
	for _, item := range c.Items {
		ids = append(ids, item.DishID)
	}
	return ids
}

// FindItemByDishID finds an item in the cart by dish ID.
func (c *Cart) FindItemByDishID(dishID uuid.UUID) *CartItem {
	for i := range c.Items {
		if c.Items[i].DishID == dishID {
			return &c.Items[i]
		}
	}
	return nil
}

// FindItemByID finds an item in the cart by its cart item ID.
func (c *Cart) FindItemByID(itemID uuid.UUID) *CartItem {
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			return &c.Items[i]
		}
	}
	return nil
}

// ValidateRestaurant checks if adding a dish from the given restaurant is allowed in this cart.
// If the cart already has items from another restaurant, ErrCartMultiRestaurant is returned.
func (c *Cart) ValidateRestaurant(restaurantID uuid.UUID) error {
	if c.RestaurantID != nil && *c.RestaurantID != restaurantID && !c.IsEmpty() {
		return ErrCartMultiRestaurant
	}
	return nil
}

// Clear removes all items from the cart and resets the attached restaurant.
func (c *Cart) Clear() {
	c.Items = make([]CartItem, 0)
	c.RestaurantID = nil
}
