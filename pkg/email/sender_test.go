package email_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"weather_forecast_sub/internal/service"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/testutils"
)

func TestGenerateBodyFromHTML(t *testing.T) {
	t.Run("Generate HTML body for Subscription confirmation", testGenerateBodyFromHTMLSubscriptionConfirmation)
	t.Run("Generate HTML body for Weather hourly email", testGenerateBodyFromHTMLWeatherHourly)
	t.Run("Generate HTML body for Weather daily email", testGenerateBodyFromHTMLWeatherDaily)
}

func testGenerateBodyFromHTMLSubscriptionConfirmation(t *testing.T) {
	cfg := testutils.SetupTestConfig(t)

	input := &email.SendEmailInput{
		To:      "test@example.com",
		Subject: "Test Email",
	}
	templateData := service.ConfirmationEmailTemplateInput{
		ConfirmationLink: "https://example.com/api/confirm",
	}

	err := input.GenerateBodyFromHTML(cfg.Email.Templates.Confirmation, templateData)

	assert.Nil(t, err)
	assert.Contains(t, input.Body, "https://example.com/api/confirm")
}

func testGenerateBodyFromHTMLWeatherHourly(t *testing.T) {
	cfg := testutils.SetupTestConfig(t)

	input := &email.SendEmailInput{
		To:      "test@example.com",
		Subject: "Test Email",
	}
	templateData := service.WeatherForecastHourlyEmailTemplateInput{
		UnsubscribeLink: "https://example.com/api/unsubscribe",
		City:            "London",
		Weather: clients.WeatherResponse{
			Temperature: 20.5,
			Humidity:    65,
			Description: "Sunny",
		},
		Date: "2025-01-01",
	}

	err := input.GenerateBodyFromHTML(cfg.Email.Templates.WeatherForecastHourly, templateData)

	assert.Nil(t, err)
	assert.Contains(t, input.Body, "https://example.com/api/unsubscribe")
	assert.Contains(t, input.Body, "2025-01-01")
}

func testGenerateBodyFromHTMLWeatherDaily(t *testing.T) {
	cfg := testutils.SetupTestConfig(t)

	input := &email.SendEmailInput{
		To:      "test@example.com",
		Subject: "Daily Email",
	}
	templateData := service.WeatherForecastDailyEmailTemplateInput{
		UnsubscribeLink: "https://example.com/api/unsubscribe",
		City:            "London",
		Weather: clients.DayWeatherResponse{
			SevenAM: clients.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy",
			},
			TenAM: clients.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			OnePM: clients.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			FourPM: clients.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			SevenPM: clients.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			TenPM: clients.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
		},
		Date: "2025-01-01",
	}

	err := input.GenerateBodyFromHTML(cfg.Email.Templates.WeatherForecastDaily, templateData)

	assert.Nil(t, err)
	assert.Contains(t, input.Body, "https://example.com/api/unsubscribe")
	assert.Contains(t, input.Body, "2025-01-01")
}

func TestSendEmailInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   email.SendEmailInput
		wantErr string
	}{
		{
			name:    "All fields valid",
			input:   email.SendEmailInput{To: "user@example.com", Subject: "Hello", Body: "Body content"},
			wantErr: "",
		},
		{
			name:    "Missing To",
			input:   email.SendEmailInput{Subject: "Hello", Body: "Body content"},
			wantErr: "email 'To' field is required",
		},
		{
			name:    "Missing Subject",
			input:   email.SendEmailInput{To: "user@example.com", Body: "Body content"},
			wantErr: "email 'Subject' field is required",
		},
		{
			name:    "Missing Body",
			input:   email.SendEmailInput{To: "user@example.com", Subject: "Hello"},
			wantErr: "email 'Body' field is required",
		},
		{
			name:    "Whitespace To field",
			input:   email.SendEmailInput{To: "   ", Subject: "Hello", Body: "Body content"},
			wantErr: "email 'To' field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
