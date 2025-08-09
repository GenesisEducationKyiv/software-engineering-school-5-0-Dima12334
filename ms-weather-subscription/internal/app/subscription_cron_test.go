package app_test

import (
	"context"
	"errors"
	"ms-weather-subscription/internal/domain"
	"ms-weather-subscription/pkg/publisher"
	"ms-weather-subscription/testutils"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"ms-weather-subscription/internal/repository"
	mockRepository "ms-weather-subscription/internal/repository/mocks"
	"ms-weather-subscription/internal/service"
	mockService "ms-weather-subscription/internal/service/mocks"
	mockPublisher "ms-weather-subscription/pkg/publisher/mocks"
)

func TestSubscriptionCron(t *testing.T) {
	t.Run("Send daily weather forecast success", testSendDailyWeatherForecastSuccess)
	t.Run(
		"Send daily weather forecast success with partial failure",
		testSendDailyWeatherForecastSuccessWithPartialFailure,
	)
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
	MockEmailPublisher           *mockPublisher.MockEmailPublisher
	CleanupFunc                  func()
}

func setupCronTestEnvironment(t *testing.T, ctrl *gomock.Controller) cronTestEnv {
	cfg := testutils.SetupTestConfig(t)
	testDB := testutils.SetupTestDB(t)

	subscriptionRepo := repository.NewSubscriptionRepo(testDB)
	mockWeatherService := mockService.NewMockWeather(ctrl)
	mockEmailPublisher := mockPublisher.NewMockEmailPublisher(ctrl)

	s := service.NewWeatherForecastSenderService(
		cfg.HTTP,
		mockWeatherService,
		subscriptionRepo,
		mockEmailPublisher,
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
		MockEmailPublisher:           mockEmailPublisher,
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

	testSettings.MockEmailPublisher.
		EXPECT().
		Publish(publisher.EmailDailyForecastQueue, gomock.Any()).
		Return(nil)

	// Execute
	err = testSettings.WeatherForecastSenderService.SendDailyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)
}

func testSendDailyWeatherForecastSuccessWithPartialFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupCronTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Insert 3 subscriptions
	_, err := testSettings.TestDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES 
            ('user1@example.com', 'Kyiv', 'daily', 'token1', true, NOW()),
            ('user2@example.com', 'Kyiv', 'daily', 'token2', true, NOW()),
            ('user3@example.com', 'Kyiv', 'daily', 'token3', true, NOW())
    `)
	assert.NoError(t, err)

	// Expect weather service once (shared city — Kyiv)
	testSettings.MockWeatherService.EXPECT().
		GetDayWeather(gomock.Any(), "Kyiv").
		Return(&domain.DayWeatherResponse{
			SevenAM: domain.WeatherResponse{Temperature: 20, Humidity: 60, Description: "Sunny"},
			TenAM:   domain.WeatherResponse{Temperature: 22, Humidity: 55, Description: "Sunny"},
		}, nil)

	// Expectations: 1st and 3rd succeed, 2nd fails
	testSettings.MockEmailPublisher.EXPECT().
		Publish(publisher.EmailDailyForecastQueue, gomock.Any()).
		DoAndReturn(
			func(queueName string, cmd domain.WeatherForecastEmailInput[*domain.DayWeatherResponse]) error {
				if cmd.Subscription.Email == "user2@example.com" {
					return errors.New("smtp failure")
				}
				return nil
			},
		).Times(3)

	// Execute
	err = testSettings.WeatherForecastSenderService.SendDailyWeatherForecast(context.Background())

	// Assert: the method still completes without global failure
	assert.NoError(t, err)
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

	cfg := testutils.SetupTestConfig(t)
	s := service.NewWeatherForecastSenderService(
		cfg.HTTP,
		mockService.NewMockWeather(ctrl),
		mockRepo,
		mockPublisher.NewMockEmailPublisher(ctrl),
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

	testSettings.MockEmailPublisher.EXPECT().
		Publish(publisher.EmailHourlyForecastQueue, gomock.Any()).
		Return(nil)

	// Execute
	err = testSettings.WeatherForecastSenderService.SendHourlyWeatherForecast(context.Background())

	// Verify
	assert.NoError(t, err)
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

	cfg := testutils.SetupTestConfig(t)
	s := service.NewWeatherForecastSenderService(
		cfg.HTTP,
		mockService.NewMockWeather(ctrl),
		mockRepo,
		mockPublisher.NewMockEmailPublisher(ctrl),
	)

	// Execute
	err := s.SendHourlyWeatherForecast(context.Background())

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
