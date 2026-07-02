package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrationFiles embed.FS

const baselineVersion int64 = 1

func Run(db *sql.DB) error {
	goose.SetBaseFS(migrationFiles)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	var hasVersionTable bool
	err := db.QueryRow(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'goose_db_version')",
	).Scan(&hasVersionTable)
	if err != nil {
		return fmt.Errorf("check goose_db_version: %w", err)
	}

	if !hasVersionTable {
		var hasUsers bool
		err := db.QueryRow(
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')",
		).Scan(&hasUsers)
		if err != nil {
			return fmt.Errorf("check users table: %w", err)
		}

		if hasUsers {
			_, err := db.Exec(`CREATE TABLE IF NOT EXISTS goose_db_version (
				version_id BIGINT NOT NULL PRIMARY KEY,
				is_applied BOOLEAN NOT NULL,
				tstamp TIMESTAMPTZ DEFAULT NOW()
			)`)
			if err != nil {
				return fmt.Errorf("create version table: %w", err)
			}
			_, err = db.Exec(
				"INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, true) ON CONFLICT DO NOTHING",
				baselineVersion,
			)
			if err != nil {
				return fmt.Errorf("set baseline version %d: %w", baselineVersion, err)
			}
		}
	}

	return goose.Up(db, ".")
}
