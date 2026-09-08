package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
	"github.com/hel1th/kitchen-service/internal/module/menu/usecase"
	"github.com/hel1th/kitchen-service/internal/shared/database"
)

var _ usecase.CategoryRepository = (*CategoryRepo)(nil)

// CategoryRepo implements usecase.CategoryRepository using PostgreSQL.
type CategoryRepo struct {
	db database.DBTX
}

// NewCategoryRepo creates a new CategoryRepo.
func NewCategoryRepo(db database.DBTX) *CategoryRepo {
	return &CategoryRepo{db: db}
}

// Create inserts a new category into the database.
func (r *CategoryRepo) Create(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	query := `
		INSERT INTO categories (id, restaurant_id, name, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, restaurant_id, name, sort_order
	`

	var created domain.Category
	err := r.db.QueryRow(ctx, query,
		category.ID,
		category.RestaurantID,
		category.Name,
		category.SortOrder,
	).Scan(
		&created.ID,
		&created.RestaurantID,
		&created.Name,
		&created.SortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("insert category: %w", err)
	}

	return &created, nil
}

// GetByID retrieves a category by its ID.
func (r *CategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	query := `
		SELECT id, restaurant_id, name, sort_order
		FROM categories
		WHERE id = $1
	`

	var cat domain.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cat.ID,
		&cat.RestaurantID,
		&cat.Name,
		&cat.SortOrder,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}

	return &cat, nil
}

// GetByIDAndRestaurantID retrieves a category by its ID and restaurant ID.
func (r *CategoryRepo) GetByIDAndRestaurantID(
	ctx context.Context,
	id, restaurantID uuid.UUID,
) (*domain.Category, error) {
	query := `
		SELECT id, restaurant_id, name, sort_order
		FROM categories
		WHERE id = $1 AND restaurant_id = $2
	`

	var cat domain.Category
	err := r.db.QueryRow(ctx, query, id, restaurantID).Scan(
		&cat.ID,
		&cat.RestaurantID,
		&cat.Name,
		&cat.SortOrder,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id and restaurant id: %w", err)
	}

	return &cat, nil
}

// ListByRestaurantID returns all categories for a given restaurant ordered by sort_order ASC, name ASC.
func (r *CategoryRepo) ListByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]domain.Category, error) {
	query := `
		SELECT id, restaurant_id, name, sort_order
		FROM categories
		WHERE restaurant_id = $1
		ORDER BY sort_order ASC, name ASC
	`

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var cat domain.Category
		if err := rows.Scan(
			&cat.ID,
			&cat.RestaurantID,
			&cat.Name,
			&cat.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category rows: %w", err)
	}

	return categories, nil
}

// Delete removes a category by ID and restaurant ID.
// If the category has dependent dishes, Postgres returns foreign_key_violation (23503),
// which is mapped to domain.ErrCategoryNotEmpty.
func (r *CategoryRepo) Delete(ctx context.Context, restaurantID, id uuid.UUID) error {
	query := `
		DELETE FROM categories
		WHERE id = $1 AND restaurant_id = $2
	`

	tag, err := r.db.Exec(ctx, query, id, restaurantID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.ErrCategoryNotEmpty
		}
		return fmt.Errorf("delete category: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}
