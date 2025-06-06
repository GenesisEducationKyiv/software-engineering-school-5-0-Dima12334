package service

import (
	"context"
	"net/url"
	"time"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/pkg/hash"
	"weather_forecast_sub/pkg/logger"
)

const (
	DailyFrequency  = "daily"
	HourlyFrequency = "hourly"
)

type SubscriptionService struct {
	repo           repository.SubscriptionRepository
	hasher         hash.EmailHasher
	emailSender    email.Sender
	emailConfig    config.EmailConfig
	httpConfig     config.HTTPConfig
	emailService   Emails
	weatherService Weather
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	hasher hash.EmailHasher,
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
	token := s.hasher.GenerateEmailHash(inp.Email)

	subscription := domain.Subscription{
		CreatedAt:  time.Now(),
		Email:      inp.Email,
		City:       inp.City,
		Frequency:  inp.Frequency,
		Token:      token,
		Confirmed:  false,
		LastSentAt: nil,
	}
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

func (s *SubscriptionService) SendDailyWeatherForecast() error {
	subs, err := s.repo.GetConfirmedByFrequency(DailyFrequency)
	if err != nil {
		logger.Errorf("failed to get daily subscriptions: %s", err.Error())
		return err
	}

	var cityToSubscriptionMap = make(map[string][]domain.Subscription)
	for _, sub := range subs {
		cityToSubscriptionMap[sub.City] = append(cityToSubscriptionMap[sub.City], sub)
	}

	var subscriptionsToUpdate []string

	for city, subscriptions := range cityToSubscriptionMap {
		escapedCity := url.QueryEscape(city)
		weather, err := s.weatherService.GetDayWeather(escapedCity)
		if err != nil {
			logger.Errorf("failed to get daily weather for city %s: %s", city, err.Error())
			continue
		}

		for _, subscription := range subscriptions {
			err = s.emailService.SendWeatherForecastDailyEmail(WeatherForecastDailyEmailInput{
				Subscription: subscription,
				Weather:      weather,
				Date:         time.Now().Format(time.DateOnly),
			})
			if err != nil {
				logger.Errorf("failed to send daily email to %s: %s", subscription.Email, err.Error())
				continue
			}
			subscriptionsToUpdate = append(subscriptionsToUpdate, subscription.Token)
		}
	}

	if len(subscriptionsToUpdate) == 0 {
		logger.Warn("no daily subscriptions to update")
		return nil
	}
	return s.repo.SetLastSentAt(time.Now(), subscriptionsToUpdate)
}

func (s *SubscriptionService) SendHourlyWeatherForecast() error {
	subs, err := s.repo.GetConfirmedByFrequency(HourlyFrequency)
	if err != nil {
		logger.Errorf("failed to get hourly subscriptions: %s", err.Error())
		return err
	}

	var cityToSubscriptionMap = make(map[string][]domain.Subscription)
	for _, sub := range subs {
		cityToSubscriptionMap[sub.City] = append(cityToSubscriptionMap[sub.City], sub)
	}

	var subscriptionsToUpdate []string

	for city, subscriptions := range cityToSubscriptionMap {
		escapedCity := url.QueryEscape(city)
		weather, err := s.weatherService.GetCurrentWeather(escapedCity)
		if err != nil {
			logger.Errorf("failed to get hourly weather for city %s: %s", city, err.Error())
			continue
		}

		for _, subscription := range subscriptions {
			err = s.emailService.SendWeatherForecastHourlyEmail(WeatherForecastHourlyEmailInput{
				Subscription: subscription,
				Weather:      weather,
				Date:         time.Now().Format(time.DateTime),
			})
			if err != nil {
				logger.Errorf("failed to send hourly email to %s: %s", subscription.Email, err.Error())
				continue
			}
			subscriptionsToUpdate = append(subscriptionsToUpdate, subscription.Token)
		}
	}

	if len(subscriptionsToUpdate) == 0 {
		logger.Warn("no hourly subscriptions to update")
		return nil
	}
	return s.repo.SetLastSentAt(time.Now(), subscriptionsToUpdate)
}
