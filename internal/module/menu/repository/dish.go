package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/hel1th/kitchen-service/internal/module/menu/domain"
	"github.com/hel1th/kitchen-service/internal/module/menu/usecase"
	"github.com/hel1th/kitchen-service/internal/shared/database"
)

var _ usecase.DishRepository = (*DishRepo)(nil)

// DishRepo implements usecase.DishRepository using PostgreSQL.
type DishRepo struct {
	db database.DBTX
}

// NewDishRepo creates a new DishRepo.
func NewDishRepo(db database.DBTX) *DishRepo {
	return &DishRepo{db: db}
}

// Create inserts a new dish into the database.
// If category_id does not exist, foreign_key_violation (23503) is mapped to domain.ErrCategoryNotFound.
func (r *DishRepo) Create(ctx context.Context, dish *domain.Dish) (*domain.Dish, error) {
	if dish.ID == uuid.Nil {
		dish.ID = uuid.New()
	}
	if dish.CreatedAt.IsZero() {
		dish.CreatedAt = time.Now().UTC()
	}
	if dish.UpdatedAt.IsZero() {
		dish.UpdatedAt = dish.CreatedAt
	}

	query := `
		INSERT INTO dishes (
			id, restaurant_id, category_id, name,
			description, price, available, deleted_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, restaurant_id, category_id, name,
				description, price, available, deleted_at, created_at, updated_at`

	var created domain.Dish
	err := r.db.QueryRow(ctx, query,
		dish.ID,
		dish.RestaurantID,
		dish.CategoryID,
		dish.Name,
		dish.Description,
		dish.Price,
		dish.Available,
		dish.DeletedAt,
		dish.CreatedAt,
		dish.UpdatedAt,
	).Scan(
		&created.ID,
		&created.RestaurantID,
		&created.CategoryID,
		&created.Name,
		&created.Description,
		&created.Price,
		&created.Available,
		&created.DeletedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("insert dish: %w", err)
	}

	return &created, nil
}

// GetByID retrieves a dish by its primary ID.
func (r *DishRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dish, error) {
	query := `
		SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
		FROM dishes
		WHERE id = $1
	`

	var dish domain.Dish
	err := r.db.QueryRow(ctx, query, id).Scan(
		&dish.ID,
		&dish.RestaurantID,
		&dish.CategoryID,
		&dish.Name,
		&dish.Description,
		&dish.Price,
		&dish.Available,
		&dish.DeletedAt,
		&dish.CreatedAt,
		&dish.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDishNotFound
		}
		return nil, fmt.Errorf("get dish by id: %w", err)
	}

	return &dish, nil
}

// GetByIDAndRestaurantID retrieves a dish by its ID and restaurant ID.
func (r *DishRepo) GetByIDAndRestaurantID(
	ctx context.Context,
	id, restaurantID uuid.UUID,
) (*domain.Dish, error) {
	query := `
		SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
		FROM dishes
		WHERE id = $1 AND restaurant_id = $2
	`

	var dish domain.Dish
	err := r.db.QueryRow(ctx, query, id, restaurantID).Scan(
		&dish.ID,
		&dish.RestaurantID,
		&dish.CategoryID,
		&dish.Name,
		&dish.Description,
		&dish.Price,
		&dish.Available,
		&dish.DeletedAt,
		&dish.CreatedAt,
		&dish.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDishNotFound
		}
		return nil, fmt.Errorf("get dish by id and restaurant id: %w", err)
	}

	return &dish, nil
}

// Update updates an existing dish's mutable properties.
func (r *DishRepo) Update(ctx context.Context, dish *domain.Dish) (*domain.Dish, error) {
	query := `
		UPDATE dishes
		SET category_id = $2,
		    name = $3,
		    description = $4,
		    price = $5,
		    available = $6,
		    deleted_at = $7,
		    updated_at = $8
		WHERE id = $1 AND restaurant_id = $9
		RETURNING id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
	`

	var updated domain.Dish
	err := r.db.QueryRow(ctx, query,
		dish.ID,
		dish.CategoryID,
		dish.Name,
		dish.Description,
		dish.Price,
		dish.Available,
		dish.DeletedAt,
		dish.UpdatedAt,
		dish.RestaurantID,
	).Scan(
		&updated.ID,
		&updated.RestaurantID,
		&updated.CategoryID,
		&updated.Name,
		&updated.Description,
		&updated.Price,
		&updated.Available,
		&updated.DeletedAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDishNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("update dish: %w", err)
	}

	return &updated, nil
}

