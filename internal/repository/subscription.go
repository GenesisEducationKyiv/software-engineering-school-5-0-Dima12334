package repository

import (
	"context"
	"database/sql"
	"errors"
	"ms-weather-subscription/internal/domain"
	customErrors "ms-weather-subscription/pkg/errors"

	"github.com/jmoiron/sqlx"
)

type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, subscription domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (created_at, email, city, token, frequency, confirmed) 
		values ($1, $2, $3, $4, $5, $6);`
	_, err := r.db.ExecContext(
		ctx,
		query,
		subscription.CreatedAt,
		subscription.Email,
		subscription.City,
		subscription.Token,
		subscription.Frequency,
		subscription.Confirmed,
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
		confirmed
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

func (r *SubscriptionRepo) Delete(ctx context.Context, token string) error {
	query := "DELETE FROM subscriptions WHERE token = $1;"
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *SubscriptionRepo) GetConfirmedByFrequency(
	ctx context.Context, frequency string,
) ([]domain.Subscription, error) {
	var subscriptions []domain.Subscription

	query := `
		SELECT     
		id,
		created_at,
		email,
		city,
		token,
		frequency,
		confirmed
		FROM subscriptions
		WHERE confirmed = true AND frequency = $1;`

	err := r.db.SelectContext(ctx, &subscriptions, query, frequency)

	return subscriptions, err
}
