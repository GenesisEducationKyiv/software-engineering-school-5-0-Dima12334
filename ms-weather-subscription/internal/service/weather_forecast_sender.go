package service

import (
	"common/logger"
	"context"
	"ms-weather-subscription/internal/domain"
	"time"
)

type (
	WeatherFetcherFunc[T WeatherResponseType] func(ctx context.Context, city string) (T, error)
	EmailSenderFunc[T WeatherResponseType]    func(inp WeatherForecastEmailInput[T]) error
)

type SubscriptionSenderRepository interface {
	GetConfirmedByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error)
}

type WeatherForecastSenderService struct {
	emailService           WeatherEmails
	weatherService         Weather
	subscriptionSenderRepo SubscriptionSenderRepository
}

func NewWeatherForecastSenderService(
	emailService WeatherEmails,
	weatherService Weather,
	subscriptionSenderRepo SubscriptionSenderRepository,
) *WeatherForecastSenderService {
	return &WeatherForecastSenderService{
		emailService:           emailService,
		weatherService:         weatherService,
		subscriptionSenderRepo: subscriptionSenderRepo,
	}
}

func (s *WeatherForecastSenderService) SendDailyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast(
		ctx,
		s.subscriptionSenderRepo,
		domain.DailyWeatherEmailFrequency,
		time.DateOnly,
		s.weatherService.GetDayWeather,
		s.emailService.SendWeatherForecastDailyEmail,
	)
}

func (s *WeatherForecastSenderService) SendHourlyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast(
		ctx,
		s.subscriptionSenderRepo,
		domain.HourlyWeatherEmailFrequency,
		time.DateTime,
		s.weatherService.GetCurrentWeather,
		s.emailService.SendWeatherForecastHourlyEmail,
	)
}

func sendWeatherForecast[T WeatherResponseType](
	ctx context.Context,
	subscriptionSenderRepo SubscriptionSenderRepository,
	frequency string,
	dateFormat string,
	getWeatherFunc WeatherFetcherFunc[T],
	sendEmailFunc EmailSenderFunc[T],
) error {
	subs, err := subscriptionSenderRepo.GetConfirmedByFrequency(ctx, frequency)
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
		weatherData, err := getWeatherFunc(ctx, city)
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

	return nil
}
