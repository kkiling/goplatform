package postgrebase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kkiling/goplatform/storagebase"
)

// Коды ошибок pg
const (
	pgErrCodeUniqueViolation = "23505"
	pgErrForeignKeyViolation = "23503"
)

type SQLExecutor interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type contextKey string

const txKey contextKey = "pgx_tx"

type Storage struct {
	config Config
	pool   *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool: pool,
	}
}

func (s *Storage) Close() error {
	s.pool.Close()
	return nil
}

func (s *Storage) HandleError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return storagebase.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
		switch pgErr.Code {
		case pgErrForeignKeyViolation:
			return storagebase.ErrForeignKeyViolation
		case pgErrCodeUniqueViolation:
			return storagebase.ErrAlreadyExists
		}

		return newErr
	}
	return err
}

func (s *Storage) Next(ctx context.Context) SQLExecutor {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return s.pool
}

func (s *Storage) RunTransaction(ctx context.Context, txFunc func(ctxTx context.Context) error) error {
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return txFunc(ctx)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("con.BeginTx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey, tx)
	err = txFunc(txCtx)

	if err != nil {
		rollBackErr := tx.Rollback(ctx)
		if rollBackErr != nil {
			return fmt.Errorf("tx.Rollback: %w", rollBackErr)
		}
		return s.HandleError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", s.HandleError(err))
	}

	return nil
}
