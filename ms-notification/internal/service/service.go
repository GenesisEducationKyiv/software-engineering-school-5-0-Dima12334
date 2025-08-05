package service

import (
	"ms-notification/internal/config"
	"ms-notification/internal/domain"
	"ms-notification/pkg/email"
)

//go:generate mockgen -source=service.go -destination=mocks/mock_service.go

type Email interface {
	SendConfirmationEmail(domain.ConfirmationEmailInput) error
	SendWeatherForecastDailyEmail(domain.WeatherForecastEmailInput[*domain.DayWeather]) error
	SendWeatherForecastHourlyEmail(domain.WeatherForecastEmailInput[*domain.Weather]) error
}

type Deps struct {
	EmailSender email.Sender
	EmailConfig config.EmailConfig
}

type Services struct {
	Email Email
}

func NewServices(deps Deps) *Services {
	emailService := NewEmailService(deps.EmailSender, deps.EmailConfig)
	return &Services{
		Email: emailService,
	}
}
