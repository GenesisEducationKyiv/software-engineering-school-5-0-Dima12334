//nolint:ireturn // `SetupMockDB` requires interface return (sqlmock -> Sqlmock)
package testutils

import (
	"context"
	"database/sql"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/pkg/migrations"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	dbRetryCount    = 2
	dbRetryInterval = 1 * time.Second
	pingTimeout     = 5 * time.Second
)

func waitForDB(t *testing.T, dsn string) {
	var lastErr error
	for i := 1; i <= dbRetryCount; i++ {
		db, err := sql.Open("postgres", dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
			err = db.PingContext(ctx)
			cancel()

			if err == nil {
				if err := db.Close(); err != nil {
					t.Fatal(err)
				}
				return // DB is ready
			}
		}
		lastErr = err
		t.Logf("waiting for DB to be ready (attempt %d/%d): %v", i, dbRetryCount, err)
		time.Sleep(dbRetryInterval)
	}
	t.Fatalf("database not ready after %d attempts: %v", dbRetryCount, lastErr)
}

func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	cfg, err := config.Init(config.ConfigsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	waitForDB(t, cfg.DB.DSN)

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

func SetupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	return sqlx.NewDb(db, "postgres"), mock
}
