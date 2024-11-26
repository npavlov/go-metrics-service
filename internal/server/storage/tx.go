package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"github.com/npavlov/go-metrics-service/internal/server/dbmanager"
)

type WithTxFunc func(ctx context.Context, tx pgx.Tx) error

func WithTx(ctx context.Context, db dbmanager.PgxPool, fn WithTxFunc) error {
	//nolint:exhaustruct
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errors.Wrap(err, "db.BeginTxx()")
	}

	if err = fn(ctx, tx); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return errors.Wrap(err, "Tx.Rollback")
		}

		return errors.Wrap(err, "Tx.WithTxFunc")
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "Tx.Commit")
	}

	return nil
}
