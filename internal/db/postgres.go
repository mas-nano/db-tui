package db

import (
	"context"
	"database/sql"
	"fmt"

	"db-tui/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type postgresDB struct {
	sql *sql.DB
}

func OpenPostgres(cfg config.Connection) (DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
	pgxCfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	sqldb := stdlib.OpenDB(*pgxCfg)
	return &postgresDB{sql: sqldb}, nil
}

func (d *postgresDB) Close() error { return d.sql.Close() }

func (d *postgresDB) Ping(ctx context.Context) error { return d.sql.PingContext(ctx) }

func (d *postgresDB) Tables(ctx context.Context) ([]string, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename`)
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

func (d *postgresDB) Query(ctx context.Context, q string) (Result, error) {
	return runGenericQuery(ctx, d.sql, q)
}
