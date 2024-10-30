package repository

import (
	"context"

	"github.com/pkg/errors"
)

// PgxPoolInterface defines an interface for a PostgreSQL connection pool.
type PgxPoolInterface interface {
	Ping(ctx context.Context) error
}

type Repository interface {
	Ping(ctx context.Context) error
}

type DBRepository struct {
	db PgxPoolInterface
}

func NewDBRepository(db PgxPoolInterface) *DBRepository {
	return &DBRepository{
		db: db,
	}
}

func (r *DBRepository) Ping(ctx context.Context) error {
	return errors.Wrap(r.db.Ping(ctx), "failed to ping DB")
}
