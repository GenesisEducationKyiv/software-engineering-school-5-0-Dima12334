package service

import (
	"context"
	"net/url"
	"time"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/pkg/hash"
	"weather_forecast_sub/pkg/logger"
)

const (
	DailyWeatherEmailFrequency  = "daily"
	HourlyWeatherEmailFrequency = "hourly"
)

type WeatherFetcherFunc[T WeatherResponseType] func(ctx context.Context, city string) (T, error)
type EmailSenderFunc[T WeatherResponseType] func(inp WeatherForecastEmailInput[T]) error

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

func (s *SubscriptionService) SendDailyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast[*clients.DayWeatherResponse](
		ctx,
		s,
		DailyWeatherEmailFrequency,
		time.DateOnly,
		s.weatherService.GetDayWeather,
		s.emailService.SendWeatherForecastDailyEmail,
	)
}

func (s *SubscriptionService) SendHourlyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast[*clients.WeatherResponse](
		ctx,
		s,
		HourlyWeatherEmailFrequency,
		time.DateTime,
		s.weatherService.GetCurrentWeather,
		s.emailService.SendWeatherForecastHourlyEmail,
	)
}

func sendWeatherForecast[T WeatherResponseType](
	ctx context.Context,
	s *SubscriptionService,
	frequency string,
	dateFormat string,
	getWeatherFunc WeatherFetcherFunc[T],
	sendEmailFunc EmailSenderFunc[T],
) error {
	subs, err := s.repo.GetConfirmedByFrequency(frequency)
	if err != nil {
		logger.Errorf("failed to get subscriptions (%s): %s", frequency, err.Error())
		return err
	}

	cityToSubscriptions := make(map[string][]domain.Subscription)
	for _, sub := range subs {
		cityToSubscriptions[sub.City] = append(cityToSubscriptions[sub.City], sub)
	}

	var subscriptionsToUpdate []string
	for city, subscriptions := range cityToSubscriptions {
		escapedCity := url.QueryEscape(city)
		weatherData, err := getWeatherFunc(ctx, escapedCity)
		if err != nil {
			logger.Errorf("failed to get weather (%s) for city %s: %s", frequency, city, err.Error())
			continue
		}

		for _, subscription := range subscriptions {
			inp := WeatherForecastEmailInput[T]{
				Subscription: subscription,
				Weather:      weatherData,
				Date:         time.Now().Format(dateFormat),
			}

			if err := sendEmailFunc(inp); err != nil {
				logger.Errorf(
					"failed to send email (%s) to %s: %s",
					frequency,
					subscription.Email,
					err.Error(),
				)
				continue
			}
			subscriptionsToUpdate = append(subscriptionsToUpdate, subscription.Token)
		}
	}

	if len(subscriptionsToUpdate) == 0 {
		logger.Warnf("no %s subscriptions to update", frequency)
		return nil
	}

	return s.repo.SetLastSentAt(time.Now(), subscriptionsToUpdate)
}
