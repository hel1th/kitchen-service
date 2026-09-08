package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/hel1th/kitchen-service/internal/module/restaurant/domain"
	"github.com/hel1th/kitchen-service/internal/module/restaurant/usecase"
	"github.com/hel1th/kitchen-service/internal/shared/database"
)

var _ usecase.RestaurantRepository = (*RestaurantRepo)(nil)

// RestaurantRepo implements usecase.RestaurantRepository using PostgreSQL.
type RestaurantRepo struct {
	db database.DBTX
}

// NewRestaurantRepo creates a new RestaurantRepo.
func NewRestaurantRepo(db database.DBTX) *RestaurantRepo {
	return &RestaurantRepo{db: db}
}

// Create inserts a new restaurant into the database.
func (r *RestaurantRepo) Create(
	ctx context.Context,
	restaurant *domain.Restaurant,
) (*domain.Restaurant, error) {
	if restaurant.ID == uuid.Nil {
		restaurant.ID = uuid.New()
	}
	if restaurant.Status == "" {
		restaurant.Status = domain.StatusOpen
	}
	if !restaurant.Status.IsValid() {
		return nil, domain.ErrInvalidRestaurantStatus
	}
	if restaurant.CreatedAt.IsZero() {
		restaurant.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO restaurants (id, name, status, webhook_url, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, status, webhook_url, created_at
	`

	var created domain.Restaurant
	var statusStr string
	err := r.db.QueryRow(ctx, query,
		restaurant.ID,
		restaurant.Name,
		string(restaurant.Status),
		restaurant.WebhookURL,
		restaurant.CreatedAt,
	).Scan(
		&created.ID,
		&created.Name,
		&statusStr,
		&created.WebhookURL,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert restaurant: %w", err)
	}
	created.Status = domain.RestaurantStatus(statusStr)

	return &created, nil
}

// GetByID retrieves a restaurant by its ID.
func (r *RestaurantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	query := `
		SELECT id, name, status, webhook_url, created_at
		FROM restaurants
		WHERE id = $1
	`

	var rest domain.Restaurant
	var statusStr string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rest.ID,
		&rest.Name,
		&statusStr,
		&rest.WebhookURL,
		&rest.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRestaurantNotFound
		}
		return nil, fmt.Errorf("get restaurant by id: %w", err)
	}
	rest.Status = domain.RestaurantStatus(statusStr)

	return &rest, nil
}

// UpdateStatus updates the operational status of a restaurant.
func (r *RestaurantRepo) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.RestaurantStatus,
) (*domain.Restaurant, error) {
	if !status.IsValid() {
		return nil, domain.ErrInvalidRestaurantStatus
	}

	query := `
		UPDATE restaurants
		SET status = $2
		WHERE id = $1
		RETURNING id, name, status, webhook_url, created_at
	`

	var updated domain.Restaurant
	var statusStr string
	err := r.db.QueryRow(ctx, query, id, string(status)).Scan(
		&updated.ID,
		&updated.Name,
		&statusStr,
		&updated.WebhookURL,
		&updated.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRestaurantNotFound
		}
		return nil, fmt.Errorf("update restaurant status: %w", err)
	}
	updated.Status = domain.RestaurantStatus(statusStr)

	return &updated, nil
}

// List returns a paginated list of restaurants matching the filter, along with the total count.
func (r *RestaurantRepo) List(
	ctx context.Context,
	filter usecase.ListFilter,
) ([]domain.Restaurant, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := max(filter.Offset, 0)

	var statusParam *string
	if filter.Status != nil {
		s := string(*filter.Status)
		statusParam = &s
	}

	countQuery := `
		SELECT COUNT(*)
		FROM restaurants
		WHERE ($1::text IS NULL OR status = $1)
	`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, statusParam).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count restaurants: %w", err)
	}

	query := `
		SELECT id, name, status, webhook_url, created_at
		FROM restaurants
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY created_at ASC, name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, statusParam, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query restaurants: %w", err)
	}
	defer rows.Close()

	restaurants := make([]domain.Restaurant, 0)
	for rows.Next() {
		var rest domain.Restaurant
		var statusStr string
		if err := rows.Scan(
			&rest.ID,
			&rest.Name,
			&statusStr,
			&rest.WebhookURL,
			&rest.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan restaurant: %w", err)
		}
		rest.Status = domain.RestaurantStatus(statusStr)
		restaurants = append(restaurants, rest)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate restaurant rows: %w", err)
	}

	return restaurants, total, nil
}
