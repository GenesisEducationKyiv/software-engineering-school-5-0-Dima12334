package testutils

import (
	"testing"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/migrations"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

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
		if err := db.Close(); err != nil {
			t.Logf("failed to close DB: %v", err)
		}
		if err := migrations.ApplyMigrations(cfg.DB.DSN, cfg.DB.MigrationsPath, "down"); err != nil {
			t.Logf("failed to rollback migrations: %v", err)
		}
	})

	return db
}

//nolint:ireturn
func SetupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	return sqlx.NewDb(db, "postgres"), mock
}
