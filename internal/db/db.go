package db

import (
	"weather_forecast_sub/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func ConnectDB(dbCfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dbCfg.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database connection")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	return db, nil
}
