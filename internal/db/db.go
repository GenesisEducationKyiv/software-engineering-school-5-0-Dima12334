package db

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	driver      = "postgres"
	pingTimeout = 5 * time.Second
)

func NewDBConnection(dsn string) (*sqlx.DB, error) {
	return sqlx.Open(driver, dsn)
}

func ValidateDBConnection(db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	return errors.Wrap(db.PingContext(ctx), "ping wasn't successful")
}
