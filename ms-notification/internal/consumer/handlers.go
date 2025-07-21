package consumer

import (
	"common/logger"
	"encoding/json"
	"ms-notification/internal/domain"
)

type confirmationEmailCommand struct {
	Email            string `json:"email"`
	ConfirmationLink string `json:"confirmation_link"`
}

type dailyForecastCommand struct {
	Subscription    domain.Subscription `json:"subscription"`
	Weather         domain.DayWeather   `json:"weather"`
	Date            string              `json:"date"`
	UnsubscribeLink string              `json:"unsubscribe_link"`
}

type hourlyForecastCommand struct {
	Subscription    domain.Subscription `json:"subscription"`
	Weather         domain.Weather      `json:"weather"`
	Date            string              `json:"date"`
	UnsubscribeLink string              `json:"unsubscribe_link"`
}

func (c *Consumer) handleConfirmationEmail(body []byte) {
	var cmd confirmationEmailCommand
	if err := json.Unmarshal(body, &cmd); err != nil {
		logger.Errorf("failed to decode confirmation email: %s", err)
		return
	}

	inp := domain.ConfirmationEmailInput(cmd)
	if err := c.emailsService.SendConfirmationEmail(inp); err != nil {
		logger.Errorf("email send error: %s", err)
	}
}

func (c *Consumer) handleDailyForecast(body []byte) {
	var cmd dailyForecastCommand
	if err := json.Unmarshal(body, &cmd); err != nil {
		logger.Errorf("failed to decode daily forecast: %s", err)
		return
	}

	inp := domain.WeatherForecastEmailInput[*domain.DayWeather]{
		Subscription:    cmd.Subscription,
		Weather:         &cmd.Weather,
		Date:            cmd.Date,
		UnsubscribeLink: cmd.UnsubscribeLink,
	}
	if err := c.emailsService.SendWeatherForecastDailyEmail(inp); err != nil {
		logger.Errorf("daily email send error: %s", err)
	}
}

func (c *Consumer) handleHourlyForecast(body []byte) {
	var cmd hourlyForecastCommand
	if err := json.Unmarshal(body, &cmd); err != nil {
		logger.Errorf("failed to decode hourly forecast: %s", err)
		return
	}

	inp := domain.WeatherForecastEmailInput[*domain.Weather]{
		Subscription:    cmd.Subscription,
		Weather:         &cmd.Weather,
		Date:            cmd.Date,
		UnsubscribeLink: cmd.UnsubscribeLink,
	}
	if err := c.emailsService.SendWeatherForecastHourlyEmail(inp); err != nil {
		logger.Errorf("hourly email send error: %s", err)
	}
}
