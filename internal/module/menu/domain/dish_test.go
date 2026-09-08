package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

func TestDish_IsAvailable(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		available bool
		deletedAt *time.Time
		expected  bool
	}{
		{
			name:      "available and not deleted -> available",
			available: true,
			deletedAt: nil,
			expected:  true,
		},
		{
			name:      "available but soft-deleted -> unavailable",
			available: true,
			deletedAt: &now,
			expected:  false,
		},
		{
			name:      "unavailable and not deleted -> unavailable",
			available: false,
			deletedAt: nil,
			expected:  false,
		},
		{
			name:      "unavailable and soft-deleted -> unavailable",
			available: false,
			deletedAt: &now,
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := domain.Dish{
				ID:           uuid.New(),
				RestaurantID: uuid.New(),
				CategoryID:   uuid.New(),
				Name:         "Margherita Pizza",
				Description:  "Classic cheese and tomato",
				Price:        12.50,
				Available:    tc.available,
				DeletedAt:    tc.deletedAt,
			}

			if got := d.IsAvailable(); got != tc.expected {
				t.Fatalf("expected IsAvailable() = %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestDish_IsDeleted(t *testing.T) {
	now := time.Now()

	t.Run("not deleted when deletedAt is nil", func(t *testing.T) {
		d := domain.Dish{
			ID:        uuid.New(),
			DeletedAt: nil,
		}
		if d.IsDeleted() {
			t.Fatalf("expected IsDeleted() = false, got true")
		}
	})

	t.Run("deleted when deletedAt is set", func(t *testing.T) {
		d := domain.Dish{
			ID:        uuid.New(),
			DeletedAt: &now,
		}
		if !d.IsDeleted() {
			t.Fatalf("expected IsDeleted() = true, got false")
		}
	})
}

func TestDish_SoftDelete(t *testing.T) {
	d := domain.Dish{
		ID:        uuid.New(),
		Available: true,
		DeletedAt: nil,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	deleteTime := time.Now()
	d.SoftDelete(deleteTime)

	if !d.IsDeleted() {
		t.Fatalf("expected dish to be deleted")
	}
	if d.DeletedAt == nil || !d.DeletedAt.Equal(deleteTime) {
		t.Fatalf("expected DeletedAt to equal %v, got %v", deleteTime, d.DeletedAt)
	}
	if !d.UpdatedAt.Equal(deleteTime) {
		t.Fatalf("expected UpdatedAt to equal %v, got %v", deleteTime, d.UpdatedAt)
	}
	if d.IsAvailable() {
		t.Fatalf("expected soft-deleted dish to not be available")
	}
}

func TestDish_Restore(t *testing.T) {
	deleteTime := time.Now().Add(-30 * time.Minute)
	d := domain.Dish{
		ID:        uuid.New(),
		Available: true,
		DeletedAt: &deleteTime,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: deleteTime,
	}

	restoreTime := time.Now()
	d.Restore(restoreTime)

	if d.IsDeleted() {
		t.Fatalf("expected dish not to be deleted after restore")
	}
	if d.DeletedAt != nil {
		t.Fatalf("expected DeletedAt to be nil, got %v", d.DeletedAt)
	}
	if !d.UpdatedAt.Equal(restoreTime) {
		t.Fatalf("expected UpdatedAt to equal %v, got %v", restoreTime, d.UpdatedAt)
	}
	if !d.IsAvailable() {
		t.Fatalf("expected restored dish with available=true to be available")
	}

	// Restore of dish with available=false keeps available=false
	d.Available = false
	d.DeletedAt = &deleteTime
	d.Restore(restoreTime)
	if d.IsAvailable() {
		t.Fatalf("expected restored dish with available=false to remain unavailable")
	}
}

func TestDish_SetAvailability(t *testing.T) {
	d := domain.Dish{
		ID:        uuid.New(),
		Available: false,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	updateTime := time.Now()
	d.SetAvailability(true, updateTime)

	if !d.Available {
		t.Fatalf("expected Available to be true")
	}
	if !d.UpdatedAt.Equal(updateTime) {
		t.Fatalf("expected UpdatedAt to equal %v, got %v", updateTime, d.UpdatedAt)
	}
	if !d.IsAvailable() {
		t.Fatalf("expected dish to be available")
	}
}
