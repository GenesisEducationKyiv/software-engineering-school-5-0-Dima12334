package clients

import (
	"common/logger"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"ms-weather-subscription/internal/domain"
	pb "proto_stubs"
	"time"
)

const contextTimeout = 5 * time.Second

type WeatherResponseType interface {
	*domain.WeatherResponse | *domain.DayWeatherResponse
}

type ConfirmationEmailInput struct {
	Email            string
	ConfirmationLink string
}

type WeatherForecastEmailInput[T WeatherResponseType] struct {
	Subscription    domain.Subscription
	Weather         T
	Date            string
	UnsubscribeLink string
}

type SubscriptionNotificationSender interface {
	SendConfirmationEmail(ConfirmationEmailInput) error
}

type WeatherNotificationSender interface {
	SendWeatherForecastDailyEmail(WeatherForecastEmailInput[*domain.DayWeatherResponse]) error
	SendWeatherForecastHourlyEmail(WeatherForecastEmailInput[*domain.WeatherResponse]) error
}

type NotificationSender interface {
	SubscriptionNotificationSender
	WeatherNotificationSender
}

type NotificationClient struct {
	client pb.NotificationServiceClient
}

func NewNotificationClient(NotificationServiceURL string) (*NotificationClient, error) {
	conn, err := grpc.NewClient(
		NotificationServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewNotificationServiceClient(conn)

	return &NotificationClient{
		client: client,
	}, nil
}

func (n *NotificationClient) SendConfirmationEmail(inp ConfirmationEmailInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	status, err := n.client.SendConfirmationEmail(ctx, &pb.ConfirmationEmailRequest{
		Email:            inp.Email,
		ConfirmationLink: inp.ConfirmationLink,
	})

	if err != nil {
		return err
	}

	if !status.Success {
		logger.Warnf("failed to send confirmation email for %s", inp.Email)
	}

	return nil
}

func (n *NotificationClient) SendWeatherForecastDailyEmail(
	inp WeatherForecastEmailInput[*domain.DayWeatherResponse],
) error {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	mapToProtoWeather := func(w domain.WeatherResponse) *pb.Weather {
		return &pb.Weather{
			Temperature: w.Temperature,
			Humidity:    w.Humidity,
			Description: w.Description,
		}
	}

	status, err := n.client.SendDailyForecastEmail(ctx, &pb.DailyForecastEmailRequest{
		Subscription: &pb.Subscription{
			Email: inp.Subscription.Email,
			City:  inp.Subscription.City,
		},
		Weather: &pb.DayWeather{
			SevenAm: mapToProtoWeather(inp.Weather.SevenAM),
			TenAm:   mapToProtoWeather(inp.Weather.TenAM),
			OnePm:   mapToProtoWeather(inp.Weather.OnePM),
			FourPm:  mapToProtoWeather(inp.Weather.FourPM),
			SevenPm: mapToProtoWeather(inp.Weather.SevenPM),
			TenPm:   mapToProtoWeather(inp.Weather.TenPM),
		},
		Date:            inp.Date,
		UnsubscribeLink: inp.UnsubscribeLink,
	})

	if err != nil {
		return err
	}

	if !status.Success {
		logger.Warnf("failed to send daily weather forecast for %s", inp.Subscription.Email)
	}

	return err
}

func (n *NotificationClient) SendWeatherForecastHourlyEmail(
	inp WeatherForecastEmailInput[*domain.WeatherResponse],
) error {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	status, err := n.client.SendHourlyForecastEmail(ctx, &pb.HourlyForecastEmailRequest{
		Subscription: &pb.Subscription{
			Email: inp.Subscription.Email,
			City:  inp.Subscription.City,
		},
		Weather: &pb.Weather{
			Temperature: inp.Weather.Temperature,
			Humidity:    inp.Weather.Humidity,
			Description: inp.Weather.Description,
		},
		Date:            inp.Date,
		UnsubscribeLink: inp.UnsubscribeLink,
	})

	if err != nil {
		return err
	}

	if !status.Success {
		logger.Warnf("failed to send hourly weather forecast for %s", inp.Subscription.Email)
	}

	return err
}
