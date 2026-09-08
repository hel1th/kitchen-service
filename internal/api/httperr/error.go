package httperr

import (
	"encoding/json"
	"errors"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/hel1th/kitchen-service/internal/api/gen"
	menu_domain "github.com/hel1th/kitchen-service/internal/module/menu/domain"
	order_domain "github.com/hel1th/kitchen-service/internal/module/order/domain"
	order_usecase "github.com/hel1th/kitchen-service/internal/module/order/usecase"
	restaurant_domain "github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
)

// HandleError translates domain errors into appropriate HTTP responses.
func HandleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	msg := err.Error()

	var unavailErr *order_usecase.UnavailableItemsError

	switch {
	case errors.Is(err, menu_domain.ErrDishNotFound),
		errors.Is(err, menu_domain.ErrCategoryNotFound),
		errors.Is(err, order_domain.ErrOrderNotFound),
		errors.Is(err, restaurant_domain.ErrRestaurantNotFound):
		status = http.StatusNotFound
		code = "NOT_FOUND"
	case errors.Is(err, menu_domain.ErrCategoryNotEmpty):
		status = http.StatusConflict
		code = "CATEGORY_NOT_EMPTY"
	case errors.Is(err, order_domain.ErrCartMultiRestaurant):
		status = http.StatusConflict
		code = "CART_MULTI_RESTAURANT"
	case errors.Is(err, order_domain.ErrRestaurantClosed):
		status = http.StatusConflict
		code = "RESTAURANT_CLOSED"
	case errors.Is(err, order_domain.ErrInvalidTransition):
		status = http.StatusConflict
		code = "INVALID_TRANSITION"
	case errors.Is(err, order_domain.ErrCartEmpty):
		status = http.StatusBadRequest
		code = "CART_EMPTY"
	case errors.As(err, &unavailErr):
		status = http.StatusConflict
		code = "UnavailableItemsError"

		ids := make([]openapi_types.UUID, len(unavailErr.Items))
		for i, item := range unavailErr.Items {
			ids[i] = item.DishID
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(gen.UnavailableItemsError{
			Code:               &code,
			Message:            &msg,
			UnavailableDishIds: &ids,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(gen.Error{
		Code:    &code,
		Message: &msg,
	})
}
