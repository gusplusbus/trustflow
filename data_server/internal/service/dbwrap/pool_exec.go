package dbwrap

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolExec adapts *pgxpool.Pool to the service.DBLike interface
type PoolExec struct{ Pool *pgxpool.Pool }

func (p PoolExec) Exec(ctx context.Context, sql string, args ...any) (any, error) {
	return p.Pool.Exec(ctx, sql, args...)
}
