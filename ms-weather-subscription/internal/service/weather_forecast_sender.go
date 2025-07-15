package service

import (
	"common/logger"
	"context"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/internal/domain"
	"ms-weather-subscription/pkg/clients"
	"time"
)

type (
	WeatherFetcherFunc[T clients.WeatherResponseType] func(ctx context.Context, city string) (T, error)
	EmailSenderFunc[T clients.WeatherResponseType]    func(inp clients.WeatherForecastEmailInput[T]) error
)

type sendWeatherForecastInput[T clients.WeatherResponseType] struct {
	ctx        context.Context
	repo       SubscriptionSenderRepository
	frequency  string
	dateFormat string
	getWeather WeatherFetcherFunc[T]
	sendEmail  EmailSenderFunc[T]
	baseURL    string
}

type SubscriptionSenderRepository interface {
	GetConfirmedByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error)
}

type WeatherForecastSenderService struct {
	httpConfig             config.HTTPConfig
	notificationClient     clients.WeatherNotificationSender
	weatherService         Weather
	subscriptionSenderRepo SubscriptionSenderRepository
}

func NewWeatherForecastSenderService(
	httpConfig config.HTTPConfig,
	weatherService Weather,
	subscriptionSenderRepo SubscriptionSenderRepository,
	notificationClient clients.WeatherNotificationSender,
) *WeatherForecastSenderService {
	return &WeatherForecastSenderService{
		httpConfig:             httpConfig,
		notificationClient:     notificationClient,
		weatherService:         weatherService,
		subscriptionSenderRepo: subscriptionSenderRepo,
	}
}

func (s *WeatherForecastSenderService) SendDailyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast(sendWeatherForecastInput[*domain.DayWeatherResponse]{
		ctx:        ctx,
		repo:       s.subscriptionSenderRepo,
		frequency:  domain.DailyWeatherEmailFrequency,
		dateFormat: time.DateOnly,
		getWeather: s.weatherService.GetDayWeather,
		sendEmail:  s.notificationClient.SendWeatherForecastDailyEmail,
		baseURL:    s.httpConfig.BaseURL,
	})
}

func (s *WeatherForecastSenderService) SendHourlyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast(sendWeatherForecastInput[*domain.WeatherResponse]{
		ctx:        ctx,
		repo:       s.subscriptionSenderRepo,
		frequency:  domain.HourlyWeatherEmailFrequency,
		dateFormat: time.DateTime,
		getWeather: s.weatherService.GetCurrentWeather,
		sendEmail:  s.notificationClient.SendWeatherForecastHourlyEmail,
		baseURL:    s.httpConfig.BaseURL,
	})
}

func sendWeatherForecast[T clients.WeatherResponseType](inp sendWeatherForecastInput[T]) error {
	subs, err := inp.repo.GetConfirmedByFrequency(inp.ctx, inp.frequency)
	if err != nil {
		logger.Errorf("failed to get subscriptions (%s): %s", inp.frequency, err.Error())
		return err
	}

	cityToSubscriptions := make(map[string][]domain.Subscription)
	for _, sub := range subs {
		cityToSubscriptions[sub.City] = append(cityToSubscriptions[sub.City], sub)
	}

	for city, subscriptions := range cityToSubscriptions {
		weatherData, err := inp.getWeather(inp.ctx, city)
		if err != nil {
			logger.Errorf("failed to get weather (%s) for city %s: %s", inp.frequency, city, err.Error())
			continue
		}

		for _, subscription := range subscriptions {
			emailInput := clients.WeatherForecastEmailInput[T]{
				Subscription:    subscription,
				Weather:         weatherData,
				Date:            time.Now().Format(inp.dateFormat),
				UnsubscribeLink: subscription.CreateUnsubscribeLink(inp.baseURL),
			}

			if err := inp.sendEmail(emailInput); err != nil {
				logger.Errorf(
					"failed to send email weather (%s) to %s: %s",
					inp.frequency,
					subscription.Email,
					err.Error(),
				)
				continue
			}
		}
	}

	return nil
}
