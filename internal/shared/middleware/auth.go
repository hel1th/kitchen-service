package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

// UserRepository is an interface for the auth middleware to upsert users.
type UserRepository interface {
	Upsert(ctx context.Context, userID int64) error
}

// ExtractUserID reads X-User-Id, parses it, calls Upsert, and puts it into the context.
func ExtractUserID(userRepo UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := r.Header.Get("X-User-Id")
			if userIDStr != "" {
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err == nil {
					// Upsert user in the background or synchronously?
					// Prompt says "дергает UserRepo.Upsert." so synchronously.
					_ = userRepo.Upsert(r.Context(), userID)
					ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
