// Package repository — Postgres (pgx/v5/stdlib + go-jet). Только доступ к данным,
// без бизнес-решений: что писать и как реагировать решает service.
package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx"
)

//go:embed schema.sql
var schemaSQL string

var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

// Open connects to Postgres and applies schema.sql (idempotent CREATE IF NOT EXISTS).
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() { _ = s.db.Close() }

// DB exposes the underlying handle for callers that need raw SQL (test setup, ops).
func (s *Store) DB() *sql.DB { return s.db }

// isFKViolation: Postgres SQLSTATE 23503 (foreign_key_violation).
func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
