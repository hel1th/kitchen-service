package domain

import "errors"

var (
	// ErrRestaurantNotFound indicates that a restaurant was not found.
	ErrRestaurantNotFound = errors.New("restaurant not found")

	// ErrInvalidRestaurantStatus indicates an unrecognized or invalid restaurant status.
	ErrInvalidRestaurantStatus = errors.New("invalid restaurant status")
)
