package service

import (
	"context"
	"ms-weather-subscription/internal/domain"
	"ms-weather-subscription/pkg/hash"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription domain.Subscription) error
	GetByToken(ctx context.Context, token string) (domain.Subscription, error)
	Confirm(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
}

type SubscriptionService struct {
	repo         SubscriptionRepository
	hasher       hash.SubscriptionHasher
	emailService SubscriptionEmails
}

func NewSubscriptionService(
	repo SubscriptionRepository, hasher hash.SubscriptionHasher, emailService SubscriptionEmails,
) *SubscriptionService {
	return &SubscriptionService{
		repo:         repo,
		hasher:       hasher,
		emailService: emailService,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, inp domain.CreateSubscriptionInput) error {
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
