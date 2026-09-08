package api

import (
	"encoding/json"
	"net/http"

	"github.com/hel1th/kitchen-service/internal/api/gen"
	"github.com/hel1th/kitchen-service/internal/api/httperr"
	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
	"github.com/hel1th/kitchen-service/internal/module/restaurant/usecase"
)

type RestaurantHandler struct {
	restUC *usecase.RestaurantUsecase
}

func NewRestaurantHandler(restUC *usecase.RestaurantUsecase) *RestaurantHandler {
	return &RestaurantHandler{
		restUC: restUC,
	}
}

// GetHealthz Healthcheck
func (h *RestaurantHandler) GetHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

// ListRestaurants Список всех ресторанов
func (h *RestaurantHandler) ListRestaurants(w http.ResponseWriter, r *http.Request, params gen.ListRestaurantsParams) {
	filter := usecase.ListFilter{
		Limit:  20,
		Offset: 0,
	}
	if params.Limit != nil {
		filter.Limit = *params.Limit
	}
	if params.Offset != nil {
		filter.Offset = *params.Offset
	}
	if params.Status != nil {
		status := domain.RestaurantStatus(*params.Status)
		filter.Status = &status
	}

	rests, total, err := h.restUC.List(r.Context(), filter)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	items := make([]gen.Restaurant, 0, len(rests))
	for _, rest := range rests {
		id := rest.ID
		name := rest.Name
		status := gen.RestaurantStatus(rest.Status)
		items = append(items, gen.Restaurant{
			Id:     &id,
			Name:   &name,
			Status: &status,
		})
	}

	res := gen.ListRestaurants200JSONResponse{
		Items: &items,
		Pagination: &gen.PaginationMeta{
			Limit:  &filter.Limit,
			Offset: &filter.Offset,
			Total:  &total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GetRestaurant Получить инфу о ресторане
func (h *RestaurantHandler) GetRestaurant(w http.ResponseWriter, r *http.Request, restaurantID gen.RestaurantId) {
	rest, err := h.restUC.Get(r.Context(), restaurantID)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	id := rest.ID
	name := rest.Name
	status := gen.RestaurantStatus(rest.Status)

	res := gen.GetRestaurant200JSONResponse{
		Id:     &id,
		Name:   &name,
		Status: &status,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// SetRestaurantStatus Изменить статус ресторана (открыт/закрыт)
func (h *RestaurantHandler) SetRestaurantStatus(w http.ResponseWriter, r *http.Request, restaurantID gen.RestaurantId) {
	var req gen.SetRestaurantStatusJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	status := domain.RestaurantStatus(req.Status)
	rest, err := h.restUC.SetStatus(r.Context(), restaurantID, status)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	id := rest.ID
	name := rest.Name
	newStatus := gen.RestaurantStatus(rest.Status)

	res := gen.Restaurant{
		Id:     &id,
		Name:   &name,
		Status: &newStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
