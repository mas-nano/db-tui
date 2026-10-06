// Package db menyediakan abstraksi satu interface untuk
// SQLite, PostgreSQL, dan MariaDB/MySQL.
package db

import (
	"context"
	"database/sql"
)

// Result hasil query generik untuk dirender ke tabel TUI.
type Result struct {
	Columns []string
	Rows    [][]string
	// RowsAffected untuk INSERT/UPDATE/DELETE, -1 kalau tidak relevan.
	RowsAffected int64
}

type DB interface {
	Close() error
	Ping(ctx context.Context) error
	// Tables kembalikan daftar tabel/view.
	Tables(ctx context.Context) ([]string, error)
	// Query jalankan query bebas, SELECT -> Columns+Rows, sisanya -> RowsAffected.
	Query(ctx context.Context, q string) (Result, error)
}

var _ = sql.ErrNoRows