// SoftDelete sets deleted_at = now() for a non-deleted dish.
// If the dish was already soft-deleted, it returns domain.ErrDishAlreadyDeleted.
func (r *DishRepo) SoftDelete(ctx context.Context, restaurantID, id uuid.UUID, at time.Time) error {
	query := `
		UPDATE dishes
		SET deleted_at = $3,
		    updated_at = $3
		WHERE id = $1 AND restaurant_id = $2 AND deleted_at IS NULL
	`

	tag, err := r.db.Exec(ctx, query, id, restaurantID, at)
	if err != nil {
		return fmt.Errorf("soft delete dish: %w", err)
	}

	if tag.RowsAffected() == 0 { //nolint:nestif
		var deletedAt *time.Time
		checkQuery := `SELECT deleted_at FROM dishes WHERE id = $1 AND restaurant_id = $2`
		errCheck := r.db.QueryRow(ctx, checkQuery, id, restaurantID).Scan(&deletedAt)
		if errCheck != nil {
			if errors.Is(errCheck, pgx.ErrNoRows) {
				return domain.ErrDishNotFound
			}
			return fmt.Errorf("check dish deleted status: %w", errCheck)
		}
		if deletedAt != nil {
			return domain.ErrDishAlreadyDeleted
		}
		return domain.ErrDishNotFound
	}

	return nil
}

// Restore resets deleted_at = NULL for a soft-deleted dish.
// If the dish is not deleted, it returns domain.ErrDishNotDeleted.
func (r *DishRepo) Restore(
	ctx context.Context,
	restaurantID, id uuid.UUID,
	at time.Time,
) (*domain.Dish, error) {
	query := `
		UPDATE dishes
		SET deleted_at = NULL,
		    updated_at = $3
		WHERE id = $1 AND restaurant_id = $2 AND deleted_at IS NOT NULL
		RETURNING id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
	`

	var restored domain.Dish
	err := r.db.QueryRow(ctx, query, id, restaurantID, at).Scan(
		&restored.ID,
		&restored.RestaurantID,
		&restored.CategoryID,
		&restored.Name,
		&restored.Description,
		&restored.Price,
		&restored.Available,
		&restored.DeletedAt,
		&restored.CreatedAt,
		&restored.UpdatedAt,
	)
	if err != nil { //nolint:nestif
		if errors.Is(err, pgx.ErrNoRows) {
			var deletedAt *time.Time
			checkQuery := `SELECT deleted_at FROM dishes WHERE id = $1 AND restaurant_id = $2`
			errCheck := r.db.QueryRow(ctx, checkQuery, id, restaurantID).Scan(&deletedAt)
			if errCheck != nil {
				if errors.Is(errCheck, pgx.ErrNoRows) {
					return nil, domain.ErrDishNotFound
				}
				return nil, fmt.Errorf("check dish restore status: %w", errCheck)
			}
			if deletedAt == nil {
				return nil, domain.ErrDishNotDeleted
			}
			return nil, domain.ErrDishNotFound
		}
		return nil, fmt.Errorf("restore dish: %w", err)
	}

	return &restored, nil
}

// SetAvailability toggles the available status of an active dish.
func (r *DishRepo) SetAvailability(
	ctx context.Context,
	restaurantID, id uuid.UUID,
	available bool,
	at time.Time,
) (*domain.Dish, error) {
	query := `
		UPDATE dishes
		SET available = $3,
		    updated_at = $4
		WHERE id = $1 AND restaurant_id = $2 AND deleted_at IS NULL
		RETURNING id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
	`

	var updated domain.Dish
	err := r.db.QueryRow(ctx, query, id, restaurantID, available, at).Scan(
		&updated.ID,
		&updated.RestaurantID,
		&updated.CategoryID,
		&updated.Name,
		&updated.Description,
		&updated.Price,
		&updated.Available,
		&updated.DeletedAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDishNotFound
		}
		return nil, fmt.Errorf("set dish availability: %w", err)
	}

	return &updated, nil
}

