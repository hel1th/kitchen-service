package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
)

type FakeCartRepo struct {
	Carts map[int64]*domain.Cart
}

func NewFakeCartRepo() *FakeCartRepo {
	return &FakeCartRepo{
		Carts: make(map[int64]*domain.Cart),
	}
}

func (f *FakeCartRepo) GetByUserID(ctx context.Context, userID int64) (*domain.Cart, error) {
	cart, ok := f.Carts[userID]
	if !ok {
		cart = domain.NewCart(uuid.New(), userID)
		f.Carts[userID] = cart
	}
	return cart, nil
}

func (f *FakeCartRepo) AddItem(ctx context.Context, cartID, dishID uuid.UUID, quantity int) error {
	var cart *domain.Cart
	for _, c := range f.Carts {
		if c.ID == cartID {
			cart = c
			break
		}
	}
	if cart == nil {
		return domain.ErrCartNotFound
	}

	item := cart.FindItemByDishID(dishID)
	if item != nil {
		item.Quantity += quantity
	} else {
		cart.Items = append(cart.Items, domain.CartItem{
			ID:       uuid.New(),
			DishID:   dishID,
			Quantity: quantity,
		})
	}
	return nil
}

func (f *FakeCartRepo) SetRestaurant(ctx context.Context, cartID, restaurantID uuid.UUID) error {
	var cart *domain.Cart
	for _, c := range f.Carts {
		if c.ID == cartID {
			cart = c
			break
		}
	}
	if cart == nil {
		return domain.ErrCartNotFound
	}

	cart.RestaurantID = &restaurantID
	return nil
}

func (f *FakeCartRepo) RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	for _, c := range f.Carts {
		if c.ID == cartID {
			for i, item := range c.Items {
				if item.ID == itemID {
					c.Items = append(c.Items[:i], c.Items[i+1:]...)
					if len(c.Items) == 0 {
						c.RestaurantID = nil
					}
					return nil
				}
			}
			return domain.ErrCartItemNotFound
		}
	}
	return domain.ErrCartNotFound
}

func (f *FakeCartRepo) UpdateItemQuantity(ctx context.Context, cartID, itemID uuid.UUID, quantity int) error {
	for _, c := range f.Carts {
		if c.ID == cartID {
			for i := range c.Items {
				if c.Items[i].ID == itemID {
					c.Items[i].Quantity = quantity
					return nil
				}
			}
			return domain.ErrCartItemNotFound
		}
	}
	return domain.ErrCartNotFound
}

func (f *FakeCartRepo) Clear(ctx context.Context, cartID uuid.UUID) error {
	for _, c := range f.Carts {
		if c.ID == cartID {
			c.Clear()
			return nil
		}
	}
	return domain.ErrCartNotFound
}

// ----

type FakeOrderRepo struct {
	Orders map[uuid.UUID]*domain.Order
	Carts  *FakeCartRepo
}

func NewFakeOrderRepo(carts *FakeCartRepo) *FakeOrderRepo {
	return &FakeOrderRepo{
		Orders: make(map[uuid.UUID]*domain.Order),
		Carts:  carts,
	}
}

func (f *FakeOrderRepo) CreateOrder(ctx context.Context, order *domain.Order, cartID uuid.UUID) error {
	f.Orders[order.ID] = order

	// Simulate clearing the cart inside transaction
	for _, c := range f.Carts.Carts {
		if c.ID == cartID {
			c.Clear()
			break
		}
	}

	return nil
}

func (f *FakeOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, ok := f.Orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

func (f *FakeOrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	order, ok := f.Orders[id]
	if !ok {
		return domain.ErrOrderNotFound
	}
	order.Status = status
	return nil
}

func (f *FakeOrderRepo) ListByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	var orders []domain.Order
	for _, o := range f.Orders {
		if o.UserID == userID {
			orders = append(orders, *o)
		}
	}
	return orders, nil
}

func (f *FakeOrderRepo) ListByRestaurantID(
	ctx context.Context,
	restaurantID uuid.UUID,
	status *domain.OrderStatus,
) ([]domain.Order, error) {
	var orders []domain.Order
	for _, o := range f.Orders {
		if o.RestaurantID == restaurantID {
			if status == nil || o.Status == *status {
				orders = append(orders, *o)
			}
		}
	}
	return orders, nil
}

// ----

type FakeWebhookSender struct {
	SentOrders []*domain.Order
}

func NewFakeWebhookSender() *FakeWebhookSender {
	return &FakeWebhookSender{}
}

func (f *FakeWebhookSender) SendOrderCreated(ctx context.Context, order *domain.Order) error {
	f.SentOrders = append(f.SentOrders, order)
	return nil
}
