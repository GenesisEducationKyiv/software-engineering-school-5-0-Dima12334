package service

import (
	"context"
	"net/url"
	"time"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/pkg/logger"
)

type CronJobsService struct {
	emailService     Emails
	weatherService   Weather
	subscriptionRepo repository.SubscriptionRepository
}

func NewCronJobsService(
	emailService Emails,
	weatherService Weather,
	subscriptionRepo repository.SubscriptionRepository,
) *CronJobsService {
	return &CronJobsService{
		emailService:     emailService,
		weatherService:   weatherService,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *CronJobsService) SendDailyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast[*clients.DayWeatherResponse](
		ctx,
		s.subscriptionRepo,
		domain.DailyWeatherEmailFrequency,
		time.DateOnly,
		s.weatherService.GetDayWeather,
		s.emailService.SendWeatherForecastDailyEmail,
	)
}

func (s *CronJobsService) SendHourlyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast[*clients.WeatherResponse](
		ctx,
		s.subscriptionRepo,
		domain.HourlyWeatherEmailFrequency,
		time.DateTime,
		s.weatherService.GetCurrentWeather,
		s.emailService.SendWeatherForecastHourlyEmail,
	)
}

func sendWeatherForecast[T WeatherResponseType](
	ctx context.Context,
	subscriptionRepo repository.SubscriptionRepository,
	frequency string,
	dateFormat string,
	getWeatherFunc WeatherFetcherFunc[T],
	sendEmailFunc EmailSenderFunc[T],
) error {
	subs, err := subscriptionRepo.GetConfirmedByFrequency(frequency)
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

	return subscriptionRepo.SetLastSentAt(time.Now(), subscriptionsToUpdate)
}
