package internal

import (
	"fmt"
	"keyclub-api/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// prepares the database and runs migrations
func LoadDatabase(dbConfig config.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dbConfig.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the database: %w", err)
	}

	migrations, migrationErr := migrate.New(
		"file://"+dbConfig.MigrationsPath,
		"sqlite://"+dbConfig.SQLitePath,
	)
	if migrationErr != nil {
		return nil, fmt.Errorf("Failed to initialize migrations: %w", migrationErr)
	}
	if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("Failed to run migrations: %w", err)
	}
	return db, nil
}
