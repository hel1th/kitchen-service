package main

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type SimulatedOrder struct {
	OrderID    uuid.UUID `json:"orderId"`
	Status     string    `json:"status"` // "created" | "cooking" | "ready"
	ReceivedAt time.Time `json:"receivedAt"`
}

type OrderPayload struct {
	OrderID      uuid.UUID `json:"orderId"`
	RestaurantID uuid.UUID `json:"restaurantId"`
	Items        []struct {
		DishID        uuid.UUID `json:"dishId"`
		DishName      string    `json:"dishName"`
		PriceSnapshot float64   `json:"priceSnapshot"`
		Quantity      int       `json:"quantity"`
	} `json:"items"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	mu     sync.Mutex
	orders map[uuid.UUID]*SimulatedOrder
}

func NewStore() *Store {
	return &Store{
		orders: make(map[uuid.UUID]*SimulatedOrder),
	}
}

func (s *Store) TryInsert(p OrderPayload) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.orders[p.OrderID]; exists {
		return false
	}
	s.orders[p.OrderID] = &SimulatedOrder{
		OrderID:    p.OrderID,
		Status:     "created",
		ReceivedAt: time.Now(),
	}
	return true
}

func (s *Store) UpdateStatus(id uuid.UUID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o, ok := s.orders[id]; ok {
		o.Status = status
	}
}

func (s *Store) List() []*SimulatedOrder {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]*SimulatedOrder, 0, len(s.orders))
	for _, o := range s.orders {
		list = append(list, o)
	}
	return list
}
