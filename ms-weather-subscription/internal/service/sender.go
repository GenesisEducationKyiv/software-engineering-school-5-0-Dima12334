package service

import (
	"common/logger"
	"context"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/internal/domain"
	"ms-weather-subscription/pkg/publisher"
	"time"
)

type (
	WeatherFetcherFunc[T domain.WeatherResponseType] func(ctx context.Context, city string) (T, error)
	EmailSenderFunc[T domain.WeatherResponseType]    func(inp domain.WeatherForecastEmailInput[T]) error
)

type sendWeatherForecastInput[T domain.WeatherResponseType] struct {
	ctx            context.Context
	repo           SubscriptionSenderRepository
	emailPublisher publisher.EmailPublisher
	frequency      string
	dateFormat     string
	queue          string
	getWeather     WeatherFetcherFunc[T]
	baseURL        string
}

type SubscriptionSenderRepository interface {
	GetConfirmedByFrequency(ctx context.Context, frequency string) ([]domain.Subscription, error)
}

type WeatherForecastSenderService struct {
	httpConfig             config.HTTPConfig
	emailPublisher         publisher.EmailPublisher
	weatherService         Weather
	subscriptionSenderRepo SubscriptionSenderRepository
}

func NewWeatherForecastSenderService(
	httpConfig config.HTTPConfig,
	weatherService Weather,
	subscriptionSenderRepo SubscriptionSenderRepository,
	emailPublisher publisher.EmailPublisher,
) *WeatherForecastSenderService {
	return &WeatherForecastSenderService{
		httpConfig:             httpConfig,
		emailPublisher:         emailPublisher,
		weatherService:         weatherService,
		subscriptionSenderRepo: subscriptionSenderRepo,
	}
}

func (s *WeatherForecastSenderService) SendDailyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast(sendWeatherForecastInput[*domain.DayWeatherResponse]{
		ctx:            ctx,
		repo:           s.subscriptionSenderRepo,
		emailPublisher: s.emailPublisher,
		frequency:      domain.DailyWeatherEmailFrequency,
		dateFormat:     time.DateOnly,
		queue:          publisher.EmailDailyForecastQueue,
		getWeather:     s.weatherService.GetDayWeather,
		baseURL:        s.httpConfig.BaseURL,
	})
}

func (s *WeatherForecastSenderService) SendHourlyWeatherForecast(ctx context.Context) error {
	return sendWeatherForecast(sendWeatherForecastInput[*domain.WeatherResponse]{
		ctx:            ctx,
		repo:           s.subscriptionSenderRepo,
		emailPublisher: s.emailPublisher,
		frequency:      domain.HourlyWeatherEmailFrequency,
		dateFormat:     time.DateTime,
		queue:          publisher.EmailHourlyForecastQueue,
		getWeather:     s.weatherService.GetCurrentWeather,
		baseURL:        s.httpConfig.BaseURL,
	})
}

func sendWeatherForecast[T domain.WeatherResponseType](inp sendWeatherForecastInput[T]) error {
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
			emailInput := domain.WeatherForecastEmailInput[T]{
				Subscription:    subscription,
				Weather:         weatherData,
				Date:            time.Now().Format(inp.dateFormat),
				UnsubscribeLink: subscription.CreateUnsubscribeLink(inp.baseURL),
			}

			if err := inp.emailPublisher.Publish(inp.queue, emailInput); err != nil {
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
