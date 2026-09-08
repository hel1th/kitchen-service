package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

type FakeRestaurantRepo struct {
	Restaurants map[uuid.UUID]*domain.Restaurant
}

func NewFakeRestaurantRepo() *FakeRestaurantRepo {
	return &FakeRestaurantRepo{
		Restaurants: make(map[uuid.UUID]*domain.Restaurant),
	}
}

func (f *FakeRestaurantRepo) Create(ctx context.Context, restaurant *domain.Restaurant) (*domain.Restaurant, error) {
	f.Restaurants[restaurant.ID] = restaurant
	return restaurant, nil
}

func (f *FakeRestaurantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	r, ok := f.Restaurants[id]
	if !ok {
		return nil, domain.ErrRestaurantNotFound
	}
	return r, nil
}

func (f *FakeRestaurantRepo) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.RestaurantStatus,
) (*domain.Restaurant, error) {
	r, ok := f.Restaurants[id]
	if !ok {
		return nil, domain.ErrRestaurantNotFound
	}
	r.Status = status
	return r, nil
}

func (f *FakeRestaurantRepo) List(ctx context.Context, filter ListFilter) ([]domain.Restaurant, int, error) {
	var list []domain.Restaurant
	for _, r := range f.Restaurants {
		if filter.Status != nil && r.Status != *filter.Status {
			continue
		}
		list = append(list, *r)
	}

	total := len(list)

	// Apply offset
	if filter.Offset > len(list) {
		list = nil
	} else {
		list = list[filter.Offset:]
	}

	// Apply limit
	if filter.Limit > 0 && filter.Limit < len(list) {
		list = list[:filter.Limit]
	}

	return list, total, nil
}
