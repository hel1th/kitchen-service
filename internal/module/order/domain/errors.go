package domain

import "errors"

var (
	ErrCartMultiRestaurant = errors.New("cart already has dishes from another restaurant")
	ErrRestaurantClosed    = errors.New("restaurant is closed")
	ErrInvalidTransition   = errors.New("invalid order status transition")
	ErrCartEmpty           = errors.New("cart is empty")
	ErrOrderNotFound       = errors.New("order not found")
	ErrCartNotFound        = errors.New("cart not found")
	ErrCartItemNotFound    = errors.New("cart item not found")
	ErrInvalidQuantity     = errors.New("quantity must be greater than zero")
	ErrInvalidOrderStatus  = errors.New("invalid order status")
)
