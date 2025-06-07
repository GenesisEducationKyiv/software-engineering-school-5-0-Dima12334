package tests

import (
	"log"
	"os"
	"testing"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/migrations"

	"github.com/jmoiron/sqlx"
)

const (
	configsDir = "../configs"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	cfg, err := config.Init(configsDir, config.TestEnvironment)
	if err != nil {
		log.Fatalf("failed to init configs: %v", err.Error())
	}

	err = migrations.ApplyMigrations(cfg.TestDB.DSN, cfg.TestDB.MigrationsPath, "up")
	if err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	testDB, err = sqlx.Open("postgres", cfg.TestDB.DSN)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	code := m.Run()

	// Cleanup
	if err := migrations.ApplyMigrations(cfg.TestDB.DSN, cfg.TestDB.MigrationsPath, "down"); err != nil {
		log.Fatalf("failed to rollback migrations: %v", err)
	}
	if err := testDB.Close(); err != nil {
		log.Fatalf("failed to close test db connection: %v", err)
	}

	os.Exit(code)
}
