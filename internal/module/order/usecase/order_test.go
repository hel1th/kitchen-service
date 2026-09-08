package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	menuDomain "github.com/hel1th/kitchen-service/internal/module/menu/domain"
	menuUsecase "github.com/hel1th/kitchen-service/internal/module/menu/usecase"
	"github.com/hel1th/kitchen-service/internal/module/order/domain"
	restDomain "github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

// Mocks

type mockMenuProvider struct {
	dishes map[uuid.UUID]menuUsecase.DishInfo
}

func (m *mockMenuProvider) GetDish(ctx context.Context, dishID uuid.UUID) (*menuUsecase.DishInfo, error) {
	if d, ok := m.dishes[dishID]; ok {
		return &d, nil
	}
	return nil, menuDomain.ErrDishNotFound
}

func (m *mockMenuProvider) GetDishesByIDs(
	ctx context.Context,
	dishIDs []uuid.UUID,
) (map[uuid.UUID]menuUsecase.DishInfo, error) {
	res := make(map[uuid.UUID]menuUsecase.DishInfo)
	for _, id := range dishIDs {
		if d, ok := m.dishes[id]; ok {
			res[id] = d
		}
	}
	return res, nil
}

type mockAvailChecker struct {
	unavailable []menuUsecase.UnavailableDish
}

func (m *mockAvailChecker) CheckAvailability(
	ctx context.Context,
	restaurantID uuid.UUID,
	dishIDs []uuid.UUID,
) ([]menuUsecase.UnavailableDish, error) {
	_, _ = dishIDs, restaurantID
	return m.unavailable, nil
}

type mockRestProvider struct {
	restaurants map[uuid.UUID]*restDomain.Restaurant
}

func (m *mockRestProvider) GetRestaurant(ctx context.Context, id uuid.UUID) (*restDomain.Restaurant, error) {
	_ = ctx
	if r, ok := m.restaurants[id]; ok {
		return r, nil
	}
	return nil, restDomain.ErrRestaurantNotFound
}

func (m *mockRestProvider) IsOpen(ctx context.Context, id uuid.UUID) (bool, error) {
	_ = ctx
	if r, ok := m.restaurants[id]; ok {
		return r.IsOpen(), nil
	}
	return false, restDomain.ErrRestaurantNotFound
}

func setupUsecase() (*OrderUsecase, *FakeCartRepo, *FakeOrderRepo, *mockMenuProvider, *mockAvailChecker, *mockRestProvider) {
	carts := NewFakeCartRepo()
	orders := NewFakeOrderRepo(carts)
	menu := &mockMenuProvider{dishes: make(map[uuid.UUID]menuUsecase.DishInfo)}
	avail := &mockAvailChecker{}
	rest := &mockRestProvider{restaurants: make(map[uuid.UUID]*restDomain.Restaurant)}
	wh := NewFakeWebhookSender()

	// userRepo is not used by order usecase internally for cart ops, we can pass nil or mock
	uc := NewOrderUsecase(carts, orders, menu, avail, rest, wh)
	return uc, carts, orders, menu, avail, rest
}

