package clients

import (
	"common/logger"
	"context"
	"ms-weather-subscription/internal/domain"
	pb "proto_stubs"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const contextTimeout = 5 * time.Second

//go:generate mockgen -source=notification.go -destination=mocks/mock_notification.go

type NotificationSender interface {
	SendConfirmationEmail(domain.ConfirmationEmailInput) error
	SendWeatherForecastDailyEmail(domain.WeatherForecastEmailInput[*domain.DayWeatherResponse]) error
	SendWeatherForecastHourlyEmail(domain.WeatherForecastEmailInput[*domain.WeatherResponse]) error
}

type NotificationClient struct {
	client pb.NotificationServiceClient
}

func NewNotificationClient(notificationServiceURL string) (*NotificationClient, error) {
	conn, err := grpc.NewClient(
		notificationServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewNotificationServiceClient(conn)

	return &NotificationClient{
		client: client,
	}, nil
}

func (n *NotificationClient) SendConfirmationEmail(inp domain.ConfirmationEmailInput) error {
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
	inp domain.WeatherForecastEmailInput[*domain.DayWeatherResponse],
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

	return nil
}

func (n *NotificationClient) SendWeatherForecastHourlyEmail(
	inp domain.WeatherForecastEmailInput[*domain.WeatherResponse],
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

	return nil
}
