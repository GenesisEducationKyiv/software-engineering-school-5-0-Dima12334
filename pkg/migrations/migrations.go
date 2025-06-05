package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func ApplyMigrations(dsn, migrationsPath string, direction string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func(db *sql.DB) {
		if closeErr := db.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close db connection: %w", err, closeErr)
			} else {
				err = fmt.Errorf("failed to close db connection: %w", closeErr)
			}
		}
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	switch direction {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations up: %w", err)
		}
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations down: %w", err)
		}
	default:
		return fmt.Errorf("unknown migration direction: %s", direction)
	}

	return nil
}
