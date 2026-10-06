package db

import (
	"context"
	"database/sql"
	"fmt"

	"db-tui/internal/config"
)

// Open pilih driver berdasarkan cfg.Driver.
func Open(cfg config.Connection) (DB, error) {
	switch cfg.Driver {
	case "sqlite":
		return OpenSQLite(cfg)
	case "postgres", "postgresql", "pg":
		return OpenPostgres(cfg)
	case "mysql", "mariadb":
		return OpenMariaDB(cfg)
	default:
		return nil, fmt.Errorf("unknown driver: %s", cfg.Driver)
	}
}

// runGenericQuery dipakai ketiga driver database/sql.
func runGenericQuery(ctx context.Context, sqldb *sql.DB, q string) (Result, error) {
	rows, err := sqldb.QueryContext(ctx, q)
	if err != nil {
		return Result{RowsAffected: -1}, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return Result{RowsAffected: -1}, err
	}

	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	var out Result
	out.Columns = cols
	out.RowsAffected = -1
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return out, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			row[i] = fmt.Sprintf("%v", v)
			if v == nil {
				row[i] = "NULL"
			} else if b, ok := v.([]byte); ok {
				row[i] = string(b)
			}
		}
		out.Rows = append(out.Rows, row)
	}
	return out, rows.Err()
}
