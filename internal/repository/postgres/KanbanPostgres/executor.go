package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBExecutor interface {
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PoolExecutor struct {
    DB *pgxpool.Pool
}

func (p PoolExecutor) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
    return p.DB.Exec(ctx, sql, args...)
}

func (p PoolExecutor) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
    return p.DB.Query(ctx, sql, args...)
}

func (p PoolExecutor) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
    return p.DB.QueryRow(ctx, sql, args...)
}

type TxExecutor struct {
    Tx pgx.Tx
}

func (t TxExecutor) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
    return t.Tx.Exec(ctx, sql, args...)
}

func (t TxExecutor) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
    return t.Tx.Query(ctx, sql, args...)
}

func (t TxExecutor) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
    return t.Tx.QueryRow(ctx, sql, args...)
}
