package app_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/testutils"

	"github.com/jmoiron/sqlx"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"weather_forecast_sub/internal/repository"
	mockRepository "weather_forecast_sub/internal/repository/mocks"
	"weather_forecast_sub/internal/service"
	mockService "weather_forecast_sub/internal/service/mocks"
	mockSender "weather_forecast_sub/pkg/email/mocks"
)

func TestSubscriptionCron(t *testing.T) {
	t.Run("Send daily weather forecast success", testSendDailyWeatherForecastSuccess)
	t.Run("Send daily weather forecast no subscriptions", testSendDailyWeatherForecastNoSubs)
	t.Run("Send daily weather forecast repo error", testSendDailyWeatherForecastRepoError)
	t.Run("Send hourly weather forecast success", testSendHourlyWeatherForecastSuccess)
	t.Run("Send hourly weather forecast no subscriptions", testSendHourlyWeatherForecastNoSubs)
	t.Run("Send hourly weather forecast repo error", testSendHourlyWeatherForecastRepoError)
}

type cronTestEnv struct {
	TestDB                       *sqlx.DB
	WeatherForecastSenderService *service.WeatherForecastSenderService
	MockWeatherService           *mockService.MockWeather
	MockEmailSender              *mockSender.MockSender
	CleanupFunc                  func()
}

func setupCronTestEnvironment(t *testing.T, ctrl *gomock.Controller) cronTestEnv {
	cfg := testutils.SetupTestConfig(t)
	testDB := testutils.SetupTestDB(t)

	subscriptionRepo := repository.NewSubscriptionRepo(testDB)
	mockEmailSender := mockSender.NewMockSender(ctrl)
	mockWeatherService := mockService.NewMockWeather(ctrl)
	emailsService := service.NewEmailsService(mockEmailSender, cfg.Email, cfg.HTTP)

	s := service.NewWeatherForecastSenderService(
		emailsService,
		mockWeatherService,
		subscriptionRepo,
	)

	cleanupFunc := func() {
		_, err := testDB.Exec(`DELETE FROM subscriptions;`)
		if err != nil {
			t.Fatalf("cleanup failed: could not delete subscriptions data: %v", err)
		}
	}

	return cronTestEnv{
		TestDB:                       testDB,
		WeatherForecastSenderService: s,
		MockWeatherService:           mockWeatherService,
		MockEmailSender:              mockEmailSender,
		CleanupFunc:                  cleanupFunc,
	}
}

func testSendDailyWeatherForecastSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupCronTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Insert test subscription
	_, err := testSettings.TestDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('daily@example.com', 'Kyiv', 'daily', 'token1', true, NOW())
    `)
	assert.NoError(t, err)

	// Mock expectations
	testSettings.MockWeatherService.EXPECT().
		GetDayWeather(context.Background(), "Kyiv").
		Return(&domain.DayWeatherResponse{
			SevenAM: domain.WeatherResponse{Temperature: 20, Humidity: 60, Description: "Sunny"},
			TenAM:   domain.WeatherResponse{Temperature: 22, Humidity: 55, Description: "Sunny"},
			// ... other times
		}, nil)

	testSettings.MockEmailSender.EXPECT().
		Send(gomock.Any()).
		Return(nil)

	var lastSentAt *time.Time
	err = testSettings.TestDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'daily@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.Nil(t, lastSentAt)

	// Execute
	err = testSettings.WeatherForecastSenderService.SendDailyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)

	// Check last_sent_at was updated
	err = testSettings.TestDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'daily@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.NotNil(t, lastSentAt)
}

func testSendDailyWeatherForecastNoSubs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupCronTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Execute
	err := testSettings.WeatherForecastSenderService.SendDailyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)
}

func testSendDailyWeatherForecastRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock repo that returns error
	mockRepo := mockRepository.NewMockSubscriptionRepository(ctrl)
	mockRepo.EXPECT().GetConfirmedByFrequency(context.Background(), "daily").Return(
		nil, errors.New("database error"),
	)

	s := service.NewWeatherForecastSenderService(
		mockService.NewMockWeatherEmails(ctrl),
		mockService.NewMockWeather(ctrl),
		mockRepo,
	)

	// Execute
	err := s.SendDailyWeatherForecast(context.Background())

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func testSendHourlyWeatherForecastSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupCronTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Insert test subscription
	_, err := testSettings.TestDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('hourly@example.com', 'Kyiv', 'hourly', 'token2', true, NOW())
    `)
	assert.NoError(t, err)

	// Mock expectations
	testSettings.MockWeatherService.EXPECT().
		GetCurrentWeather(context.Background(), "Kyiv").
		Return(&domain.WeatherResponse{
			Temperature: 21.5,
			Humidity:    58,
			Description: "Partly Cloudy",
		}, nil)

	testSettings.MockEmailSender.EXPECT().
		Send(gomock.Any()).
		Return(nil)

	var lastSentAt *time.Time
	err = testSettings.TestDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'hourly@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.Nil(t, lastSentAt)

	// Execute
	err = testSettings.WeatherForecastSenderService.SendHourlyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)

	// Check last_sent_at was updated
	err = testSettings.TestDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'hourly@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.NotNil(t, lastSentAt)
}

func testSendHourlyWeatherForecastNoSubs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupCronTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Execute
	err := testSettings.WeatherForecastSenderService.SendHourlyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)
}

func testSendHourlyWeatherForecastRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock repo that returns error
	mockRepo := mockRepository.NewMockSubscriptionRepository(ctrl)
	mockRepo.EXPECT().GetConfirmedByFrequency(context.Background(), "hourly").Return(
		nil, errors.New("database error"),
	)

	s := service.NewWeatherForecastSenderService(
		mockService.NewMockWeatherEmails(ctrl),
		mockService.NewMockWeather(ctrl),
		mockRepo,
	)

	// Execute
	err := s.SendHourlyWeatherForecast(context.Background())

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
