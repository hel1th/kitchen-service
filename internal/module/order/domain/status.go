package domain

type OrderStatus string

const (
	StatusCreated   OrderStatus = "created"
	StatusCooking   OrderStatus = "cooking"
	StatusReady     OrderStatus = "ready"
	StatusCancelled OrderStatus = "cancelled"
)

var allowedTransitions = map[OrderStatus]map[OrderStatus]bool{
	StatusCreated:   {StatusCreated: true, StatusCooking: true, StatusCancelled: true},
	StatusCooking:   {StatusCooking: true, StatusReady: true},
	StatusReady:     {StatusReady: true},
	StatusCancelled: {StatusCancelled: true},
}

// CanTransition returns true if moving from the 'from' status to the 'to' status is allowed.
// Note that transitioning to the same status is considered valid (idempotent).
func CanTransition(from, to OrderStatus) bool {
	destinations, exists := allowedTransitions[from]
	if !exists {
		return false
	}
	return destinations[to]
}

// IsValid checks whether the OrderStatus value is a recognized status.
func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusCooking, StatusReady, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the order status is final (ready or cancelled).
func (s OrderStatus) IsTerminal() bool {
	return s == StatusReady || s == StatusCancelled
}

// String returns string representation of OrderStatus.
func (s OrderStatus) String() string {
	return string(s)
}
