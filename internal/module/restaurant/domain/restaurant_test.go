package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

func TestRestaurantStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.RestaurantStatus
		expected bool
	}{
		{
			name:     "open status is valid",
			status:   domain.StatusOpen,
			expected: true,
		},
		{
			name:     "closed status is valid",
			status:   domain.StatusClosed,
			expected: true,
		},
		{
			name:     "empty status is invalid",
			status:   domain.RestaurantStatus(""),
			expected: false,
		},
		{
			name:     "unknown status is invalid",
			status:   domain.RestaurantStatus("pending"),
			expected: false,
		},
		{
			name:     "uppercase status is invalid",
			status:   domain.RestaurantStatus("OPEN"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestRestaurant_IsOpen(t *testing.T) {
	tests := []struct {
		name       string
		restaurant *domain.Restaurant
		expected   bool
	}{
		{
			name: "open restaurant returns true",
			restaurant: &domain.Restaurant{
				ID:         uuid.New(),
				Name:       "Burger Place",
				Status:     domain.StatusOpen,
				WebhookURL: "http://example.com/webhook",
				CreatedAt:  time.Now(),
			},
			expected: true,
		},
		{
			name: "closed restaurant returns false",
			restaurant: &domain.Restaurant{
				ID:         uuid.New(),
				Name:       "Pizza Place",
				Status:     domain.StatusClosed,
				WebhookURL: "http://example.com/webhook",
				CreatedAt:  time.Now(),
			},
			expected: false,
		},
		{
			name: "restaurant with unknown status returns false",
			restaurant: &domain.Restaurant{
				ID:         uuid.New(),
				Name:       "Taco Place",
				Status:     domain.RestaurantStatus("unknown"),
				WebhookURL: "http://example.com/webhook",
				CreatedAt:  time.Now(),
			},
			expected: false,
		},
		{
			name:       "nil restaurant returns false",
			restaurant: nil,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.restaurant.IsOpen())
		})
	}
}

func TestRestaurant_Fields(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	name := "Pasta Bar"
	webhookURL := "http://restaurant-simulator:8081/webhooks/orders"

	r := domain.Restaurant{
		ID:         id,
		Name:       name,
		Status:     domain.StatusOpen,
		WebhookURL: webhookURL,
		CreatedAt:  createdAt,
	}

	assert.Equal(t, id, r.ID)
	assert.Equal(t, name, r.Name)
	assert.Equal(t, domain.StatusOpen, r.Status)
	assert.Equal(t, webhookURL, r.WebhookURL)
	assert.Equal(t, createdAt, r.CreatedAt)
	assert.True(t, r.IsOpen())
}

func TestDomainErrors(t *testing.T) {
	require.EqualError(t, domain.ErrRestaurantNotFound, "restaurant not found")
	require.EqualError(t, domain.ErrInvalidRestaurantStatus, "invalid restaurant status")
}
