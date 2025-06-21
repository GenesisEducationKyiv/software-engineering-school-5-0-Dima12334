package email_test

import (
	"errors"
	"os"
	"testing"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/service"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/testutils"

	"github.com/stretchr/testify/assert"
)

func TestGenerateBodyFromHTML(t *testing.T) {
	t.Run(
		"Generate HTML body for Subscription confirmation",
		testGenerateBodyFromHTMLSubscriptionConfirmation,
	)
	t.Run("Generate HTML body for Weather hourly email", testGenerateBodyFromHTMLWeatherHourly)
	t.Run("Generate HTML body for Weather daily email", testGenerateBodyFromHTMLWeatherDaily)
	t.Run("Template file does not exist", testGenerateBodyFromHTMLInvalidTemplateFile)
	t.Run("Template execution error", testGenerateBodyFromHTMLTemplateExecutionError)
}

func testGenerateBodyFromHTMLSubscriptionConfirmation(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	cfg := testutils.SetupTestConfig(t)

	input := &email.SendEmailInput{
		To:      "test@example.com",
		Subject: "Test Email",
	}
	templateData := service.WeatherForecastHourlyEmailTemplateInput{
		UnsubscribeLink: "https://example.com/api/unsubscribe",
		City:            "London",
		Weather: domain.WeatherResponse{
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
	t.Parallel()

	cfg := testutils.SetupTestConfig(t)

	input := &email.SendEmailInput{
		To:      "test@example.com",
		Subject: "Daily Email",
	}
	templateData := service.WeatherForecastDailyEmailTemplateInput{
		UnsubscribeLink: "https://example.com/api/unsubscribe",
		City:            "London",
		Weather: domain.DayWeatherResponse{
			SevenAM: domain.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy",
			},
			TenAM: domain.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			OnePM: domain.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			FourPM: domain.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			SevenPM: domain.WeatherResponse{
				Temperature: 18.0,
				Humidity:    70,
				Description: "Cloudy"},
			TenPM: domain.WeatherResponse{
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

func testGenerateBodyFromHTMLInvalidTemplateFile(t *testing.T) {
	t.Parallel()

	input := &email.SendEmailInput{
		To:      "test@example.com",
		Subject: "Invalid Template",
	}

	// Non-existent file
	err := input.GenerateBodyFromHTML("non_existent_template.html", nil)

	assert.Error(t, err)
	assert.Empty(t, input.Body)
}

func testGenerateBodyFromHTMLTemplateExecutionError(t *testing.T) {
	t.Parallel()

	// Create a temporary file with an invalid template (refers to a missing field)
	tmpFile, err := os.CreateTemp("", "bad_template_*.html")
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("failed to remove temp file %s: %v", name, err)
		}
	}(tmpFile.Name())

	// Write template with invalid field (e.g., {{.MissingField}})
	content := `<!DOCTYPE html><html><body>{{.MissingField}}</body></html>`
	_, err = tmpFile.Write([]byte(content))
	assert.NoError(t, err)
	err = tmpFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	input := &email.SendEmailInput{
		To:      "test@example.com",
		Subject: "Bad Execution",
	}

	// Pass data that doesn't have `.MissingField`
	err = input.GenerateBodyFromHTML(tmpFile.Name(), struct{ ValidField string }{ValidField: "hello"})

	assert.Error(t, err)
	assert.Empty(t, input.Body)
}

func TestSendEmailInput_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   email.SendEmailInput
		wantErr error
	}{
		{
			name:    "All fields valid",
			input:   email.SendEmailInput{To: "user@example.com", Subject: "Hello", Body: "Body content"},
			wantErr: nil,
		},
		{
			name:    "Missing To",
			input:   email.SendEmailInput{Subject: "Hello", Body: "Body content"},
			wantErr: errors.New("email 'To' field is required"),
		},
		{
			name:    "Missing Subject",
			input:   email.SendEmailInput{To: "user@example.com", Body: "Body content"},
			wantErr: errors.New("email 'Subject' field is required"),
		},
		{
			name:    "Missing Body",
			input:   email.SendEmailInput{To: "user@example.com", Subject: "Hello"},
			wantErr: errors.New("email 'Body' field is required"),
		},
		{
			name:    "Whitespace To field",
			input:   email.SendEmailInput{To: "   ", Subject: "Hello", Body: "Body content"},
			wantErr: errors.New("email 'To' field is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
