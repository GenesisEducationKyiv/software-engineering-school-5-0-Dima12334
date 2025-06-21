package service

import (
	"fmt"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
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
	Weather         domain.DayWeatherResponse
	Date            string
}

type weatherForecastHourlyEmailInput struct {
	UnsubscribeLink string
	City            string
	Weather         domain.WeatherResponse
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

	if err := sendInput.Validate(); err != nil {
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

func sendWeatherForecastEmail(
	sender email.Sender,
	subscription domain.Subscription,
	subject string,
	templateName string,
	templateData any,
) error {
	sendInput := email.SendEmailInput{Subject: subject, To: subscription.Email}

	if err := sendInput.GenerateBodyFromHTML(templateName, templateData); err != nil {
		logger.Errorf("failed to generate weather email body (%s): %s", templateName, err.Error())
		return err
	}

	if err := sendInput.Validate(); err != nil {
		return err
	}

	return sender.Send(sendInput)
}

func (s *EmailService) SendWeatherForecastDailyEmail(
	inp WeatherForecastEmailInput[*domain.DayWeatherResponse],
) error {
	templateInput := weatherForecastDailyEmailInput{
		UnsubscribeLink: s.createUnsubscribeLink(inp.Subscription.Token),
		City:            inp.Subscription.City,
		Weather:         *inp.Weather,
		Date:            inp.Date,
	}

	subject := fmt.Sprintf(s.emailConfig.Subjects.WeatherForecast, inp.Subscription.City)

	return sendWeatherForecastEmail(
		s.sender,
		inp.Subscription,
		subject,
		s.emailConfig.Templates.WeatherForecastDaily,
		templateInput,
	)
}

func (s *EmailService) SendWeatherForecastHourlyEmail(
	inp WeatherForecastEmailInput[*domain.WeatherResponse],
) error {
	templateInput := weatherForecastHourlyEmailInput{
		UnsubscribeLink: s.createUnsubscribeLink(inp.Subscription.Token),
		City:            inp.Subscription.City,
		Weather:         *inp.Weather,
		Date:            inp.Date,
	}

	subject := fmt.Sprintf(s.emailConfig.Subjects.WeatherForecast, inp.Subscription.City)

	return sendWeatherForecastEmail(
		s.sender,
		inp.Subscription,
		subject,
		s.emailConfig.Templates.WeatherForecastHourly,
		templateInput,
	)
}
