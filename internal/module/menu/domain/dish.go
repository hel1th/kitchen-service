package domain

import (
	"time"

	"github.com/google/uuid"
)

// Dish represents a menu item belonging to a restaurant and category.
type Dish struct {
	ID           uuid.UUID
	RestaurantID uuid.UUID
	CategoryID   uuid.UUID
	Name         string
	Description  string
	Price        float64
	Available    bool
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsAvailable returns true if the dish is flagged available and has not been soft-deleted.
func (d Dish) IsAvailable() bool {
	return d.Available && d.DeletedAt == nil
}

// IsDeleted returns true if the dish has been soft-deleted.
func (d Dish) IsDeleted() bool {
	return d.DeletedAt != nil
}

// SoftDelete marks the dish as deleted at the given time.
func (d *Dish) SoftDelete(at time.Time) {
	d.DeletedAt = &at
	d.UpdatedAt = at
}

// Restore removes the soft-deletion mark. The available flag is preserved as-is.
func (d *Dish) Restore(at time.Time) {
	d.DeletedAt = nil
	d.UpdatedAt = at
}

// SetAvailability updates the availability status of the dish.
func (d *Dish) SetAvailability(available bool, at time.Time) {
	d.Available = available
	d.UpdatedAt = at
}
