package api

import (
	"encoding/json"
	"net/http"

	openapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/hel1th/kitchen-service/internal/api/gen"
	"github.com/hel1th/kitchen-service/internal/api/httperr"
	"github.com/hel1th/kitchen-service/internal/module/order/domain"
	"github.com/hel1th/kitchen-service/internal/module/order/usecase"
)

type OrderHandler struct {
	orderUC *usecase.OrderUsecase
}

func NewOrderHandler(orderUC *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{
		orderUC: orderUC,
	}
}

// ClearCart Очистить корзину
func (h *OrderHandler) ClearCart(w http.ResponseWriter, r *http.Request, params gen.ClearCartParams) {
	if err := h.orderUC.ClearCart(r.Context(), params.XUserId); err != nil {
		httperr.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetCart Получить текущую корзину пользователя
func (h *OrderHandler) GetCart(w http.ResponseWriter, r *http.Request, params gen.GetCartParams) {
	cartWithDetails, err := h.orderUC.GetCart(r.Context(), params.XUserId)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	res := mapCart(cartWithDetails)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// AddCartItem Добавить блюдо в корзину
func (h *OrderHandler) AddCartItem(w http.ResponseWriter, r *http.Request, params gen.AddCartItemParams) {
	var req gen.AddCartItemJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	if err := h.orderUC.AddToCart(r.Context(), params.XUserId, req.DishId, req.Quantity); err != nil {
		httperr.HandleError(w, err)
		return
	}

	h.GetCart(w, r, gen.GetCartParams(params))
}

// RemoveCartItem Удалить блюдо из корзины
func (h *OrderHandler) RemoveCartItem(
	w http.ResponseWriter,
	r *http.Request,
	itemId openapiTypes.UUID,
	params gen.RemoveCartItemParams,
) {
	if err := h.orderUC.RemoveFromCart(r.Context(), params.XUserId, itemId); err != nil {
		httperr.HandleError(w, err)
		return
	}

	h.GetCart(w, r, gen.GetCartParams(params))
}

// UpdateCartItem Изменить количество блюда в корзине
func (h *OrderHandler) UpdateCartItem(
	w http.ResponseWriter,
	r *http.Request,
	itemId openapiTypes.UUID,
	params gen.UpdateCartItemParams,
) {
	var req gen.UpdateCartItemJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	if err := h.orderUC.UpdateCartItemQuantity(r.Context(), params.XUserId, itemId, req.Quantity); err != nil {
		httperr.HandleError(w, err)
		return
	}

	h.GetCart(w, r, gen.GetCartParams(params))
}

// ListOrders Список заказов пользователя
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request, params gen.ListOrdersParams) {
	orders, err := h.orderUC.ListUserOrders(r.Context(), params.XUserId)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	items := make([]gen.Order, 0, len(orders))
	for _, o := range orders {
		items = append(items, mapOrder(&o))
	}

	total := len(orders)
	limit := 0
	offset := 0

	res := gen.ListOrders200JSONResponse{
		Items: &items,
		Pagination: &gen.PaginationMeta{
			Limit:  &limit,
			Offset: &offset,
			Total:  &total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// Checkout Оформить заказ (checkout текущей корзины)
func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request, params gen.CheckoutParams) {
	order, err := h.orderUC.Checkout(r.Context(), params.XUserId)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	res := mapOrder(order)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}

// GetOrder Получить заказ по id (мониторинг статуса)
func (h *OrderHandler) GetOrder(
	w http.ResponseWriter,
	r *http.Request,
	orderID gen.OrderId,
	params gen.GetOrderParams,
) {
	order, err := h.orderUC.GetOrder(r.Context(), orderID)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	res := mapOrder(order)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// UpdateOrderStatus Обновить статус заказа (перевод по стейт-машине)
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request, orderID gen.OrderId) {
	var req gen.UpdateOrderStatusJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	status := domain.OrderStatus(req.Status)
	order, err := h.orderUC.UpdateOrderStatus(r.Context(), orderID, status)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	res := mapOrder(order)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// ListRestaurantOrders Получить список заказов ресторана (с фильтрацией по статусу)
func (h *OrderHandler) ListRestaurantOrders(
	w http.ResponseWriter,
	r *http.Request,
	restaurantId gen.RestaurantId,
	params gen.ListRestaurantOrdersParams,
) {
	var status *domain.OrderStatus
	if params.Status != nil {
		s := domain.OrderStatus(*params.Status)
		status = &s
	}

	orders, err := h.orderUC.ListRestaurantOrders(r.Context(), restaurantId, status)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	items := make([]gen.Order, 0, len(orders))
	for _, o := range orders {
		items = append(items, mapOrder(&o))
	}

	total := len(orders)
	limit := 0
	offset := 0

	res := gen.ListRestaurantOrders200JSONResponse{
		Items: &items,
		Pagination: &gen.PaginationMeta{
			Limit:  &limit,
			Offset: &offset,
			Total:  &total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func mapCart(cd *usecase.CartWithDetails) gen.Cart {
	items := make([]gen.CartItem, 0, len(cd.Items))
	for _, item := range cd.Items {
		id := item.Item.ID
		dishId := item.Item.DishID
		quantity := item.Item.Quantity
		name := item.DishName
		price := float32(item.Price)
		available := item.Available

		items = append(items, gen.CartItem{
			Id:        &id,
			DishId:    &dishId,
			Quantity:  &quantity,
			DishName:  &name,
			Price:     &price,
			Available: &available,
		})
	}

	id := cd.Cart.ID
	var restID *openapiTypes.UUID
	if cd.Cart.RestaurantID != nil {
		rID := *cd.Cart.RestaurantID
		restID = &rID
	}

	return gen.Cart{
		Id:           &id,
		RestaurantId: restID,
		Items:        &items,
	}
}

func mapOrder(o *domain.Order) gen.Order {
	items := make([]gen.OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		dishId := item.DishID
		name := item.DishNameSnapshot
		price := float32(item.PriceSnapshot)
		quantity := item.Quantity

		items = append(items, gen.OrderItem{
			DishId:        &dishId,
			DishName:      &name,
			PriceSnapshot: &price,
			Quantity:      &quantity,
		})
	}

	id := o.ID
	userID := o.UserID
	restID := o.RestaurantID
	status := gen.OrderStatus(o.Status)
	createdAt := o.CreatedAt
	updatedAt := o.UpdatedAt

	return gen.Order{
		Id:           &id,
		UserId:       &userID,
		RestaurantId: &restID,
		Status:       &status,
		Items:        &items,
		CreatedAt:    &createdAt,
		UpdatedAt:    &updatedAt,
	}
}
