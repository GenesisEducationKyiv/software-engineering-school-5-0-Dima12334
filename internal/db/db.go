package db

import (
	"weather_forecast_sub/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ConnectDB(dbCfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dbCfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
