package testutils

import (
	"testing"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/migrations"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	cfg, err := config.Init(config.ConfigsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	err = migrations.ApplyMigrations(cfg.DB.DSN, cfg.DB.MigrationsPath, "up")
	if err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	db, err := sqlx.Open("postgres", cfg.DB.DSN)
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		_ = migrations.ApplyMigrations(cfg.DB.DSN, cfg.DB.MigrationsPath, "down")
	})

	return db
}
