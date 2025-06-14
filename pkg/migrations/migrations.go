package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func ApplyMigrations(dsn, migrationsPath, direction string) (err error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close db connection: %v", err, closeErr)
			} else {
				err = fmt.Errorf("failed to close db connection: %w", closeErr)
			}
		}
	}()

	migrator, err := createMigrator(db, migrationsPath)
	if err != nil {
		return err
	}

	return applyDirection(migrator, direction)
}

func createMigrator(db *sql.DB, migrationsPath string) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrations: %w", err)
	}

	return m, nil
}

func applyDirection(m *migrate.Migrate, direction string) error {
	switch direction {
	case "up":
		err := m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migration up failed: %w", err)
		}
	case "down":
		err := m.Down()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migration down failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown migration direction: %s", direction)
	}

	return nil
}
