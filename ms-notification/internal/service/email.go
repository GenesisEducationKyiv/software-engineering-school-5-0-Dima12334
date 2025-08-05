package service

import (
	"common/logger"
	"fmt"
	"ms-notification/internal/config"
	"ms-notification/internal/domain"
	"ms-notification/pkg/email"
)

type ConfirmationEmailTemplateInput struct {
	ConfirmationLink string
}

type WeatherForecastDailyEmailTemplateInput struct {
	UnsubscribeLink string
	City            string
	Weather         domain.DayWeather
	Date            string
}

type WeatherForecastHourlyEmailTemplateInput struct {
	UnsubscribeLink string
	City            string
	Weather         domain.Weather
	Date            string
}

type EmailService struct {
	sender      email.Sender
	emailConfig config.EmailConfig
}

func NewEmailService(
	sender email.Sender,
	emailConfig config.EmailConfig,
) *EmailService {
	return &EmailService{
		sender:      sender,
		emailConfig: emailConfig,
	}
}

func (s *EmailService) SendConfirmationEmail(inp domain.ConfirmationEmailInput) error {
	subject := s.emailConfig.Subjects.Confirmation

	templateInput := ConfirmationEmailTemplateInput{
		ConfirmationLink: inp.ConfirmationLink,
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
	inp domain.WeatherForecastEmailInput[*domain.DayWeather],
) error {
	templateInput := WeatherForecastDailyEmailTemplateInput{
		UnsubscribeLink: inp.UnsubscribeLink,
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
	inp domain.WeatherForecastEmailInput[*domain.Weather],
) error {
	templateInput := WeatherForecastHourlyEmailTemplateInput{
		UnsubscribeLink: inp.UnsubscribeLink,
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
