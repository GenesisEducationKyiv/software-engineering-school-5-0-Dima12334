package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDBConnection(dsn string) (*sqlx.DB, error) {
	return sqlx.Open("postgres", dsn)
}

func ValidateDBConnection(db *sqlx.DB) error {
	return db.Ping()
}
