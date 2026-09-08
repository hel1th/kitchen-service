package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

func TestDishRepo_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dishID := uuid.New()
	restaurantID := uuid.New()
	categoryID := uuid.New()
	now := time.Now().UTC()

	dish := &domain.Dish{
		ID:           dishID,
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         "Burger",
		Description:  "Juicy beef burger",
		Price:        499.00,
		Available:    true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO dishes").
			WithArgs(dishID, restaurantID, categoryID, "Burger", "Juicy beef burger", 499.00, true, (*time.Time)(nil), now, now).
			WillReturnRows(
				pgxmock.NewRows(
					[]string{
						"id",
						"restaurant_id",
						"category_id",
						"name",
						"description",
						"price",
						"available",
						"deleted_at",
						"created_at",
						"updated_at",
					}).
					AddRow(dishID, restaurantID, categoryID, "Burger", "Juicy beef burger", 499.00, true, nil, now, now),
			)

		created, err := repo.Create(ctx, dish)
		require.NoError(t, err)
		assert.Equal(t, dishID, created.ID)
		assert.Equal(t, "Burger", created.Name)
		assert.True(t, created.Available)
	})

	t.Run("foreign key violation (category not found)", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO dishes").
			WithArgs(dishID, restaurantID, categoryID, "Burger", "Juicy beef burger", 499.00, true, (*time.Time)(nil), now, now).
			WillReturnError(&pgconn.PgError{Code: "23503"})

		created, err := repo.Create(ctx, dish)
		require.ErrorIs(t, err, domain.ErrCategoryNotFound)
		assert.Nil(t, created)
	})
}

func TestDishRepo_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dishID := uuid.New()
	restaurantID := uuid.New()
	categoryID := uuid.New()
	now := time.Now().UTC()

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE id = \\$1").
			WithArgs(dishID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dishID, restaurantID, categoryID, "Pizza", "Cheese pizza", 550.00, true, nil, now, now),
			)

		dish, err := repo.GetByID(ctx, dishID)
		require.NoError(t, err)
		assert.Equal(t, dishID, dish.ID)
		assert.Equal(t, "Pizza", dish.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE id = \\$1").
			WithArgs(dishID).
			WillReturnError(pgx.ErrNoRows)

		dish, err := repo.GetByID(ctx, dishID)
		require.ErrorIs(t, err, domain.ErrDishNotFound)
		assert.Nil(t, dish)
	})
}

func TestDishRepo_GetByIDAndRestaurantID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dishID := uuid.New()
	restaurantID := uuid.New()
	categoryID := uuid.New()
	now := time.Now().UTC()

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(dishID, restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dishID, restaurantID, categoryID, "Pizza", "Cheese pizza", 550.00, true, nil, now, now),
			)

		dish, err := repo.GetByIDAndRestaurantID(ctx, dishID, restaurantID)
		require.NoError(t, err)
		assert.Equal(t, dishID, dish.ID)
		assert.Equal(t, "Pizza", dish.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(dishID, restaurantID).
			WillReturnError(pgx.ErrNoRows)

		dish, err := repo.GetByIDAndRestaurantID(ctx, dishID, restaurantID)
		require.ErrorIs(t, err, domain.ErrDishNotFound)
		assert.Nil(t, dish)
	})
}

func TestDishRepo_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dishID := uuid.New()
	restaurantID := uuid.New()
	categoryID := uuid.New()
	now := time.Now().UTC()

	dish := &domain.Dish{
		ID:           dishID,
		RestaurantID: restaurantID,
		CategoryID:   categoryID,
		Name:         "Updated Pizza",
		Description:  "Double cheese pizza",
		Price:        650.00,
		Available:    true,
		UpdatedAt:    now,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("UPDATE dishes").
			WithArgs(dishID, categoryID, "Updated Pizza", "Double cheese pizza", 650.00, true, (*time.Time)(nil), now, restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dishID, restaurantID, categoryID, "Updated Pizza", "Double cheese pizza", 650.00, true, nil, now, now),
			)

		updated, err := repo.Update(ctx, dish)
		require.NoError(t, err)
		assert.Equal(t, "Updated Pizza", updated.Name)
		assert.InEpsilon(t, 650.00, updated.Price, 0.0001)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("UPDATE dishes").
			WithArgs(dishID, categoryID, "Updated Pizza", "Double cheese pizza", 650.00, true, (*time.Time)(nil), now, restaurantID).
			WillReturnError(pgx.ErrNoRows)

		updated, err := repo.Update(ctx, dish)
		require.ErrorIs(t, err, domain.ErrDishNotFound)
		assert.Nil(t, updated)
	})
}

