package postgrebase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	ConnString string
}

func NewPgConn(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return conn, nil
}
