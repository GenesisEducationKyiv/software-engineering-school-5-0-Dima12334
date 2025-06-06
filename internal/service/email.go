package service

import (
	"fmt"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/pkg/logger"
)

type EmailService struct {
	sender      email.Sender
	emailConfig config.EmailConfig
	httpConfig  config.HTTPConfig
}

type confirmationEmailInput struct {
	ConfirmationLink string
}

type weatherForecastDailyEmailInput struct {
	UnsubscribeLink string
	City            string
	Weather         clients.DayWeatherResponse
	Date            string
}

type weatherForecastHourlyEmailInput struct {
	UnsubscribeLink string
	City            string
	Weather         clients.WeatherResponse
	Date            string
}

func NewEmailsService(
	sender email.Sender,
	emailConfig config.EmailConfig,
	httpConfig config.HTTPConfig,
) *EmailService {
	return &EmailService{
		sender:      sender,
		emailConfig: emailConfig,
		httpConfig:  httpConfig,
	}
}

func (s *EmailService) SendConfirmationEmail(inp ConfirmationEmailInput) error {
	subject := s.emailConfig.Subjects.Confirmation

	templateInput := confirmationEmailInput{
		ConfirmationLink: s.createConfirmationLink(inp.Token),
	}
	sendInput := email.SendEmailInput{Subject: subject, To: inp.Email}

	if err := sendInput.GenerateBodyFromHTML(
		s.emailConfig.Templates.Confirmation,
		templateInput,
	); err != nil {
		logger.Errorf("failed to generate confirmation email body: %s", err.Error())
		return err
	}

	return s.sender.Send(sendInput)
}

func (s *EmailService) createConfirmationLink(token string) string {
	return fmt.Sprintf("%s/api/confirm/%s", s.httpConfig.BaseURL, token)
}

func (s *EmailService) createUnsubscribeLink(token string) string {
	return fmt.Sprintf("%s/api/unsubscribe/%s", s.httpConfig.BaseURL, token)
}

func (s *EmailService) SendWeatherForecastDailyEmail(inp WeatherForecastDailyEmailInput) error {
	subject := fmt.Sprintf(s.emailConfig.Subjects.WeatherForecast, inp.Subscription.City)

	templateInput := weatherForecastDailyEmailInput{
		UnsubscribeLink: s.createUnsubscribeLink(inp.Subscription.Token),
		City:            inp.Subscription.City,
		Weather:         *inp.Weather,
		Date:            inp.Date,
	}
	sendInput := email.SendEmailInput{Subject: subject, To: inp.Subscription.Email}

	if err := sendInput.GenerateBodyFromHTML(
		s.emailConfig.Templates.WeatherForecastDaily, templateInput,
	); err != nil {
		logger.Errorf("failed to generate weather daily email body: %s", err.Error())
		return err
	}

	return s.sender.Send(sendInput)
}

func (s *EmailService) SendWeatherForecastHourlyEmail(inp WeatherForecastHourlyEmailInput) error {
	subject := fmt.Sprintf(s.emailConfig.Subjects.WeatherForecast, inp.Subscription.City)

	templateInput := weatherForecastHourlyEmailInput{
		UnsubscribeLink: s.createUnsubscribeLink(inp.Subscription.Token),
		City:            inp.Subscription.City,
		Weather:         *inp.Weather,
		Date:            inp.Date,
	}
	sendInput := email.SendEmailInput{Subject: subject, To: inp.Subscription.Email}

	if err := sendInput.GenerateBodyFromHTML(
		s.emailConfig.Templates.WeatherForecastHourly,
		templateInput,
	); err != nil {
		logger.Errorf("failed to generate weather hourly email body: %s", err.Error())
		return err
	}

	return s.sender.Send(sendInput)
}