// ListByRestaurantID retrieves dishes for a restaurant, optionally including soft-deleted dishes.
func (r *DishRepo) ListByRestaurantID(
	ctx context.Context,
	restaurantID uuid.UUID,
	includeDeleted bool,
) ([]domain.Dish, error) {
	var query string
	if includeDeleted {
		query = `
			SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
			FROM dishes
			WHERE restaurant_id = $1
			ORDER BY created_at ASC, name ASC
		`
	} else {
		query = `
			SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
			FROM dishes
			WHERE restaurant_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC, name ASC
		`
	}

	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("query dishes by restaurant id: %w", err)
	}
	defer rows.Close()

	dishes := make([]domain.Dish, 0)
	for rows.Next() {
		var dish domain.Dish
		if err := rows.Scan(
			&dish.ID,
			&dish.RestaurantID,
			&dish.CategoryID,
			&dish.Name,
			&dish.Description,
			&dish.Price,
			&dish.Available,
			&dish.DeletedAt,
			&dish.CreatedAt,
			&dish.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dish: %w", err)
		}
		dishes = append(dishes, dish)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dish rows: %w", err)
	}

	return dishes, nil
}

// GetByIDs fetches multiple dishes by their IDs using ANY($1).
func (r *DishRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Dish, error) {
	if len(ids) == 0 {
		return []domain.Dish{}, nil
	}

	query := `
		SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
		FROM dishes
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("query dishes by ids: %w", err)
	}
	defer rows.Close()

	dishes := make([]domain.Dish, 0, len(ids))
	for rows.Next() {
		var dish domain.Dish
		if err := rows.Scan(
			&dish.ID,
			&dish.RestaurantID,
			&dish.CategoryID,
			&dish.Name,
			&dish.Description,
			&dish.Price,
			&dish.Available,
			&dish.DeletedAt,
			&dish.CreatedAt,
			&dish.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dish: %w", err)
		}
		dishes = append(dishes, dish)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dish rows by ids: %w", err)
	}

	return dishes, nil
}

// GetMenuWithDishes returns categories along with their active (non-deleted) dishes for a given restaurant.
func (r *DishRepo) GetMenuWithDishes(
	ctx context.Context,
	restaurantID uuid.UUID,
) ([]domain.CategoryWithDishes, error) { //nolint:funlen
	categoriesQuery := `
		SELECT id, restaurant_id, name, sort_order
		FROM categories
		WHERE restaurant_id = $1
		ORDER BY sort_order ASC, name ASC
	`

	catRows, err := r.db.Query(ctx, categoriesQuery, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("query categories for menu: %w", err)
	}
	defer catRows.Close()

	categories := make([]domain.Category, 0)
	for catRows.Next() {
		var cat domain.Category
		if err := catRows.Scan(
			&cat.ID,
			&cat.RestaurantID,
			&cat.Name,
			&cat.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scan category for menu: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := catRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories for menu: %w", err)
	}

	dishesQuery := `
		SELECT id, restaurant_id, category_id, name, description, price, available, deleted_at, created_at, updated_at
		FROM dishes
		WHERE restaurant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC, name ASC
	`

	dishRows, err := r.db.Query(ctx, dishesQuery, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("query dishes for menu: %w", err)
	}
	defer dishRows.Close()

	dishesByCat := make(map[uuid.UUID][]domain.Dish)
	for dishRows.Next() {
		var dish domain.Dish
		if err := dishRows.Scan(
			&dish.ID,
			&dish.RestaurantID,
			&dish.CategoryID,
			&dish.Name,
			&dish.Description,
			&dish.Price,
			&dish.Available,
			&dish.DeletedAt,
			&dish.CreatedAt,
			&dish.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dish for menu: %w", err)
		}
		dishesByCat[dish.CategoryID] = append(dishesByCat[dish.CategoryID], dish)
	}

	if err := dishRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dishes for menu: %w", err)
	}

	result := make([]domain.CategoryWithDishes, 0, len(categories))
	for _, cat := range categories {
		dList := dishesByCat[cat.ID]
		if dList == nil {
			dList = []domain.Dish{}
		}
		result = append(result, domain.CategoryWithDishes{
			Category: cat,
			Dishes:   dList,
		})
	}

	return result, nil
}
