package pgxdt

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbImpl struct {
	pool *pgxpool.Pool
}

func NewPgxdt(pool *pgxpool.Pool) Pgxdt {
	return &dbImpl{pool: pool}
}

func (d *dbImpl) Exec(ctx context.Context, sql string, args ...any) (commandTag pgconn.CommandTag, err error) {
	if tx := getTx(ctx); tx != nil {
		return tx.Exec(ctx, sql, args...)
	}
	return d.pool.Exec(ctx, sql, args...)
}

func (d *dbImpl) Query(ctx context.Context, sql string, args ...any) (rows pgx.Rows, err error) {
	if tx := getTx(ctx); tx != nil {
		return tx.Query(ctx, sql, args...)
	}
	return d.pool.Query(ctx, sql, args...)
}

func (d *dbImpl) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := getTx(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}
	return d.pool.QueryRow(ctx, sql, args...)
}

func (d *dbImpl) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if tx := getTx(ctx); tx != nil {
		return tx.WithTx(ctx, fn)
	}

	newTx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer newTx.Rollback(ctx)

	wrapped := &txImpl{tx: newTx}
	ctx = putTx(ctx, wrapped)

	if err := fn(ctx); err != nil {
		return err
	}

	return newTx.Commit(ctx)
}
