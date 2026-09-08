package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	menuUsecase "github.com/hel1th/kitchen-service/internal/module/menu/usecase"
	"github.com/hel1th/kitchen-service/internal/module/order/domain"
	restUsecase "github.com/hel1th/kitchen-service/internal/module/restaurant/usecase"
)

// UnavailableItemsError represents a domain error where checkout cannot proceed
// because some dishes in the cart are unavailable or deleted.
type UnavailableItemsError struct {
	Items []menuUsecase.UnavailableDish
}

func (e *UnavailableItemsError) Error() string {
	return fmt.Sprintf("cart contains %d unavailable items", len(e.Items))
}

// WebhookSender defines the contract for asynchronously sending webhook events.
type WebhookSender interface {
	SendOrderCreated(ctx context.Context, order *domain.Order) error
}

type OrderUsecase struct {
	cartRepo      CartRepository
	orderRepo     OrderRepository
	menuProvider  menuUsecase.MenuDishProvider
	availChecker  menuUsecase.AvailabilityChecker
	restProvider  restUsecase.RestaurantProvider
	webhookSender WebhookSender
}

func NewOrderUsecase(
	cartRepo CartRepository,
	orderRepo OrderRepository,
	menuProvider menuUsecase.MenuDishProvider,
	availChecker menuUsecase.AvailabilityChecker,
	restProvider restUsecase.RestaurantProvider,
	webhookSender WebhookSender,
) *OrderUsecase {
	return &OrderUsecase{
		cartRepo:      cartRepo,
		orderRepo:     orderRepo,
		menuProvider:  menuProvider,
		availChecker:  availChecker,
		restProvider:  restProvider,
		webhookSender: webhookSender,
	}
}

// AddToCart adds a specific quantity of a dish to the user's cart.
func (uc *OrderUsecase) AddToCart(ctx context.Context, userID int64, dishID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return domain.ErrInvalidQuantity
	}

	dish, err := uc.menuProvider.GetDish(ctx, dishID)
	if err != nil {
		return err // ErrDishNotFound
	}
	if !dish.Available || dish.DeletedAt != nil {
		return fmt.Errorf("dish is unavailable") // We can return 409 or something based on phase 7
	}

	cart, err := uc.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if valErr := cart.ValidateRestaurant(dish.RestaurantID); valErr != nil {
		return valErr
	}

	isOpen, err := uc.restProvider.IsOpen(ctx, dish.RestaurantID)
	if err != nil {
		return err
	}
	if !isOpen {
		return domain.ErrRestaurantClosed
	}

	if cart.IsEmpty() || cart.RestaurantID == nil || *cart.RestaurantID != dish.RestaurantID {
		if err := uc.cartRepo.SetRestaurant(ctx, cart.ID, dish.RestaurantID); err != nil {
			return err
		}
	}

	return uc.cartRepo.AddItem(ctx, cart.ID, dishID, quantity)
}

// CartWithDetails represents a cart with live dish information enriched from the menu module.
type CartWithDetails struct {
	Cart  *domain.Cart
	Items []CartItemDetails
	Total float64
}

// CartItemDetails represents a cart item with live price and name.
type CartItemDetails struct {
	Item      domain.CartItem
	DishName  string
	Price     float64
	Available bool
}

// GetCart retrieves the user's cart and enriches it with live menu information.
func (uc *OrderUsecase) GetCart(ctx context.Context, userID int64) (*CartWithDetails, error) {
	cart, err := uc.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	dishIDs := cart.DishIDs()
	if len(dishIDs) == 0 {
		return &CartWithDetails{Cart: cart, Items: nil, Total: 0}, nil
	}

	dishes, err := uc.menuProvider.GetDishesByIDs(ctx, dishIDs)
	if err != nil {
		return nil, err
	}

	var items []CartItemDetails
	var total float64

	for _, item := range cart.Items {
		info, ok := dishes[item.DishID]
		if !ok {
			continue // Should not happen in a consistent state, but safe to skip
		}

		isAvailable := info.Available && info.DeletedAt == nil

		items = append(items, CartItemDetails{
			Item:      item,
			DishName:  info.Name,
			Price:     info.Price,
			Available: isAvailable,
		})

		if isAvailable {
			total += info.Price * float64(item.Quantity)
		}
	}

	return &CartWithDetails{
		Cart:  cart,
		Items: items,
		Total: total,
	}, nil
}

