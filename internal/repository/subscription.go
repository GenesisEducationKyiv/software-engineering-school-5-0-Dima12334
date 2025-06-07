package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"weather_forecast_sub/internal/domain"
	customErrors "weather_forecast_sub/pkg/errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, subscription domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (created_at, email, city, token, frequency, confirmed, last_sent_at) 
		values ($1, $2, $3, $4, $5, $6, $7);`
	_, err := r.db.ExecContext(
		ctx,
		query,
		subscription.CreatedAt,
		subscription.Email,
		subscription.City,
		subscription.Token,
		subscription.Frequency,
		subscription.Confirmed,
		subscription.LastSentAt,
	)
	if err != nil {
		if customErrors.IsDuplicateDBError(err) {
			return customErrors.ErrSubscriptionAlreadyExists
		}
	}
	return err
}

func (r *SubscriptionRepo) GetByToken(ctx context.Context, token string) (domain.Subscription, error) {
	var subscription domain.Subscription

	query := `
		SELECT     
		 id,
		created_at,
		email,
		city,
		token,
		frequency,
		confirmed,
		last_sent_at
		FROM subscriptions
		WHERE token = $1;`

	err := r.db.QueryRowxContext(ctx, query, token).StructScan(&subscription)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Subscription{}, customErrors.ErrSubscriptionNotFound
		}

		return domain.Subscription{}, err
	}

	return subscription, nil
}

func (r *SubscriptionRepo) Confirm(ctx context.Context, token string) error {
	query := "UPDATE subscriptions SET confirmed = true WHERE token = $1;"
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *SubscriptionRepo) SetLastSentAt(lastSentAt time.Time, tokens []string) error {
	query := "UPDATE subscriptions SET last_sent_at = $1 WHERE token = ANY($2);"
	_, err := r.db.Exec(query, lastSentAt, pq.Array(tokens))
	return err
}

func (r *SubscriptionRepo) Delete(ctx context.Context, token string) error {
	query := "DELETE FROM subscriptions WHERE token = $1;"
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *SubscriptionRepo) GetConfirmedByFrequency(frequency string) ([]domain.Subscription, error) {
	var subscriptions []domain.Subscription

	query := `
		SELECT     
		 id,
		created_at,
		email,
		city,
		token,
		frequency,
		confirmed,
		last_sent_at
		FROM subscriptions
		WHERE confirmed = true AND frequency = $1;`

	err := r.db.Select(&subscriptions, query, frequency)

	return subscriptions, err
}
