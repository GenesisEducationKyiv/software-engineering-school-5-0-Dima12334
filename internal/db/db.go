package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"weather_forecast_sub/internal/config"
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
