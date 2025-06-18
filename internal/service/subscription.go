package service

import (
	"context"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/pkg/hash"
)

type SubscriptionService struct {
	repo           repository.SubscriptionRepository
	hasher         hash.SubscriptionHasher
	emailSender    email.Sender
	emailConfig    config.EmailConfig
	httpConfig     config.HTTPConfig
	emailService   Emails
	weatherService Weather
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	hasher hash.SubscriptionHasher,
	emailSender email.Sender,
	emailConfig config.EmailConfig,
	httpConfig config.HTTPConfig,
	emailService Emails,
	weatherService Weather,
) *SubscriptionService {
	return &SubscriptionService{
		repo:           repo,
		hasher:         hasher,
		emailSender:    emailSender,
		emailConfig:    emailConfig,
		httpConfig:     httpConfig,
		emailService:   emailService,
		weatherService: weatherService,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, inp CreateSubscriptionInput) error {
	token := s.hasher.GenerateSubscriptionHash(inp.Email, inp.City, inp.Frequency)

	subscription := domain.NewSubscription(inp.Email, inp.City, inp.Frequency, token)
	err := s.repo.Create(ctx, subscription)

	if err != nil {
		return err
	}

	return s.emailService.SendConfirmationEmail(ConfirmationEmailInput{Email: inp.Email, Token: token})
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
