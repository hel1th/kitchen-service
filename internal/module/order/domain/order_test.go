package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
)

func TestOrderItem_Subtotal(t *testing.T) {
	tests := []struct {
		name     string
		item     domain.OrderItem
		expected float64
	}{
		{
			name: "single item",
			item: domain.OrderItem{
				PriceSnapshot: 250.50,
				Quantity:      1,
			},
			expected: 250.50,
		},
		{
			name: "multiple quantity",
			item: domain.OrderItem{
				PriceSnapshot: 150.00,
				Quantity:      3,
			},
			expected: 450.00,
		},
		{
			name: "zero price",
			item: domain.OrderItem{
				PriceSnapshot: 0.00,
				Quantity:      2,
			},
			expected: 0.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, tt.item.Subtotal(), 0.001)
		})
	}
}

func TestOrder_TotalAmount(t *testing.T) {
	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       1,
		RestaurantID: uuid.New(),
		Status:       domain.StatusCreated,
		Items: []domain.OrderItem{
			{
				DishID:           uuid.New(),
				DishNameSnapshot: "Pizza Margherita",
				PriceSnapshot:    450.0,
				Quantity:         2,
			},
			{
				DishID:           uuid.New(),
				DishNameSnapshot: "Cola",
				PriceSnapshot:    120.5,
				Quantity:         3,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 450 * 2 + 120.5 * 3 = 900 + 361.5 = 1261.5
	assert.InDelta(t, 1261.5, order.TotalAmount(), 0.001)
}

func TestOrder_TotalAmount_EmptyItems(t *testing.T) {
	order := &domain.Order{
		ID:     uuid.New(),
		Status: domain.StatusCreated,
		Items:  []domain.OrderItem{},
	}

	assert.InDelta(t, 0.0, order.TotalAmount(), 0.0001)
}

func TestOrder_TransitionTo(t *testing.T) {
	t.Run("valid transition created -> cooking", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCreated}
		err := order.TransitionTo(domain.StatusCooking)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCooking, order.Status)
	})

	t.Run("valid transition cooking -> ready", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCooking}
		err := order.TransitionTo(domain.StatusReady)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusReady, order.Status)
		assert.True(t, order.IsTerminal())
	})

	t.Run("valid transition created -> cancelled", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCreated}
		err := order.TransitionTo(domain.StatusCancelled)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCancelled, order.Status)
		assert.True(t, order.IsTerminal())
	})

	t.Run("idempotent transition cooking -> cooking", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCooking}
		err := order.TransitionTo(domain.StatusCooking)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusCooking, order.Status)
	})

	t.Run("invalid transition created -> ready", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCreated}
		err := order.TransitionTo(domain.StatusReady)
		require.ErrorIs(t, err, domain.ErrInvalidTransition)
		assert.Equal(t, domain.StatusCreated, order.Status)
	})

	t.Run("invalid transition cooking -> cancelled", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCooking}
		err := order.TransitionTo(domain.StatusCancelled)
		require.ErrorIs(t, err, domain.ErrInvalidTransition)
		assert.Equal(t, domain.StatusCooking, order.Status)
	})

	t.Run("invalid transition ready -> cooking", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusReady}
		err := order.TransitionTo(domain.StatusCooking)
		require.ErrorIs(t, err, domain.ErrInvalidTransition)
		assert.Equal(t, domain.StatusReady, order.Status)
	})

	t.Run("invalid transition cancelled -> created", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCancelled}
		err := order.TransitionTo(domain.StatusCreated)
		require.ErrorIs(t, err, domain.ErrInvalidTransition)
		assert.Equal(t, domain.StatusCancelled, order.Status)
	})

	t.Run("invalid status value", func(t *testing.T) {
		order := &domain.Order{Status: domain.StatusCreated}
		err := order.TransitionTo(domain.OrderStatus("invalid"))
		require.ErrorIs(t, err, domain.ErrInvalidOrderStatus)
		assert.Equal(t, domain.StatusCreated, order.Status)
	})
}

func TestOrder_IsTerminal(t *testing.T) {
	assert.False(t, (&domain.Order{Status: domain.StatusCreated}).IsTerminal())
	assert.False(t, (&domain.Order{Status: domain.StatusCooking}).IsTerminal())
	assert.True(t, (&domain.Order{Status: domain.StatusReady}).IsTerminal())
	assert.True(t, (&domain.Order{Status: domain.StatusCancelled}).IsTerminal())
}