// Checkout creates an order from the user's cart and clears the cart.
func (uc *OrderUsecase) Checkout(ctx context.Context, userID int64) (*domain.Order, error) {
	cart, err := uc.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cart.IsEmpty() || cart.RestaurantID == nil {
		return nil, domain.ErrCartEmpty
	}

	isOpen, err := uc.restProvider.IsOpen(ctx, *cart.RestaurantID)
	if err != nil {
		return nil, err
	}
	if !isOpen {
		return nil, domain.ErrRestaurantClosed
	}

	dishIDs := cart.DishIDs()
	unavailable, err := uc.availChecker.CheckAvailability(ctx, *cart.RestaurantID, dishIDs)
	if err != nil {
		return nil, err
	}
	if len(unavailable) > 0 {
		return nil, &UnavailableItemsError{Items: unavailable}
	}

	dishes, err := uc.menuProvider.GetDishesByIDs(ctx, dishIDs)
	if err != nil {
		return nil, err
	}

	order := &domain.Order{
		ID:           uuid.New(),
		UserID:       userID,
		RestaurantID: *cart.RestaurantID,
		Status:       domain.StatusCreated,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	for _, item := range cart.Items {
		dishInfo := dishes[item.DishID]
		order.Items = append(order.Items, domain.OrderItem{
			DishID:           item.DishID,
			DishNameSnapshot: dishInfo.Name,
			PriceSnapshot:    dishInfo.Price,
			Quantity:         item.Quantity,
		})
	}

	if err := uc.orderRepo.CreateOrder(ctx, order, cart.ID); err != nil {
		return nil, err
	}

	if uc.webhookSender != nil {
		bgCtx := context.WithoutCancel(ctx)
		go func(ctx context.Context) {
			// simulate an async processing
			_ = uc.webhookSender.SendOrderCreated(ctx, order)
		}(bgCtx)
	}

	return order, nil
}

// UpdateOrderStatus transitions an order to a new status.
func (uc *OrderUsecase) UpdateOrderStatus(
	ctx context.Context,
	orderID uuid.UUID,
	newStatus domain.OrderStatus,
) (*domain.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.Status == newStatus {
		return order, nil // Idempotent
	}

	if err := order.TransitionTo(newStatus); err != nil {
		return nil, err
	}

	if err := uc.orderRepo.UpdateStatus(ctx, order.ID, newStatus); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUsecase) RemoveFromCart(ctx context.Context, userID int64, itemID uuid.UUID) error {
	cart, err := uc.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return uc.cartRepo.RemoveItem(ctx, cart.ID, itemID)
}

func (uc *OrderUsecase) UpdateCartItemQuantity(
	ctx context.Context,
	userID int64,
	itemID uuid.UUID,
	quantity int,
) error {
	if quantity <= 0 {
		return domain.ErrInvalidQuantity
	}
	cart, err := uc.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return uc.cartRepo.UpdateItemQuantity(ctx, cart.ID, itemID, quantity)
}

func (uc *OrderUsecase) ClearCart(ctx context.Context, userID int64) error {
	cart, err := uc.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return uc.cartRepo.Clear(ctx, cart.ID)
}

func (uc *OrderUsecase) ListUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	return uc.orderRepo.ListByUserID(ctx, userID)
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	return uc.orderRepo.GetByID(ctx, orderID)
}

func (uc *OrderUsecase) ListRestaurantOrders(
	ctx context.Context,
	restaurantID uuid.UUID,
	status *domain.OrderStatus,
) ([]domain.Order, error) {
	return uc.orderRepo.ListByRestaurantID(ctx, restaurantID, status)
}
