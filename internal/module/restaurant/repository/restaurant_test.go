package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
	"github.com/hel1th/kitchen-service/internal/module/restaurant/usecase"
)

func TestRestaurantRepo_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRestaurantRepo(mock)
	ctx := context.Background()

	restID := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		restaurant := &domain.Restaurant{
			ID:         restID,
			Name:       "Burger Palace",
			Status:     domain.StatusOpen,
			WebhookURL: "http://webhook.local/orders",
			CreatedAt:  now,
		}

		mock.ExpectQuery("INSERT INTO restaurants").
			WithArgs(restID, "Burger Palace", "open", "http://webhook.local/orders", now).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "name", "status", "webhook_url", "created_at"}).
					AddRow(restID, "Burger Palace", "open", "http://webhook.local/orders", now),
			)

		created, err := repo.Create(ctx, restaurant)
		require.NoError(t, err)
		assert.Equal(t, restID, created.ID)
		assert.Equal(t, "Burger Palace", created.Name)
		assert.Equal(t, domain.StatusOpen, created.Status)
	})

	t.Run("invalid status", func(t *testing.T) {
		restaurant := &domain.Restaurant{
			ID:     restID,
			Name:   "Burger Palace",
			Status: domain.RestaurantStatus("invalid_status"),
		}

		created, err := repo.Create(ctx, restaurant)
		assert.ErrorIs(t, err, domain.ErrInvalidRestaurantStatus)
		assert.Nil(t, created)
	})

	t.Run("db error", func(t *testing.T) {
		restaurant := &domain.Restaurant{
			ID:        restID,
			Name:      "Burger Palace",
			Status:    domain.StatusOpen,
			CreatedAt: now,
		}

		mock.ExpectQuery("INSERT INTO restaurants").
			WithArgs(restID, "Burger Palace", "open", "", now).
			WillReturnError(errors.New("db error"))

		created, err := repo.Create(ctx, restaurant)
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestRestaurantRepo_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRestaurantRepo(mock)
	ctx := context.Background()

	restID := uuid.New()
	now := time.Now().UTC()

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, status, webhook_url, created_at FROM restaurants WHERE id = \\$1").
			WithArgs(restID).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "name", "status", "webhook_url", "created_at"}).
					AddRow(restID, "Pizza Hub", "closed", "http://pizza.local/webhook", now),
			)

		rest, err := repo.GetByID(ctx, restID)
		require.NoError(t, err)
		assert.Equal(t, restID, rest.ID)
		assert.Equal(t, "Pizza Hub", rest.Name)
		assert.Equal(t, domain.StatusClosed, rest.Status)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, status, webhook_url, created_at FROM restaurants WHERE id = \\$1").
			WithArgs(restID).
			WillReturnError(pgx.ErrNoRows)

		rest, err := repo.GetByID(ctx, restID)
		assert.ErrorIs(t, err, domain.ErrRestaurantNotFound)
		assert.Nil(t, rest)
	})
}

func TestRestaurantRepo_UpdateStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRestaurantRepo(mock)
	ctx := context.Background()

	restID := uuid.New()
	now := time.Now().UTC()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("UPDATE restaurants SET status = \\$2 WHERE id = \\$1").
			WithArgs(restID, "closed").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "name", "status", "webhook_url", "created_at"}).
					AddRow(restID, "Taco Corner", "closed", "http://taco.local/webhook", now),
			)

		updated, err := repo.UpdateStatus(ctx, restID, domain.StatusClosed)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusClosed, updated.Status)
	})

	t.Run("invalid status", func(t *testing.T) {
		updated, err := repo.UpdateStatus(ctx, restID, domain.RestaurantStatus("invalid"))
		assert.ErrorIs(t, err, domain.ErrInvalidRestaurantStatus)
		assert.Nil(t, updated)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("UPDATE restaurants SET status = \\$2 WHERE id = \\$1").
			WithArgs(restID, "open").
			WillReturnError(pgx.ErrNoRows)

		updated, err := repo.UpdateStatus(ctx, restID, domain.StatusOpen)
		assert.ErrorIs(t, err, domain.ErrRestaurantNotFound)
		assert.Nil(t, updated)
	})
}

func TestRestaurantRepo_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRestaurantRepo(mock)
	ctx := context.Background()

	rest1 := uuid.New()
	rest2 := uuid.New()
	now := time.Now().UTC()

	t.Run("with status filter", func(t *testing.T) {
		status := domain.StatusOpen
		statusStr := "open"
		filter := usecase.ListFilter{
			Status: &status,
			Limit:  10,
			Offset: 0,
		}

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM restaurants WHERE \\(\\$1::text IS NULL OR status = \\$1\\)").
			WithArgs(&statusStr).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery("SELECT id, name, status, webhook_url, created_at FROM restaurants WHERE \\(\\$1::text IS NULL OR status = \\$1\\) ORDER BY created_at ASC, name ASC LIMIT \\$2 OFFSET \\$3").
			WithArgs(&statusStr, 10, 0).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "name", "status", "webhook_url", "created_at"}).
					AddRow(rest1, "Rest 1", "open", "", now),
			)

		list, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, list, 1)
		assert.Equal(t, "Rest 1", list[0].Name)
	})

	t.Run("without filter (nil status)", func(t *testing.T) {
		filter := usecase.ListFilter{
			Limit:  20,
			Offset: 0,
		}

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM restaurants WHERE \\(\\$1::text IS NULL OR status = \\$1\\)").
			WithArgs((*string)(nil)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))

		mock.ExpectQuery("SELECT id, name, status, webhook_url, created_at FROM restaurants WHERE \\(\\$1::text IS NULL OR status = \\$1\\) ORDER BY created_at ASC, name ASC LIMIT \\$2 OFFSET \\$3").
			WithArgs((*string)(nil), 20, 0).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "name", "status", "webhook_url", "created_at"}).
					AddRow(rest1, "Rest 1", "open", "", now).
					AddRow(rest2, "Rest 2", "closed", "", now),
			)

		list, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, list, 2)
	})
}
