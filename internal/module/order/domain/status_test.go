package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
)

func TestCanTransition_All16Combinations(t *testing.T) {
	allStatuses := []domain.OrderStatus{
		domain.StatusCreated,
		domain.StatusCooking,
		domain.StatusReady,
		domain.StatusCancelled,
	}

	expectedTransitions := map[domain.OrderStatus]map[domain.OrderStatus]bool{
		domain.StatusCreated: {
			domain.StatusCreated:   true,  // Idempotent
			domain.StatusCooking:   true,  // Normal flow
			domain.StatusReady:     false, // Cannot skip cooking
			domain.StatusCancelled: true,  // Cancellation from created
		},
		domain.StatusCooking: {
			domain.StatusCreated:   false, // Cannot rollback
			domain.StatusCooking:   true,  // Idempotent
			domain.StatusReady:     true,  // Normal flow
			domain.StatusCancelled: false, // Cannot cancel once cooking
		},
		domain.StatusReady: {
			domain.StatusCreated:   false, // Terminal state
			domain.StatusCooking:   false, // Terminal state
			domain.StatusReady:     true,  // Idempotent
			domain.StatusCancelled: false, // Terminal state
		},
		domain.StatusCancelled: {
			domain.StatusCreated:   false, // Terminal state
			domain.StatusCooking:   false, // Terminal state
			domain.StatusReady:     false, // Terminal state
			domain.StatusCancelled: true,  // Idempotent
		},
	}

	totalCombinations := 0
	for _, from := range allStatuses {
		for _, to := range allStatuses {
			totalCombinations++
			fromStatus := from
			toStatus := to
			expected := expectedTransitions[fromStatus][toStatus]

			t.Run(string(fromStatus)+" -> "+string(toStatus), func(t *testing.T) {
				actual := domain.CanTransition(fromStatus, toStatus)
				assert.Equal(
					t,
					expected,
					actual,
					"Transition from %s to %s should be %v",
					fromStatus,
					toStatus,
					expected,
				)
			})
		}
	}

	assert.Equal(t, 16, totalCombinations, "Must test exactly 16 (4x4) status combinations")
}

func TestCanTransition_UnknownStatuses(t *testing.T) {
	unknownStatus := domain.OrderStatus("unknown")

	assert.False(t, domain.CanTransition(unknownStatus, domain.StatusCreated))
	assert.False(t, domain.CanTransition(domain.StatusCreated, unknownStatus))
	assert.False(t, domain.CanTransition(unknownStatus, unknownStatus))
}

func TestOrderStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.OrderStatus
		want   bool
	}{
		{name: "created is valid", status: domain.StatusCreated, want: true},
		{name: "cooking is valid", status: domain.StatusCooking, want: true},
		{name: "ready is valid", status: domain.StatusReady, want: true},
		{name: "cancelled is valid", status: domain.StatusCancelled, want: true},
		{name: "empty string is invalid", status: domain.OrderStatus(""), want: false},
		{name: "arbitrary string is invalid", status: domain.OrderStatus("pending"), want: false},
		{name: "case sensitive check", status: domain.OrderStatus("CREATED"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestOrderStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status domain.OrderStatus
		want   bool
	}{
		{name: "created is not terminal", status: domain.StatusCreated, want: false},
		{name: "cooking is not terminal", status: domain.StatusCooking, want: false},
		{name: "ready is terminal", status: domain.StatusReady, want: true},
		{name: "cancelled is terminal", status: domain.StatusCancelled, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsTerminal())
		})
	}
}

func TestOrderStatus_String(t *testing.T) {
	assert.Equal(t, "created", domain.StatusCreated.String())
	assert.Equal(t, "cooking", domain.StatusCooking.String())
	assert.Equal(t, "ready", domain.StatusReady.String())
	assert.Equal(t, "cancelled", domain.StatusCancelled.String())
}
