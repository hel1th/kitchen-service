package repository

import (
	"context"
	"fmt"

	"github.com/hel1th/kitchen-service/internal/module/order/usecase"
	"github.com/hel1th/kitchen-service/internal/shared/database"
)

var _ usecase.UserRepository = (*UserRepo)(nil)

type UserRepo struct {
	db database.DBTX
}

func NewUserRepo(db database.DBTX) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Upsert(ctx context.Context, id int64) error {
	query := `INSERT INTO users(id) VALUES ($1) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	return nil
}
