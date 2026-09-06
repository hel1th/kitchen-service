package domain

import (
	"time"

	"github.com/google/uuid"
)

// RestaurantStatus represents the operating status of a restaurant.
type RestaurantStatus string

const (
	StatusOpen   RestaurantStatus = "open"
	StatusClosed RestaurantStatus = "closed"
)

// IsValid reports whether the restaurant status is valid.
func (s RestaurantStatus) IsValid() bool {
	return s == StatusOpen || s == StatusClosed
}

// Restaurant represents a restaurant entity in the domain model.
type Restaurant struct {
	ID         uuid.UUID
	Name       string
	Status     RestaurantStatus
	WebhookURL string
	CreatedAt  time.Time
}

// IsOpen returns true if the restaurant is currently open for orders.
func (r *Restaurant) IsOpen() bool {
	if r == nil {
		return false
	}
	return r.Status == StatusOpen
}
