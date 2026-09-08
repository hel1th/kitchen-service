package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

func TestRestaurantUsecase_SetStatus(t *testing.T) {
	repo := NewFakeRestaurantRepo()
	uc := NewRestaurantUsecase(repo)

	id := uuid.New()
	repo.Restaurants[id] = &domain.Restaurant{
		ID:     id,
		Status: domain.StatusClosed,
	}

	updated, err := uc.SetStatus(context.Background(), id, domain.StatusOpen)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, updated.Status)

	// test invalid status
	_, err = uc.SetStatus(context.Background(), id, "invalid_status")
	assert.ErrorIs(t, err, domain.ErrInvalidRestaurantStatus)
}

func TestRestaurantUsecase_List(t *testing.T) {
	repo := NewFakeRestaurantRepo()
	uc := NewRestaurantUsecase(repo)

	for i := 0; i < 5; i++ {
		status := domain.StatusOpen
		if i%2 == 0 {
			status = domain.StatusClosed
		}
		id := uuid.New()
		repo.Restaurants[id] = &domain.Restaurant{
			ID:     id,
			Status: status,
		}
	}

	openStatus := domain.StatusOpen
	list, total, err := uc.List(context.Background(), ListFilter{
		Status: &openStatus,
		Limit:  10,
	})

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, list, 2)
}

func TestRestaurantUsecase_GetAndProvider(t *testing.T) {
	repo := NewFakeRestaurantRepo()
	uc := NewRestaurantUsecase(repo)

	id := uuid.New()
	repo.Restaurants[id] = &domain.Restaurant{
		ID:     id,
		Status: domain.StatusOpen,
	}

	r, err := uc.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, r.ID)

	r2, err := uc.GetRestaurant(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, r2.ID)

	isOpen, err := uc.IsOpen(context.Background(), id)
	require.NoError(t, err)
	assert.True(t, isOpen)

	notFoundID := uuid.New()
	_, err = uc.IsOpen(context.Background(), notFoundID)
	assert.ErrorIs(t, err, domain.ErrRestaurantNotFound)
}