func TestDishRepo_SoftDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dishID := uuid.New()
	restaurantID := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("UPDATE dishes SET deleted_at = \\$3, updated_at = \\$3 WHERE id = \\$1 AND restaurant_id = \\$2 AND deleted_at IS NULL").
			WithArgs(dishID, restaurantID, now).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.SoftDelete(ctx, restaurantID, dishID, now)
		require.NoError(t, err)
	})

	t.Run("already deleted", func(t *testing.T) {
		mock.ExpectExec("UPDATE dishes SET deleted_at = \\$3, updated_at = \\$3 WHERE id = \\$1 AND restaurant_id = \\$2 AND deleted_at IS NULL").
			WithArgs(dishID, restaurantID, now).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		mock.ExpectQuery("SELECT deleted_at FROM dishes WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(dishID, restaurantID).
			WillReturnRows(pgxmock.NewRows([]string{"deleted_at"}).AddRow(&now))

		err := repo.SoftDelete(ctx, restaurantID, dishID, now)
		assert.ErrorIs(t, err, domain.ErrDishAlreadyDeleted)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec("UPDATE dishes SET deleted_at = \\$3, updated_at = \\$3 WHERE id = \\$1 AND restaurant_id = \\$2 AND deleted_at IS NULL").
			WithArgs(dishID, restaurantID, now).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		mock.ExpectQuery("SELECT deleted_at FROM dishes WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(dishID, restaurantID).
			WillReturnError(pgx.ErrNoRows)

		err := repo.SoftDelete(ctx, restaurantID, dishID, now)
		assert.ErrorIs(t, err, domain.ErrDishNotFound)
	})
}

func TestDishRepo_Restore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dishID := uuid.New()
	restaurantID := uuid.New()
	categoryID := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("UPDATE dishes SET deleted_at = NULL, updated_at = \\$3 WHERE id = \\$1 AND restaurant_id = \\$2 AND deleted_at IS NOT NULL").
			WithArgs(dishID, restaurantID, now).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dishID, restaurantID, categoryID, "Soup", "Hot tomato soup", 300.00, true, nil, now, now),
			)

		restored, err := repo.Restore(ctx, restaurantID, dishID, now)
		require.NoError(t, err)
		assert.Nil(t, restored.DeletedAt)
		assert.Equal(t, "Soup", restored.Name)
	})

	t.Run("dish not deleted", func(t *testing.T) {
		mock.ExpectQuery("UPDATE dishes SET deleted_at = NULL, updated_at = \\$3 WHERE id = \\$1 AND restaurant_id = \\$2 AND deleted_at IS NOT NULL").
			WithArgs(dishID, restaurantID, now).
			WillReturnError(pgx.ErrNoRows)

		mock.ExpectQuery("SELECT deleted_at FROM dishes WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(dishID, restaurantID).
			WillReturnRows(pgxmock.NewRows([]string{"deleted_at"}).AddRow(nil))

		restored, err := repo.Restore(ctx, restaurantID, dishID, now)
		require.ErrorIs(t, err, domain.ErrDishNotDeleted)
		assert.Nil(t, restored)
	})
}

func TestDishRepo_SetAvailability(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dishID := uuid.New()
	restaurantID := uuid.New()
	categoryID := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("UPDATE dishes SET available = \\$3, updated_at = \\$4 WHERE id = \\$1 AND restaurant_id = \\$2 AND deleted_at IS NULL").
			WithArgs(dishID, restaurantID, false, now).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dishID, restaurantID, categoryID, "Soup", "Hot tomato soup", 300.00, false, nil, now, now),
			)

		updated, err := repo.SetAvailability(ctx, restaurantID, dishID, false, now)
		require.NoError(t, err)
		assert.False(t, updated.Available)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("UPDATE dishes SET available = \\$3, updated_at = \\$4 WHERE id = \\$1 AND restaurant_id = \\$2 AND deleted_at IS NULL").
			WithArgs(dishID, restaurantID, true, now).
			WillReturnError(pgx.ErrNoRows)

		updated, err := repo.SetAvailability(ctx, restaurantID, dishID, true, now)
		require.ErrorIs(t, err, domain.ErrDishNotFound)
		assert.Nil(t, updated)
	})
}

