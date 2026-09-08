package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
)

func TestNewCart(t *testing.T) {
	cartID := uuid.New()
	userID := int64(42)

	cart := domain.NewCart(cartID, userID)

	assert.Equal(t, cartID, cart.ID)
	assert.Equal(t, userID, cart.UserID)
	assert.Nil(t, cart.RestaurantID)
	assert.Empty(t, cart.Items)
	assert.True(t, cart.IsEmpty())
	assert.Equal(t, 0, cart.TotalItems())
	assert.Empty(t, cart.DishIDs())
}

func TestCart_ItemOperations(t *testing.T) {
	cart := domain.NewCart(uuid.New(), 100)
	restID := uuid.New()
	cart.RestaurantID = &restID

	dish1 := uuid.New()
	dish2 := uuid.New()
	item1ID := uuid.New()
	item2ID := uuid.New()

	cart.Items = append(cart.Items,
		domain.CartItem{ID: item1ID, DishID: dish1, Quantity: 2},
		domain.CartItem{ID: item2ID, DishID: dish2, Quantity: 3},
	)

	assert.False(t, cart.IsEmpty())
	assert.Equal(t, 5, cart.TotalItems())
	assert.ElementsMatch(t, []uuid.UUID{dish1, dish2}, cart.DishIDs())

	t.Run("FindItemByDishID found", func(t *testing.T) {
		item := cart.FindItemByDishID(dish1)
		require.NotNil(t, item)
		assert.Equal(t, item1ID, item.ID)
		assert.Equal(t, 2, item.Quantity)
	})

	t.Run("FindItemByDishID not found", func(t *testing.T) {
		item := cart.FindItemByDishID(uuid.New())
		assert.Nil(t, item)
	})

	t.Run("FindItemByID found", func(t *testing.T) {
		item := cart.FindItemByID(item2ID)
		require.NotNil(t, item)
		assert.Equal(t, dish2, item.DishID)
		assert.Equal(t, 3, item.Quantity)
	})

	t.Run("FindItemByID not found", func(t *testing.T) {
		item := cart.FindItemByID(uuid.New())
		assert.Nil(t, item)
	})
}

func TestCart_ValidateRestaurant(t *testing.T) {
	rest1 := uuid.New()
	rest2 := uuid.New()

	t.Run("empty cart with nil restaurant allows any restaurant", func(t *testing.T) {
		cart := domain.NewCart(uuid.New(), 1)
		err := cart.ValidateRestaurant(rest1)
		assert.NoError(t, err)
	})

	t.Run("empty cart with same restaurant allows same restaurant", func(t *testing.T) {
		cart := domain.NewCart(uuid.New(), 1)
		cart.RestaurantID = &rest1
		err := cart.ValidateRestaurant(rest1)
		assert.NoError(t, err)
	})

	t.Run("empty cart with different restaurant allows transition since items is empty", func(t *testing.T) {
		cart := domain.NewCart(uuid.New(), 1)
		cart.RestaurantID = &rest1
		err := cart.ValidateRestaurant(rest2)
		assert.NoError(t, err)
	})

	t.Run("non-empty cart with same restaurant succeeds", func(t *testing.T) {
		cart := domain.NewCart(uuid.New(), 1)
		cart.RestaurantID = &rest1
		cart.Items = append(cart.Items, domain.CartItem{ID: uuid.New(), DishID: uuid.New(), Quantity: 1})

		err := cart.ValidateRestaurant(rest1)
		assert.NoError(t, err)
	})

	t.Run("non-empty cart with different restaurant returns ErrCartMultiRestaurant", func(t *testing.T) {
		cart := domain.NewCart(uuid.New(), 1)
		cart.RestaurantID = &rest1
		cart.Items = append(cart.Items, domain.CartItem{ID: uuid.New(), DishID: uuid.New(), Quantity: 1})

		err := cart.ValidateRestaurant(rest2)
		assert.ErrorIs(t, err, domain.ErrCartMultiRestaurant)
	})
}

func TestCart_Clear(t *testing.T) {
	cart := domain.NewCart(uuid.New(), 1)
	restID := uuid.New()
	cart.RestaurantID = &restID
	cart.Items = append(cart.Items, domain.CartItem{ID: uuid.New(), DishID: uuid.New(), Quantity: 3})

	assert.False(t, cart.IsEmpty())
	assert.NotNil(t, cart.RestaurantID)

	cart.Clear()

	assert.True(t, cart.IsEmpty())
	assert.Nil(t, cart.RestaurantID)
	assert.Empty(t, cart.Items)
}
