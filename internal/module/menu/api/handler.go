package api

import (
	"encoding/json"
	"net/http"

	openapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/hel1th/kitchen-service/internal/api/gen"
	"github.com/hel1th/kitchen-service/internal/api/httperr"
	"github.com/hel1th/kitchen-service/internal/module/menu/usecase"
)

type MenuHandler struct {
	dishUC     *usecase.DishUsecase
	categoryUC *usecase.CategoryUsecase
}

func NewMenuHandler(dishUC *usecase.DishUsecase, categoryUC *usecase.CategoryUsecase) *MenuHandler {
	return &MenuHandler{
		dishUC:     dishUC,
		categoryUC: categoryUC,
	}
}

// GetRestaurantMenu Получить меню ресторана
func (h *MenuHandler) GetRestaurantMenu(
	w http.ResponseWriter,
	r *http.Request,
	restaurantId gen.RestaurantId,
) {
	menu, err := h.dishUC.GetMenu(r.Context(), restaurantId)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	res := make(gen.GetRestaurantMenu200JSONResponse, 0)
	for _, c := range menu {
		dishes := make([]gen.Dish, 0, len(c.Dishes))
		for _, d := range c.Dishes {
			id := d.ID
			cid := d.CategoryID
			name := d.Name
			desc := d.Description
			price := float32(d.Price)
			available := d.Available

			dish := gen.Dish{
				Id:          &id,
				CategoryId:  &cid,
				Name:        &name,
				Description: &desc,
				Price:       &price,
				Available:   &available,
			}
			if d.DeletedAt != nil {
				dish.DeletedAt = d.DeletedAt
			}
			dishes = append(dishes, dish)
		}

		cid := c.Category.ID
		rid := c.Category.RestaurantID
		cname := c.Category.Name
		sortOrder := c.Category.SortOrder

		res = append(res, gen.CategoryWithDishes{
			Id:           &cid,
			RestaurantId: &rid,
			Name:         &cname,
			SortOrder:    &sortOrder,
			Dishes:       &dishes,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// CreateCategory Создать категорию в меню
func (h *MenuHandler) CreateCategory(w http.ResponseWriter, r *http.Request, restaurantId gen.RestaurantId) {
	var req gen.CreateCategoryJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	var sortOrder int
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	cat, err := h.categoryUC.AddCategory(r.Context(), restaurantId, req.Name, sortOrder)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	id := cat.ID
	rid := cat.RestaurantID
	name := cat.Name
	so := cat.SortOrder

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(gen.Category{
		Id:           &id,
		RestaurantId: &rid,
		Name:         &name,
		SortOrder:    &so,
	})
}

// DeleteCategory Удалить категорию
func (h *MenuHandler) DeleteCategory(
	w http.ResponseWriter,
	r *http.Request,
	restaurantId gen.RestaurantId,
	categoryID openapiTypes.UUID,
) {
	err := h.categoryUC.DeleteCategory(r.Context(), restaurantId, categoryID)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateDish Добавить блюдо
func (h *MenuHandler) CreateDish(w http.ResponseWriter, r *http.Request, restaurantId gen.RestaurantId) {
	var req gen.CreateDishJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	var desc string
	if req.Description != nil {
		desc = *req.Description
	}
	// Missing Available field in DishInput. Defaulting to false or true? Usually a new dish might be available or not, but domain defaults to whatever we pass. Let's pass true. //nolint:lll
	dish, err := h.dishUC.AddDish(
		r.Context(),
		restaurantId,
		req.CategoryId,
		req.Name,
		desc,
		float64(req.Price),
		true,
	)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	id := dish.ID
	cid := dish.CategoryID
	name := dish.Name
	desc = dish.Description
	price := float32(dish.Price)
	avail := dish.Available

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(gen.Dish{
		Id:          &id,
		CategoryId:  &cid,
		Name:        &name,
		Description: &desc,
		Price:       &price,
		Available:   &avail,
	})
}

// DeleteDish Удалить блюдо (soft delete)
func (h *MenuHandler) DeleteDish(
	w http.ResponseWriter,
	r *http.Request,
	restaurantId gen.RestaurantId,
	dishId gen.DishId,
) {
	err := h.dishUC.DeleteDish(r.Context(), restaurantId, dishId)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateDish Изменить параметры блюда
func (h *MenuHandler) UpdateDish(
	w http.ResponseWriter,
	r *http.Request,
	restaurantId gen.RestaurantId,
	dishId gen.DishId,
) {
	var req gen.UpdateDishJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	params := usecase.PatchDishParams{
		CategoryID:  req.CategoryId,
		Name:        req.Name,
		Description: req.Description,
	}
	if req.Price != nil {
		p := float64(*req.Price)
		params.Price = &p
	}

	dish, err := h.dishUC.PatchDish(r.Context(), restaurantId, dishId, params)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	id := dish.ID
	cid := dish.CategoryID
	name := dish.Name
	desc := dish.Description
	price := float32(dish.Price)
	avail := dish.Available

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gen.Dish{
		Id:          &id,
		CategoryId:  &cid,
		Name:        &name,
		Description: &desc,
		Price:       &price,
		Available:   &avail,
	})
}

// SetDishAvailability Установить доступность блюда
func (h *MenuHandler) SetDishAvailability(
	w http.ResponseWriter,
	r *http.Request,
	restaurantId gen.RestaurantId,
	dishId gen.DishId,
) {
	var req gen.SetDishAvailabilityJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.HandleError(w, err)
		return
	}

	dish, err := h.dishUC.SetAvailability(r.Context(), restaurantId, dishId, req.Available)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	id := dish.ID
	cid := dish.CategoryID
	name := dish.Name
	desc := dish.Description
	price := float32(dish.Price)
	avail := dish.Available

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gen.Dish{
		Id:          &id,
		CategoryId:  &cid,
		Name:        &name,
		Description: &desc,
		Price:       &price,
		Available:   &avail,
	})
}

// RestoreDish Восстановить блюдо (undo soft delete)
func (h *MenuHandler) RestoreDish(
	w http.ResponseWriter,
	r *http.Request,
	restaurantId gen.RestaurantId,
	dishId gen.DishId,
) {
	dish, err := h.dishUC.RestoreDish(r.Context(), restaurantId, dishId)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}

	id := dish.ID
	cid := dish.CategoryID
	name := dish.Name
	desc := dish.Description
	price := float32(dish.Price)
	avail := dish.Available

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gen.Dish{
		Id:          &id,
		CategoryId:  &cid,
		Name:        &name,
		Description: &desc,
		Price:       &price,
		Available:   &avail,
	})
}
