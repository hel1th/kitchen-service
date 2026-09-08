package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

type DishUsecase struct {
	dishRepo     DishRepository
	categoryRepo CategoryRepository
}

func NewDishUsecase(dishRepo DishRepository, categoryRepo CategoryRepository) *DishUsecase {
	return &DishUsecase{
		dishRepo:     dishRepo,
		categoryRepo: categoryRepo,
	}
}

func (uc *DishUsecase) GetMenu(ctx context.Context, restaurantID uuid.UUID) ([]domain.CategoryWithDishes, error) {
	return uc.dishRepo.GetMenuWithDishes(ctx, restaurantID)
}

func (uc *DishUsecase) AddDish(
	ctx context.Context,
	restaurantID, categoryID uuid.UUID,
	name, description string,
	price float64,
	available bool,
) (*domain.Dish, error) {
	// 1. Check if category belongs to restaurant (returns ErrCategoryNotFound if not)
	_, err := uc.categoryRepo.GetByIDAndRestaurantID(ctx, categoryID, restaurantID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	dish := &domain.Dish{
		ID:           uuid.New(),
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         name,
		Description:  description,
		Price:        price,
		Available:    available,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return uc.dishRepo.Create(ctx, dish)
}

type PatchDishParams struct {
	CategoryID  *uuid.UUID
	Name        *string
	Description *string
	Price       *float64
}

func (uc *DishUsecase) PatchDish(
	ctx context.Context,
	restaurantID, id uuid.UUID,
	params PatchDishParams,
) (*domain.Dish, error) {
	dish, err := uc.dishRepo.GetByIDAndRestaurantID(ctx, id, restaurantID)
	if err != nil {
		return nil, err
	}

	if dish.IsDeleted() {
		return nil, domain.ErrDishAlreadyDeleted
	}

	if params.CategoryID != nil && *params.CategoryID != dish.CategoryID {
		_, err := uc.categoryRepo.GetByIDAndRestaurantID(ctx, *params.CategoryID, restaurantID)
		if err != nil {
			return nil, err
		}
		dish.CategoryID = *params.CategoryID
	}

	if params.Name != nil {
		dish.Name = *params.Name
	}
	if params.Description != nil {
		dish.Description = *params.Description
	}
	if params.Price != nil {
		dish.Price = *params.Price
	}

	dish.UpdatedAt = time.Now().UTC()
	return uc.dishRepo.Update(ctx, dish)
}

func (uc *DishUsecase) DeleteDish(ctx context.Context, restaurantID, id uuid.UUID) error {
	return uc.dishRepo.SoftDelete(ctx, restaurantID, id, time.Now().UTC())
}

func (uc *DishUsecase) RestoreDish(ctx context.Context, restaurantID, id uuid.UUID) (*domain.Dish, error) {
	return uc.dishRepo.Restore(ctx, restaurantID, id, time.Now().UTC())
}

func (uc *DishUsecase) SetAvailability(
	ctx context.Context,
	restaurantID, id uuid.UUID,
	available bool,
) (*domain.Dish, error) {
	return uc.dishRepo.SetAvailability(ctx, restaurantID, id, available, time.Now().UTC())
}

// CheckAvailability implements AvailabilityChecker
func (uc *DishUsecase) CheckAvailability(
	ctx context.Context,
	restaurantID uuid.UUID,
	dishIDs []uuid.UUID,
) ([]UnavailableDish, error) {
	if len(dishIDs) == 0 {
		return nil, nil
	}

	dishes, err := uc.dishRepo.GetByIDs(ctx, dishIDs)
	if err != nil {
		return nil, err
	}

	dishMap := make(map[uuid.UUID]domain.Dish)
	for _, d := range dishes {
		dishMap[d.ID] = d
	}

	var unavailable []UnavailableDish
	for _, id := range dishIDs {
		d, found := dishMap[id]
		if !found || d.RestaurantID != restaurantID {
			unavailable = append(unavailable, UnavailableDish{DishID: id, Reason: ReasonNotFound})
			continue
		}

		if d.IsDeleted() {
			unavailable = append(unavailable, UnavailableDish{DishID: id, Reason: ReasonDeleted})
			continue
		}

		if !d.Available {
			unavailable = append(unavailable, UnavailableDish{DishID: id, Reason: ReasonUnavailable})
			continue
		}
	}

	return unavailable, nil
}

// GetDish implements MenuDishProvider
func (uc *DishUsecase) GetDish(ctx context.Context, dishID uuid.UUID) (*DishInfo, error) {
	d, err := uc.dishRepo.GetByID(ctx, dishID)
	if err != nil {
		return nil, err
	}
	info := mapDomainDishToInfo(*d)
	return &info, nil
}

// GetDishesByIDs implements MenuDishProvider
func (uc *DishUsecase) GetDishesByIDs(ctx context.Context, dishIDs []uuid.UUID) (map[uuid.UUID]DishInfo, error) {
	dishes, err := uc.dishRepo.GetByIDs(ctx, dishIDs)
	if err != nil {
		return nil, err
	}

	res := make(map[uuid.UUID]DishInfo, len(dishes))
	for _, d := range dishes {
		res[d.ID] = mapDomainDishToInfo(d)
	}
	return res, nil
}

func mapDomainDishToInfo(d domain.Dish) DishInfo {
	info := DishInfo{
		ID:           d.ID,
		RestaurantID: d.RestaurantID,
		CategoryID:   d.CategoryID,
		Name:         d.Name,
		Price:        d.Price,
		Available:    d.Available,
	}
	if d.DeletedAt != nil {
		del := true
		info.DeletedAt = &del
	}
	return info
}
