package pgxdt

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) Transactor {
	return &transactor{pool: pool}
}

func (t *transactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if getTx(ctx) != nil {
		return fn(ctx)
	}

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	wrapped := &txImpl{tx: tx}
	ctx = putTx(ctx, wrapped)

	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
