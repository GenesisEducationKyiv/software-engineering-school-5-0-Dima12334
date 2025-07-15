package service

import (
	"context"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/internal/domain"
	"ms-weather-subscription/pkg/hash"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription domain.Subscription) error
	GetByToken(ctx context.Context, token string) (domain.Subscription, error)
	Confirm(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
}

type SubscriptionNotificationSender interface {
	SendConfirmationEmail(domain.ConfirmationEmailInput) error
}

type SubscriptionService struct {
	repo               SubscriptionRepository
	hasher             hash.SubscriptionHasher
	httpConfig         config.HTTPConfig
	notificationClient SubscriptionNotificationSender
}

func NewSubscriptionService(
	httpConfig config.HTTPConfig,
	repo SubscriptionRepository,
	hasher hash.SubscriptionHasher,
	notificationClient SubscriptionNotificationSender,
) *SubscriptionService {
	return &SubscriptionService{
		httpConfig:         httpConfig,
		repo:               repo,
		hasher:             hasher,
		notificationClient: notificationClient,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, inp domain.CreateSubscriptionInput) error {
	token := s.hasher.GenerateSubscriptionHash(inp.Email, inp.City, inp.Frequency)

	subscription := domain.NewSubscription(inp.Email, inp.City, inp.Frequency, token)
	err := s.repo.Create(ctx, subscription)

	if err != nil {
		return err
	}

	err = s.notificationClient.SendConfirmationEmail(
		domain.ConfirmationEmailInput{
			Email:            inp.Email,
			ConfirmationLink: subscription.CreateConfirmationLink(s.httpConfig.BaseURL),
		},
	)
	if err != nil {
		delErr := s.repo.Delete(ctx, token)
		if delErr != nil {
			return delErr
		}
		return err
	}

	return nil
}

func (s *SubscriptionService) Confirm(ctx context.Context, token string) error {
	_, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return err
	}

	return s.repo.Confirm(ctx, token)
}

func (s *SubscriptionService) Delete(ctx context.Context, token string) error {
	_, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, token)
}
