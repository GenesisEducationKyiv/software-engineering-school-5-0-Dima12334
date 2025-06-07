package tests

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/repository"
	mockRepository "weather_forecast_sub/internal/repository/mocks"
	"weather_forecast_sub/internal/service"
	mockService "weather_forecast_sub/internal/service/mocks"
	"weather_forecast_sub/pkg/clients"
	mockSender "weather_forecast_sub/pkg/email/mocks"
	"weather_forecast_sub/pkg/hash"
)

func TestSubscriptionCron(t *testing.T) {
	t.Run("Send daily weather forecast success", testSendDailyWeatherForecastSuccess)
	t.Run("Send daily weather forecast no subscriptions", testSendDailyWeatherForecastNoSubs)
	t.Run("Send daily weather forecast repo error", testSendDailyWeatherForecastRepoError)
	t.Run("Send hourly weather forecast success", testSendHourlyWeatherForecastSuccess)
	t.Run("Send hourly weather forecast no subscriptions", testSendHourlyWeatherForecastNoSubs)
	t.Run("Send hourly weather forecast repo error", testSendHourlyWeatherForecastRepoError)
}

func setupCronTestEnvironment(t *testing.T, ctrl *gomock.Controller) (*service.SubscriptionService, *mockService.MockWeather, *mockSender.MockSender, func()) {
	repo := repository.NewSubscriptionRepo(testDB)
	hasher := hash.NewSHA256Hasher()
	cfg, err := config.Init(configsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to init configs: %v", err.Error())
	}

	mockEmailSender := mockSender.NewMockSender(ctrl)
	mockWeatherService := mockService.NewMockWeather(ctrl)
	emailsService := service.NewEmailsService(mockEmailSender, cfg.Email, cfg.HTTP)

	subService := service.NewSubscriptionService(
		repo,
		hasher,
		mockEmailSender,
		cfg.Email,
		cfg.HTTP,
		emailsService,
		mockWeatherService,
	)

	cleanup := func() {
		testDB.Exec(`DELETE FROM subscriptions;`)
	}

	return subService, mockWeatherService, mockEmailSender, cleanup
}

func testSendDailyWeatherForecastSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	s, mockWeather, mockEmailSender, cleanup := setupCronTestEnvironment(t, ctrl)
	defer cleanup()

	// Insert test subscription
	_, err := testDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('daily@example.com', 'Kyiv', 'daily', 'token1', true, NOW())
    `)
	assert.NoError(t, err)

	// Mock expectations
	mockWeather.EXPECT().
		GetDayWeather("Kyiv").
		Return(&clients.DayWeatherResponse{
			SevenAM: clients.WeatherResponse{Temperature: 20, Humidity: 60, Description: "Sunny"},
			TenAM:   clients.WeatherResponse{Temperature: 22, Humidity: 55, Description: "Sunny"},
			// ... other times
		}, nil)

	mockEmailSender.EXPECT().
		Send(gomock.Any()).
		Return(nil)

	var lastSentAt *time.Time
	err = testDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'daily@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.Nil(t, lastSentAt)

	// Execute
	err = s.SendDailyWeatherForecast()

	// Verify
	assert.NoError(t, err)

	// Check last_sent_at was updated
	err = testDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'daily@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.NotNil(t, lastSentAt)
}

func testSendDailyWeatherForecastNoSubs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	s, _, _, cleanup := setupCronTestEnvironment(t, ctrl)
	defer cleanup()

	// Execute
	err := s.SendDailyWeatherForecast()

	// Verify
	assert.NoError(t, err)
}

func testSendDailyWeatherForecastRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg, err := config.Init(configsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to init configs: %v", err.Error())
	}

	// Setup mock repo that returns error
	mockRepo := mockRepository.NewMockSubscriptionRepository(ctrl)
	mockRepo.EXPECT().GetConfirmedByFrequency("daily").Return(nil, errors.New("database error"))

	s := service.NewSubscriptionService(
		mockRepo,
		hash.NewSHA256Hasher(),
		mockSender.NewMockSender(ctrl),
		cfg.Email,
		cfg.HTTP,
		mockService.NewMockEmails(ctrl),
		mockService.NewMockWeather(ctrl),
	)

	// Execute
	err = s.SendDailyWeatherForecast()

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func testSendHourlyWeatherForecastSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	s, mockWeather, mockEmailSender, cleanup := setupCronTestEnvironment(t, ctrl)
	defer cleanup()

	// Insert test subscription
	_, err := testDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('hourly@example.com', 'Kyiv', 'hourly', 'token2', true, NOW())
    `)
	assert.NoError(t, err)

	// Mock expectations
	mockWeather.EXPECT().
		GetCurrentWeather("Kyiv").
		Return(&clients.WeatherResponse{
			Temperature: 21.5,
			Humidity:    58,
			Description: "Partly Cloudy",
		}, nil)

	mockEmailSender.EXPECT().
		Send(gomock.Any()).
		Return(nil)

	var lastSentAt *time.Time
	err = testDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'hourly@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.Nil(t, lastSentAt)

	// Execute
	err = s.SendHourlyWeatherForecast()

	// Verify
	assert.NoError(t, err)

	// Check last_sent_at was updated
	err = testDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'hourly@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.NotNil(t, lastSentAt)
}

func testSendHourlyWeatherForecastNoSubs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	s, _, _, cleanup := setupCronTestEnvironment(t, ctrl)
	defer cleanup()

	// Execute
	err := s.SendHourlyWeatherForecast()

	// Verify
	assert.NoError(t, err)
}

func testSendHourlyWeatherForecastRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg, err := config.Init(configsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to init configs: %v", err.Error())
	}

	// Setup mock repo that returns error
	mockRepo := mockRepository.NewMockSubscriptionRepository(ctrl)
	mockRepo.EXPECT().GetConfirmedByFrequency("hourly").Return(nil, errors.New("database error"))

	s := service.NewSubscriptionService(
		mockRepo,
		hash.NewSHA256Hasher(),
		mockSender.NewMockSender(ctrl),
		cfg.Email,
		cfg.HTTP,
		mockService.NewMockEmails(ctrl),
		mockService.NewMockWeather(ctrl),
	)

	// Execute
	err = s.SendHourlyWeatherForecast()

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