func TestAddToCart(t *testing.T) {
	uc, _, _, menu, _, rest := setupUsecase()
	userID := int64(1)
	restID := uuid.New()
	dishID := uuid.New()

	rest.restaurants[restID] = &restDomain.Restaurant{
		ID:     restID,
		Status: restDomain.StatusOpen,
	}

	menu.dishes[dishID] = menuUsecase.DishInfo{
		ID:           dishID,
		RestaurantID: restID,
		Available:    true,
	}

	t.Run("success_add", func(t *testing.T) {
		err := uc.AddToCart(context.Background(), userID, dishID, 2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("restaurant_closed", func(t *testing.T) {
		rest.restaurants[restID].Status = restDomain.StatusClosed
		err := uc.AddToCart(context.Background(), userID, dishID, 1)
		if !errors.Is(err, domain.ErrRestaurantClosed) {
			t.Fatalf("expected ErrRestaurantClosed, got %v", err)
		}
		rest.restaurants[restID].Status = restDomain.StatusOpen
	})

	t.Run("multi_restaurant", func(t *testing.T) {
		otherRestID := uuid.New()
		otherDishID := uuid.New()
		rest.restaurants[otherRestID] = &restDomain.Restaurant{
			ID:     otherRestID,
			Status: restDomain.StatusOpen,
		}
		menu.dishes[otherDishID] = menuUsecase.DishInfo{
			ID:           otherDishID,
			RestaurantID: otherRestID,
			Available:    true,
		}

		err := uc.AddToCart(context.Background(), userID, otherDishID, 1)
		if !errors.Is(err, domain.ErrCartMultiRestaurant) {
			t.Fatalf("expected ErrCartMultiRestaurant, got %v", err)
		}
	})

	t.Run("invalid_quantity", func(t *testing.T) {
		err := uc.AddToCart(context.Background(), userID, dishID, 0)
		if !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("dish_not_found", func(t *testing.T) {
		err := uc.AddToCart(context.Background(), userID, uuid.New(), 1)
		if !errors.Is(err, menuDomain.ErrDishNotFound) {
			t.Fatalf("expected ErrDishNotFound, got %v", err)
		}
	})

	t.Run("dish_unavailable", func(t *testing.T) {
		unavailDishID := uuid.New()
		menu.dishes[unavailDishID] = menuUsecase.DishInfo{
			ID:           unavailDishID,
			RestaurantID: restID,
			Available:    false,
		}
		err := uc.AddToCart(context.Background(), userID, unavailDishID, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCheckout(t *testing.T) {
	uc, carts, _, menu, avail, rest := setupUsecase()
	userID := int64(2)
	restID := uuid.New()
	dishID := uuid.New()

	rest.restaurants[restID] = &restDomain.Restaurant{
		ID:     restID,
		Status: restDomain.StatusOpen,
	}

	menu.dishes[dishID] = menuUsecase.DishInfo{
		ID:           dishID,
		RestaurantID: restID,
		Available:    true,
		Price:        10.5,
		Name:         "Pizza",
	}

	t.Run("empty_cart", func(t *testing.T) {
		_, err := uc.Checkout(context.Background(), userID)
		if !errors.Is(err, domain.ErrCartEmpty) {
			t.Fatalf("expected ErrCartEmpty, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		err := uc.AddToCart(context.Background(), userID, dishID, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		order, err := uc.Checkout(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order == nil {
			t.Fatal("expected order to be created")
		}
		if len(order.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(order.Items))
		}
		if order.TotalAmount() != 21.0 {
			t.Fatalf("expected total 21.0, got %f", order.TotalAmount())
		}

		// check cart is cleared
		cart, _ := carts.GetByUserID(context.Background(), userID)
		if !cart.IsEmpty() {
			t.Fatal("expected cart to be empty after checkout")
		}
	})

	t.Run("unavailable_dish", func(t *testing.T) {
		_ = uc.AddToCart(context.Background(), userID, dishID, 1)

		avail.unavailable = []menuUsecase.UnavailableDish{
			{DishID: dishID, Reason: menuUsecase.ReasonUnavailable},
		}

		_, err := uc.Checkout(context.Background(), userID)
		if err == nil {
			t.Fatal("expected error due to unavailable dish")
		}
		var unavailErr *UnavailableItemsError
		if !errors.As(err, &unavailErr) {
			t.Fatalf("expected *UnavailableItemsError, got %v", err)
		}
	})

	t.Run("restaurant_closed", func(t *testing.T) {
		userIDClosed := int64(3)
		restIDClosed := uuid.New()
		dishIDClosed := uuid.New()
		rest.restaurants[restIDClosed] = &restDomain.Restaurant{
			ID:     restIDClosed,
			Status: restDomain.StatusClosed,
		}
		menu.dishes[dishIDClosed] = menuUsecase.DishInfo{
			ID:           dishIDClosed,
			RestaurantID: restIDClosed,
			Available:    true,
		}

		carts.Carts[userIDClosed] = domain.NewCart(uuid.New(), userIDClosed)
		carts.Carts[userIDClosed].RestaurantID = &restIDClosed
		carts.Carts[userIDClosed].Items = []domain.CartItem{{DishID: dishIDClosed, Quantity: 1}}

		_, err := uc.Checkout(context.Background(), userIDClosed)
		if !errors.Is(err, domain.ErrRestaurantClosed) {
			t.Fatalf("expected ErrRestaurantClosed, got %v", err)
		}
	})
}

func TestGetCart(t *testing.T) {
	uc, carts, _, menu, _, _ := setupUsecase()
	userID := int64(10)
	restID := uuid.New()
	dishID := uuid.New()

	menu.dishes[dishID] = menuUsecase.DishInfo{
		ID:           dishID,
		RestaurantID: restID,
		Available:    true,
		Price:        15.0,
		Name:         "Burger",
	}

	t.Run("empty_cart", func(t *testing.T) {
		res, err := uc.GetCart(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Total != 0 {
			t.Fatalf("expected 0 total, got %f", res.Total)
		}
	})

	t.Run("with_items", func(t *testing.T) {
		cart := domain.NewCart(uuid.New(), userID)
		cart.RestaurantID = &restID
		cart.Items = []domain.CartItem{{DishID: dishID, Quantity: 2}}
		carts.Carts[userID] = cart

		res, err := uc.GetCart(context.Background(), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(res.Items))
		}
		if res.Total != 30.0 {
			t.Fatalf("expected 30.0 total, got %f", res.Total)
		}
	})
}

func TestUpdateOrderStatus(t *testing.T) {
	uc, _, orders, _, _, _ := setupUsecase()
	orderID := uuid.New()
	orders.Orders[orderID] = &domain.Order{
		ID:     orderID,
		Status: domain.StatusCreated,
	}

	t.Run("success_transition", func(t *testing.T) {
		order, err := uc.UpdateOrderStatus(context.Background(), orderID, domain.StatusCooking)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Status != domain.StatusCooking {
			t.Fatalf("expected status cooking, got %s", order.Status)
		}
	})

	t.Run("idempotent_transition", func(t *testing.T) {
		order, err := uc.UpdateOrderStatus(context.Background(), orderID, domain.StatusCooking)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Status != domain.StatusCooking {
			t.Fatalf("expected status cooking, got %s", order.Status)
		}
	})

	t.Run("invalid_transition", func(t *testing.T) {
		_, err := uc.UpdateOrderStatus(context.Background(), orderID, domain.StatusCreated)
		if !errors.Is(err, domain.ErrInvalidTransition) {
			t.Fatalf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("order_not_found", func(t *testing.T) {
		_, err := uc.UpdateOrderStatus(context.Background(), uuid.New(), domain.StatusCooking)
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})
}
