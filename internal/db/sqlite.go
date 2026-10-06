package db

import (
	"context"
	"database/sql"
	"fmt"

	"db-tui/internal/config"
	_ "modernc.org/sqlite"
)

type sqliteDB struct {
	sql *sql.DB
}

func OpenSQLite(cfg config.Connection) (DB, error) {
	if cfg.FilePath == "" {
		return nil, fmt.Errorf("sqlite requires file_path")
	}
	sqldb, err := sql.Open("sqlite", cfg.FilePath)
	if err != nil {
		return nil, err
	}
	return &sqliteDB{sql: sqldb}, nil
}

func (d *sqliteDB) Close() error { return d.sql.Close() }

func (d *sqliteDB) Ping(ctx context.Context) error { return d.sql.PingContext(ctx) }

func (d *sqliteDB) Tables(ctx context.Context) ([]string, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type IN ('table','view') AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (d *sqliteDB) Query(ctx context.Context, q string) (Result, error) {
	return runGenericQuery(ctx, d.sql, q)
}
