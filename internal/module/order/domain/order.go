package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderItem represents an immutable snapshot of an item in a placed order.
type OrderItem struct {
	DishID           uuid.UUID
	DishNameSnapshot string
	PriceSnapshot    float64
	Quantity         int
}

// Subtotal returns the calculated total price for this order item.
func (oi OrderItem) Subtotal() float64 {
	return oi.PriceSnapshot * float64(oi.Quantity)
}

// Order represents an order placed by a customer.
type Order struct {
	ID           uuid.UUID
	UserID       int64
	RestaurantID uuid.UUID
	Status       OrderStatus
	Items        []OrderItem
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TotalAmount calculates the sum of all order item subtotals.
func (o *Order) TotalAmount() float64 {
	var total float64
	for _, item := range o.Items {
		total += item.Subtotal()
	}
	return total
}

// TransitionTo updates the order's status if the transition is allowed.
// If the transition is not allowed, it returns ErrInvalidTransition.
func (o *Order) TransitionTo(newStatus OrderStatus) error {
	if !newStatus.IsValid() {
		return ErrInvalidOrderStatus
	}
	if !CanTransition(o.Status, newStatus) {
		return ErrInvalidTransition
	}
	o.Status = newStatus
	return nil
}

// IsTerminal checks if the order has reached a terminal state (ready or cancelled).
func (o *Order) IsTerminal() bool {
	return o.Status.IsTerminal()
}
