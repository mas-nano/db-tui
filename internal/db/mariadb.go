package db

import (
	"context"
	"database/sql"
	"fmt"

	"db-tui/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

type mariaDB struct {
	sql *sql.DB
}

func OpenMariaDB(cfg config.Connection) (DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &mariaDB{sql: sqldb}, nil
}

func (d *mariaDB) Close() error { return d.sql.Close() }

func (d *mariaDB) Ping(ctx context.Context) error { return d.sql.PingContext(ctx) }

func (d *mariaDB) Tables(ctx context.Context) ([]string, error) {
	rows, err := d.sql.QueryContext(ctx, `SHOW FULL TABLES WHERE TABLE_TYPE LIKE 'BASE TABLE'`)
	if err != nil {
		// fallback: SHOW TABLES
		rows, err = d.sql.QueryContext(ctx, `SHOW TABLES`)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var a, b sql.NullString
		// SHOW FULL TABLES -> 2 kolom, SHOW TABLES -> 1 kolom
		cols, _ := rows.Columns()
		if len(cols) == 1 {
			var n string
			if err := rows.Scan(&n); err != nil {
				return nil, err
			}
			out = append(out, n)
		} else {
			if err := rows.Scan(&a, &b); err != nil {
				return nil, err
			}
			out = append(out, a.String)
		}
	}
	return out, rows.Err()
}

func (d *mariaDB) Query(ctx context.Context, q string) (Result, error) {
	return runGenericQuery(ctx, d.sql, q)
}
