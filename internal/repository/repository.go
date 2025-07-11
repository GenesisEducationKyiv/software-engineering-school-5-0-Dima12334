package repository

import (
	"context"
	"time"
	"weather_forecast_sub/internal/domain"

	"github.com/jmoiron/sqlx"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription domain.Subscription) error
	GetByToken(ctx context.Context, token string) (domain.Subscription, error)
	Confirm(ctx context.Context, token string) error
	SetLastSentAt(ctx context.Context, lastSentAt time.Time, tokens []string) error
	Delete(ctx context.Context, token string) error
	GetConfirmedByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error)
}

type Repositories struct {
	Subscription SubscriptionRepository
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Subscription: NewSubscriptionRepo(db),
	}
}
