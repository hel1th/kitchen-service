package domain

import "errors"

var (
	// ErrDishNotFound indicates that the requested dish does not exist.
	ErrDishNotFound = errors.New("dish not found")

	// ErrCategoryNotFound indicates that the requested category does not exist.
	ErrCategoryNotFound = errors.New("category not found")

	// ErrCategoryNotEmpty indicates that the category cannot be deleted because it contains dishes.
	ErrCategoryNotEmpty = errors.New("category is not empty")

	// ErrDishAlreadyDeleted indicates an operation was attempted on an already soft-deleted dish.
	ErrDishAlreadyDeleted = errors.New("dish is already deleted")

	// ErrDishNotDeleted indicates a restore operation was attempted on an active dish.
	ErrDishNotDeleted = errors.New("dish is not deleted")
)
