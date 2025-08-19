package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handle struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, conn string) (*Handle, error) {
	cfg, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 5
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Handle{Pool: pool}, nil
}

func (h *Handle) Close() {
	if h != nil && h.Pool != nil {
		h.Pool.Close()
	}
}
