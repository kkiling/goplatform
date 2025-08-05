package sqlitebase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/kkiling/goplatform/storagebase"
	"strings"

	"github.com/kkiling/goplatform/log"
	"github.com/mattn/go-sqlite3"
)

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type txKey struct{}

type Config struct {
	DSN string
}

type Storage struct {
	config Config
	db     *sql.DB
	logger log.Logger
}

func NewStorage(config Config, logger log.Logger) (*Storage, error) {
	// Обязательные параметры в DSN
	dsn := fmt.Sprintf("%s?_busy_timeout=5000&_foreign_keys=on", config.DSN)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Дополнительные настройки через PRAGMA
	_, err = db.Exec(`
        PRAGMA journal_mode = WAL;
        PRAGMA synchronous = NORMAL;
        PRAGMA journal_size_limit = 16384;
        PRAGMA wal_autocheckpoint = 100;
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to set database pragmas: %w", err)
	}

	return &Storage{config: config, db: db, logger: logger}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) HandleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return storagebase.ErrNotFound
	}

	// Проверяем, является ли ошибка ошибкой уникальности
	sqliteErr := sqlite3.Error{}
	if ok := errors.As(err, &sqliteErr); ok {
		if strings.Contains(sqliteErr.Error(), "FOREIGN KEY constraint failed") {
			return err
		}
		if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
			return storagebase.ErrAlreadyExists
		}
	}

	return err
}

func (s *Storage) Next(ctx context.Context) SQLExecutor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		// Используем внутренний db из Tx
		return tx
	}
	return s.db
}

func (s *Storage) RunTransaction(ctx context.Context, txFunc func(ctxTx context.Context) error) error {
	// Начало транзакции
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Отложенный вызов Rollback в случае ошибки
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				s.logger.Errorf("failed to rollback transaction: %v", err)
			}
		}
	}()

	ctxTx := context.WithValue(ctx, txKey{}, tx)
	if tErr := txFunc(ctxTx); tErr != nil {
		return fmt.Errorf("run transaction: %w", tErr)
	}

	// Фиксация транзакции
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
