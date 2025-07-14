package service

import (
	"ms-notification/internal/config"
	"ms-notification/internal/domain"
	"ms-notification/pkg/email"
)

//go:generate mockgen -source=service.go -destination=mocks/mock_service.go

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

type SubscriptionEmails interface {
	SendConfirmationEmail(ConfirmationEmailInput) error
}

type WeatherEmails interface {
	SendWeatherForecastDailyEmail(WeatherForecastEmailInput[*domain.DayWeatherResponse]) error
	SendWeatherForecastHourlyEmail(WeatherForecastEmailInput[*domain.WeatherResponse]) error
}

type Emails interface {
	SubscriptionEmails
	WeatherEmails
}

type Deps struct {
	EmailSender email.Sender
	EmailConfig config.EmailConfig
	HTTPConfig  config.HTTPConfig
}

type Services struct {
	Emails Emails
}

func NewServices(deps Deps) *Services {
	emailsService := NewEmailsService(deps.EmailSender, deps.EmailConfig, deps.HTTPConfig)
	return &Services{
		Emails: emailsService,
	}
}