func TestDishRepo_ListByRestaurantID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dish1 := uuid.New()
	dish2 := uuid.New()
	restaurantID := uuid.New()
	catID := uuid.New()
	now := time.Now().UTC()

	t.Run("active only", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE restaurant_id = \\$1 AND deleted_at IS NULL ORDER BY created_at ASC, name ASC").
			WithArgs(restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dish1, restaurantID, catID, "Dish 1", "Desc 1", 100.00, true, nil, now, now),
			)

		dishes, err := repo.ListByRestaurantID(ctx, restaurantID, false)
		require.NoError(t, err)
		assert.Len(t, dishes, 1)
		assert.Equal(t, "Dish 1", dishes[0].Name)
	})

	t.Run("include deleted", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE restaurant_id = \\$1 ORDER BY created_at ASC, name ASC").
			WithArgs(restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dish1, restaurantID, catID, "Dish 1", "Desc 1", 100.00, true, nil, now, now).
					AddRow(dish2, restaurantID, catID, "Dish 2", "Desc 2", 200.00, false, &now, now, now),
			)

		dishes, err := repo.ListByRestaurantID(ctx, restaurantID, true)
		require.NoError(t, err)
		assert.Len(t, dishes, 2)
	})
}

func TestDishRepo_GetByIDs(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	dish1 := uuid.New()
	dish2 := uuid.New()
	restaurantID := uuid.New()
	catID := uuid.New()
	now := time.Now().UTC()

	t.Run("empty ids", func(t *testing.T) {
		dishes, err := repo.GetByIDs(ctx, []uuid.UUID{})
		require.NoError(t, err)
		assert.Empty(t, dishes)
	})

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE id = ANY\\(\\$1\\)").
			WithArgs([]uuid.UUID{dish1, dish2}).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dish1, restaurantID, catID, "Dish 1", "", 100.00, true, nil, now, now).
					AddRow(dish2, restaurantID, catID, "Dish 2", "", 200.00, false, nil, now, now),
			)

		dishes, err := repo.GetByIDs(ctx, []uuid.UUID{dish1, dish2})
		require.NoError(t, err)
		assert.Len(t, dishes, 2)
	})
}

func TestDishRepo_GetMenuWithDishes(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewDishRepo(mock)
	ctx := context.Background()

	restaurantID := uuid.New()
	cat1 := uuid.New()
	cat2 := uuid.New()
	dish1 := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, name, sort_order FROM categories WHERE restaurant_id = \\$1 ORDER BY sort_order ASC, name ASC").
			WithArgs(restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "name", "sort_order"}).
					AddRow(cat1, restaurantID, "Mains", 1).
					AddRow(cat2, restaurantID, "Desserts", 2),
			)

		mock.ExpectQuery("SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at FROM dishes WHERE restaurant_id = \\$1 AND deleted_at IS NULL ORDER BY created_at ASC, name ASC").
			WithArgs(restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "category_id", "name", "description", "price", "available", "deleted_at", "created_at", "updated_at"}).
					AddRow(dish1, restaurantID, cat1, "Steak", "Ribeye steak", 1200.00, true, nil, now, now),
			)

		menu, err := repo.GetMenuWithDishes(ctx, restaurantID)
		require.NoError(t, err)
		require.Len(t, menu, 2)

		assert.Equal(t, "Mains", menu[0].Category.Name)
		require.Len(t, menu[0].Dishes, 1)
		assert.Equal(t, "Steak", menu[0].Dishes[0].Name)

		assert.Equal(t, "Desserts", menu[1].Category.Name)
		assert.Empty(t, menu[1].Dishes)
	})
}
