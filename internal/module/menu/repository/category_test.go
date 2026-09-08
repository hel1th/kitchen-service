package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
)

func TestCategoryRepo_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCategoryRepo(mock)
	ctx := context.Background()

	catID := uuid.New()
	restaurantID := uuid.New()

	category := &domain.Category{
		ID:           catID,
		RestaurantID: restaurantID,
		Name:         "Soups",
		SortOrder:    1,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO categories").
			WithArgs(catID, restaurantID, "Soups", 1).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "name", "sort_order"}).
					AddRow(catID, restaurantID, "Soups", 1),
			)

		created, err := repo.Create(ctx, category)
		require.NoError(t, err)
		assert.Equal(t, catID, created.ID)
		assert.Equal(t, "Soups", created.Name)
		assert.Equal(t, 1, created.SortOrder)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO categories").
			WithArgs(catID, restaurantID, "Soups", 1).
			WillReturnError(pgx.ErrTxClosed)

		created, err := repo.Create(ctx, category)
		require.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestCategoryRepo_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCategoryRepo(mock)
	ctx := context.Background()

	catID := uuid.New()
	restaurantID := uuid.New()

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, name, sort_order FROM categories WHERE id = \\$1").
			WithArgs(catID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "name", "sort_order"}).
					AddRow(catID, restaurantID, "Burgers", 2),
			)

		cat, err := repo.GetByID(ctx, catID)
		require.NoError(t, err)
		assert.Equal(t, catID, cat.ID)
		assert.Equal(t, "Burgers", cat.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, name, sort_order FROM categories WHERE id = \\$1").
			WithArgs(catID).
			WillReturnError(pgx.ErrNoRows)

		cat, err := repo.GetByID(ctx, catID)
		require.ErrorIs(t, err, domain.ErrCategoryNotFound)
		assert.Nil(t, cat)
	})
}

func TestCategoryRepo_GetByIDAndRestaurantID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCategoryRepo(mock)
	ctx := context.Background()

	catID := uuid.New()
	restaurantID := uuid.New()

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, name, sort_order FROM categories WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(catID, restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "name", "sort_order"}).
					AddRow(catID, restaurantID, "Burgers", 2),
			)

		cat, err := repo.GetByIDAndRestaurantID(ctx, catID, restaurantID)
		require.NoError(t, err)
		assert.Equal(t, catID, cat.ID)
		assert.Equal(t, restaurantID, cat.RestaurantID)
		assert.Equal(t, "Burgers", cat.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, name, sort_order FROM categories WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(catID, restaurantID).
			WillReturnError(pgx.ErrNoRows)

		cat, err := repo.GetByIDAndRestaurantID(ctx, catID, restaurantID)
		require.ErrorIs(t, err, domain.ErrCategoryNotFound)
		assert.Nil(t, cat)
	})
}

func TestCategoryRepo_ListByRestaurantID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCategoryRepo(mock)
	ctx := context.Background()

	restaurantID := uuid.New()
	cat1 := uuid.New()
	cat2 := uuid.New()

	t.Run("multiple rows", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, name, sort_order FROM categories WHERE restaurant_id = \\$1 ORDER BY sort_order ASC, name ASC").
			WithArgs(restaurantID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "restaurant_id", "name", "sort_order"}).
					AddRow(cat1, restaurantID, "Appetizers", 1).
					AddRow(cat2, restaurantID, "Mains", 2),
			)

		categories, err := repo.ListByRestaurantID(ctx, restaurantID)
		require.NoError(t, err)
		assert.Len(t, categories, 2)
		assert.Equal(t, "Appetizers", categories[0].Name)
		assert.Equal(t, "Mains", categories[1].Name)
	})

	t.Run("empty", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, restaurant_id, name, sort_order FROM categories WHERE restaurant_id = \\$1 ORDER BY sort_order ASC, name ASC").
			WithArgs(restaurantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "restaurant_id", "name", "sort_order"}))

		categories, err := repo.ListByRestaurantID(ctx, restaurantID)
		require.NoError(t, err)
		assert.Empty(t, categories)
	})
}

func TestCategoryRepo_Delete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewCategoryRepo(mock)
	ctx := context.Background()

	catID := uuid.New()
	restaurantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM categories WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(catID, restaurantID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(ctx, restaurantID, catID)
		require.NoError(t, err)
	})

	t.Run("not found (0 rows affected)", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM categories WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(catID, restaurantID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.Delete(ctx, restaurantID, catID)
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("foreign key violation (category not empty)", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM categories WHERE id = \\$1 AND restaurant_id = \\$2").
			WithArgs(catID, restaurantID).
			WillReturnError(&pgconn.PgError{Code: "23503"})

		err := repo.Delete(ctx, restaurantID, catID)
		assert.ErrorIs(t, err, domain.ErrCategoryNotEmpty)
	})
}
