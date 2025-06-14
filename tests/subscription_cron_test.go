package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/repository"
	mockRepository "weather_forecast_sub/internal/repository/mocks"
	"weather_forecast_sub/internal/service"
	mockService "weather_forecast_sub/internal/service/mocks"
	"weather_forecast_sub/pkg/clients"
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

type CronTestEnv struct {
	CronJobService     *service.CronJobsService
	MockWeatherService *mockService.MockWeather
	MockEmailSender    *mockSender.MockSender
	CleanupFunc        func()
}

func setupCronTestEnvironment(t *testing.T, ctrl *gomock.Controller) CronTestEnv {
	subscriptionRepo := repository.NewSubscriptionRepo(testDB)
	cfg, err := config.Init(configsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to init configs: %v", err.Error())
	}

	mockEmailSender := mockSender.NewMockSender(ctrl)
	mockWeatherService := mockService.NewMockWeather(ctrl)
	emailsService := service.NewEmailsService(mockEmailSender, cfg.Email, cfg.HTTP)

	s := service.NewCronJobsService(
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

	return CronTestEnv{
		CronJobService:     s,
		MockWeatherService: mockWeatherService,
		MockEmailSender:    mockEmailSender,
		CleanupFunc:        cleanupFunc,
	}
}

func testSendDailyWeatherForecastSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	cronTestEnv := setupCronTestEnvironment(t, ctrl)
	defer cronTestEnv.CleanupFunc()

	// Insert test subscription
	_, err := testDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('daily@example.com', 'Kyiv', 'daily', 'token1', true, NOW())
    `)
	assert.NoError(t, err)

	// Mock expectations
	cronTestEnv.MockWeatherService.EXPECT().
		GetDayWeather(context.Background(), "Kyiv").
		Return(&clients.DayWeatherResponse{
			SevenAM: clients.WeatherResponse{Temperature: 20, Humidity: 60, Description: "Sunny"},
			TenAM:   clients.WeatherResponse{Temperature: 22, Humidity: 55, Description: "Sunny"},
			// ... other times
		}, nil)

	cronTestEnv.MockEmailSender.EXPECT().
		Send(gomock.Any()).
		Return(nil)

	var lastSentAt *time.Time
	err = testDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'daily@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.Nil(t, lastSentAt)

	// Execute
	err = cronTestEnv.CronJobService.SendDailyWeatherForecast(context.Background())

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
	cronTestEnv := setupCronTestEnvironment(t, ctrl)
	defer cronTestEnv.CleanupFunc()

	// Execute
	err := cronTestEnv.CronJobService.SendDailyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)
}

func testSendDailyWeatherForecastRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock repo that returns error
	mockRepo := mockRepository.NewMockSubscriptionRepository(ctrl)
	mockRepo.EXPECT().GetConfirmedByFrequency("daily").Return(nil, errors.New("database error"))

	s := service.NewCronJobsService(
		mockService.NewMockEmails(ctrl),
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
	cronTestEnv := setupCronTestEnvironment(t, ctrl)
	defer cronTestEnv.CleanupFunc()

	// Insert test subscription
	_, err := testDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('hourly@example.com', 'Kyiv', 'hourly', 'token2', true, NOW())
    `)
	assert.NoError(t, err)

	// Mock expectations
	cronTestEnv.MockWeatherService.EXPECT().
		GetCurrentWeather(context.Background(), "Kyiv").
		Return(&clients.WeatherResponse{
			Temperature: 21.5,
			Humidity:    58,
			Description: "Partly Cloudy",
		}, nil)

	cronTestEnv.MockEmailSender.EXPECT().
		Send(gomock.Any()).
		Return(nil)

	var lastSentAt *time.Time
	err = testDB.QueryRowx(`
        SELECT last_sent_at FROM subscriptions WHERE email = 'hourly@example.com'
    `).Scan(&lastSentAt)
	assert.NoError(t, err)
	assert.Nil(t, lastSentAt)

	// Execute
	err = cronTestEnv.CronJobService.SendHourlyWeatherForecast(context.Background())

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
	cronTestEnv := setupCronTestEnvironment(t, ctrl)
	defer cronTestEnv.CleanupFunc()

	// Execute
	err := cronTestEnv.CronJobService.SendHourlyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)
}

func testSendHourlyWeatherForecastRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock repo that returns error
	mockRepo := mockRepository.NewMockSubscriptionRepository(ctrl)
	mockRepo.EXPECT().GetConfirmedByFrequency("hourly").Return(nil, errors.New("database error"))

	s := service.NewCronJobsService(
		mockService.NewMockEmails(ctrl),
		mockService.NewMockWeather(ctrl),
		mockRepo,
	)

	// Execute
	err := s.SendHourlyWeatherForecast(context.Background())

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
