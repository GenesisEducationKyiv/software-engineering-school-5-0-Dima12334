package service

import (
	"ms-notification/internal/config"
	"ms-notification/internal/domain"
	"ms-notification/pkg/email"
)

//go:generate mockgen -source=service.go -destination=mocks/mock_service.go

type Emails interface {
	SendConfirmationEmail(domain.ConfirmationEmailInput) error
	SendWeatherForecastDailyEmail(domain.WeatherForecastEmailInput[*domain.DayWeatherResponse]) error
	SendWeatherForecastHourlyEmail(domain.WeatherForecastEmailInput[*domain.WeatherResponse]) error
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
