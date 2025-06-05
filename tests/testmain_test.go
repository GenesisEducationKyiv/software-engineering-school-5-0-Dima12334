package tests

import (
	"fmt"
	"log"
	"os"
	"testing"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/migrations"

	"github.com/jmoiron/sqlx"
)

const (
	testEnvironment = "test"
	configsDir      = "../configs"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	cfg, err := config.Init(configsDir, testEnvironment)
	if err != nil {
		log.Fatalf("failed to init configs: %v", err.Error())
	}

	err = migrations.ApplyMigrations(cfg.TestDB.DSN, cfg.TestDB.MigrationsPath, "up")
	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	testDB, err = sqlx.Open("postgres", cfg.TestDB.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer func(db *sqlx.DB) {
		if closeErr := db.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close test db connection: %w", err, closeErr)
			} else {
				err = fmt.Errorf("failed to close test db connection: %w", closeErr)
			}
		}
	}(testDB)

	code := m.Run()

	if err := migrations.ApplyMigrations(cfg.TestDB.DSN, cfg.TestDB.MigrationsPath, "down"); err != nil {
		log.Printf("Failed to rollback migrations: %v", err)
	}

	os.Exit(code)
}
