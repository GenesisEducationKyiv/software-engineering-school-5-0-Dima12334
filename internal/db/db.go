package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"weather_forecast_sub/internal/config"
)

func ConnectDB(dbCfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dbCfg.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "description")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "description")
	}

	return db, nil
}
