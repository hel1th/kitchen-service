package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

type FakeCategoryRepo struct {
	Categories map[uuid.UUID]*domain.Category
	ErrDelete  error
}

func NewFakeCategoryRepo() *FakeCategoryRepo {
	return &FakeCategoryRepo{
		Categories: make(map[uuid.UUID]*domain.Category),
	}
}

func (f *FakeCategoryRepo) Create(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	f.Categories[category.ID] = category
	return category, nil
}

func (f *FakeCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	cat, ok := f.Categories[id]
	if !ok {
		return nil, domain.ErrCategoryNotFound
	}
	return cat, nil
}

func (f *FakeCategoryRepo) GetByIDAndRestaurantID(
	ctx context.Context,
	id, restaurantID uuid.UUID,
) (*domain.Category, error) {
	cat, ok := f.Categories[id]
	if !ok || cat.RestaurantID != restaurantID {
		return nil, domain.ErrCategoryNotFound
	}
	return cat, nil
}

func (f *FakeCategoryRepo) ListByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]domain.Category, error) {
	var list []domain.Category
	for _, c := range f.Categories {
		if c.RestaurantID == restaurantID {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (f *FakeCategoryRepo) Delete(ctx context.Context, restaurantID, id uuid.UUID) error {
	if f.ErrDelete != nil {
		return f.ErrDelete
	}
	delete(f.Categories, id)
	return nil
}

type FakeDishRepo struct {
	Dishes map[uuid.UUID]*domain.Dish
}

func NewFakeDishRepo() *FakeDishRepo {
	return &FakeDishRepo{
		Dishes: make(map[uuid.UUID]*domain.Dish),
	}
}

func (f *FakeDishRepo) Create(ctx context.Context, dish *domain.Dish) (*domain.Dish, error) {
	f.Dishes[dish.ID] = dish
	return dish, nil
}

func (f *FakeDishRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dish, error) {
	d, ok := f.Dishes[id]
	if !ok {
		return nil, domain.ErrDishNotFound
	}
	return d, nil
}

func (f *FakeDishRepo) GetByIDAndRestaurantID(ctx context.Context, id, restaurantID uuid.UUID) (*domain.Dish, error) {
	d, ok := f.Dishes[id]
	if !ok || d.RestaurantID != restaurantID {
		return nil, domain.ErrDishNotFound
	}
	return d, nil
}

func (f *FakeDishRepo) Update(ctx context.Context, dish *domain.Dish) (*domain.Dish, error) {
	f.Dishes[dish.ID] = dish
	return dish, nil
}

func (f *FakeDishRepo) SoftDelete(ctx context.Context, restaurantID, id uuid.UUID, at time.Time) error {
	d, ok := f.Dishes[id]
	if !ok || d.RestaurantID != restaurantID {
		return domain.ErrDishNotFound
	}
	d.SoftDelete(at)
	return nil
}

func (f *FakeDishRepo) Restore(ctx context.Context, restaurantID, id uuid.UUID, at time.Time) (*domain.Dish, error) {
	d, ok := f.Dishes[id]
	if !ok || d.RestaurantID != restaurantID {
		return nil, domain.ErrDishNotFound
	}
	d.Restore(at)
	return d, nil
}

func (f *FakeDishRepo) SetAvailability(
	ctx context.Context,
	restaurantID, id uuid.UUID,
	available bool,
	at time.Time,
) (*domain.Dish, error) {
	d, ok := f.Dishes[id]
	if !ok || d.RestaurantID != restaurantID {
		return nil, domain.ErrDishNotFound
	}
	d.SetAvailability(available, at)
	return d, nil
}

func (f *FakeDishRepo) ListByRestaurantID(
	ctx context.Context,
	restaurantID uuid.UUID,
	includeDeleted bool,
) ([]domain.Dish, error) {
	var list []domain.Dish
	for _, d := range f.Dishes {
		if d.RestaurantID == restaurantID {
			if includeDeleted || !d.IsDeleted() {
				list = append(list, *d)
			}
		}
	}
	return list, nil
}

func (f *FakeDishRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Dish, error) {
	var list []domain.Dish
	for _, id := range ids {
		if d, ok := f.Dishes[id]; ok {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (f *FakeDishRepo) GetMenuWithDishes(
	ctx context.Context,
	restaurantID uuid.UUID,
) ([]domain.CategoryWithDishes, error) {
	return nil, nil // not implemented for test
}
