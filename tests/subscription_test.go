package tests

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/handlers"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/internal/service"
	mockService "weather_forecast_sub/internal/service/mocks"
	mockSender "weather_forecast_sub/pkg/email/mocks"
	"weather_forecast_sub/pkg/hash"
)

func TestSubscription(t *testing.T) {
	t.Run("Show subscribe page", testShowSubscribePageMocked)
	t.Run("Successful subscription", testSuccessfulSubscribe)
	t.Run("Invalid request body", testInvalidSubscribeRequestBody)
	t.Run("Duplicate subscription", testDuplicateSubscribe)
	t.Run("Unsubscribe success", testUnsubscribeSuccess)
	t.Run("Unsubscribe not found", testUnsubscribeNotFound)
	t.Run("Confirm success", testConfirmSuccess)
	t.Run("Confirm not found", testConfirmNotFound)
	t.Run("Confirm invalid token", testConfirmInvalidToken)
}

func setupTestEnvironment(t *testing.T, ctrl *gomock.Controller) (*gin.Engine, *mockSender.MockSender, func()) {
	repo := repository.NewSubscriptionRepo(testDB)
	hasher := hash.NewSHA256Hasher()
	cfg, err := config.Init(configsDir, testEnvironment)
	if err != nil {
		t.Fatalf("failed to init configs: %v", err.Error())
	}

	mockEmailSender := mockSender.NewMockSender(ctrl)
	emailsService := service.NewEmailsService(mockEmailSender, cfg.Email, cfg.HTTP)
	mockWeatherService := mockService.NewMockWeather(ctrl)

	subService := service.NewSubscriptionService(
		repo,
		hasher,
		mockEmailSender,
		cfg.Email,
		cfg.HTTP,
		emailsService,
		mockWeatherService,
	)
	services := &service.Services{Subscriptions: subService}

	handler := handlers.NewHandler(services)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/subscribe", handler.ShowSubscribePage)
	router.POST("/api/subscribe", handler.SubscribeEmail)
	router.GET("/api/confirm/:token", handler.ConfirmEmail)
	router.GET("/api/unsubscribe/:token", handler.UnsubscribeEmail)

	// Create router with mock template
	router.LoadHTMLGlob("../templates/**/*.html")

	cleanup := func() {
		testDB.Exec(`DELETE FROM subscriptions;`)
	}

	return router, mockEmailSender, cleanup
}

func testShowSubscribePageMocked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	handler, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/subscribe", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Subscribe to Weather updates")
}

func testSuccessfulSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	handler, mockEmailSender, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Mock expectations
	mockEmailSender.
		EXPECT().
		Send(gomock.Any()).
		Return(nil)

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBufferString(
		`{"email": "test@example.com", "city": "Kyiv", "frequency": "daily"}`,
	))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	// Check database
	var sub domain.Subscription
	err := testDB.QueryRowx(`
		SELECT *
		FROM subscriptions 
		WHERE email = $1
	`, "test@example.com").StructScan(&sub)
	assert.NoError(t, err, "should find record in database")
	assert.Equal(t, "test@example.com", sub.Email)
	assert.Equal(t, "Kyiv", sub.City)
	assert.Equal(t, "daily", sub.Frequency)
	assert.False(t, sub.Confirmed, "subscription should not be confirmed yet")
}

func testInvalidSubscribeRequestBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	handler, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBufferString(
		`{"email": "test@example.com", "city": "Kyiv", "frequency": "wrong_frequency"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// No database changes expected
	var count int
	err := testDB.QueryRowx(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE email = $1
	`, "test@example.com").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "no record should be created for invalid request")
}

func testDuplicateSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	handler, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Create existing subscription
	hasher := hash.NewSHA256Hasher()
	token := hasher.GenerateEmailHash("existing@example.com")
	_, err := testDB.Exec(`
		INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
		VALUES ('existing@example.com', 'Kyiv', 'daily', $1, false, NOW())
	`, token)
	assert.NoError(t, err)

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBufferString(
		`{"email": "existing@example.com", "city": "Kyiv", "frequency": "daily"}`,
	))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusConflict, w.Code)

	// Verify existing record wasn't modified
	var originalToken string
	err = testDB.QueryRowx(`
		SELECT token 
		FROM subscriptions 
		WHERE email = $1
	`, "existing@example.com").Scan(&originalToken)
	assert.NoError(t, err)
	assert.Equal(t, token, originalToken)
}

func testUnsubscribeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Insert subscription to be deleted
	hasher := hash.NewSHA256Hasher()
	token := hasher.GenerateEmailHash("unsubscribe@example.com")

	_, err := testDB.Exec(`
		INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
		VALUES ('unsubscribe@example.com', 'Kyiv', 'daily', $1, false, NOW())
	`, token)
	assert.NoError(t, err)

	var count int
	err = testDB.QueryRowx(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE token = $1
	`, token).Scan(&count)
	assert.Equal(t, 1, count)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/unsubscribe/"+token, nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's deleted
	err = testDB.QueryRowx(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE token = $1
	`, token).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func testUnsubscribeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	handler, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	hasher := hash.NewSHA256Hasher()
	token := hasher.GenerateEmailHash("non-existen-email@example.com")

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/unsubscribe/"+token, nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func testUnsubscribeInvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Insert subscription to be deleted
	hasher := hash.NewSHA256Hasher()
	token := hasher.GenerateEmailHash("unsubscribe@example.com")

	_, err := testDB.Exec(`
		INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
		VALUES ('unsubscribe@example.com', 'Kyiv', 'daily', $1, false, NOW())
	`, token)
	assert.NoError(t, err)

	var count int
	err = testDB.QueryRowx(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE token = $1
	`, token).Scan(&count)
	assert.Equal(t, 1, count)

	w := httptest.NewRecorder()
	// Add "bug" instead of valid token
	req := httptest.NewRequest("GET", "/api/unsubscribe/"+"bug", nil)

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify it's not deleted
	err = testDB.QueryRowx(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE token = $1
	`, token).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func testConfirmSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Insert unconfirmed subscription
	hasher := hash.NewSHA256Hasher()
	token := hasher.GenerateEmailHash("confirm@example.com")

	_, err := testDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('confirm@example.com', 'Kyiv', 'daily', $1, false, NOW())
    `, token)
	assert.NoError(t, err)

	// Verify initial state
	var confirmed bool
	err = testDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.False(t, confirmed, "subscription should be unconfirmed initially")

	// Execute confirmation
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/confirm/"+token, nil)
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's confirmed in DB
	err = testDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.True(t, confirmed, "subscription should be confirmed after request")
}

func testConfirmNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	router, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	hasher := hash.NewSHA256Hasher()
	token := hasher.GenerateEmailHash("non-existen-email@example.com")

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/confirm/"+token, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func testConfirmInvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router, _, cleanup := setupTestEnvironment(t, ctrl)
	defer cleanup()

	// Insert unconfirmed subscription
	hasher := hash.NewSHA256Hasher()
	token := hasher.GenerateEmailHash("confirm@example.com")

	_, err := testDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('confirm@example.com', 'Kyiv', 'daily', $1, false, NOW())
    `, token)
	assert.NoError(t, err)

	// Verify initial state
	var confirmed bool
	err = testDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.False(t, confirmed, "subscription should be unconfirmed initially")

	// Execute confirmation
	w := httptest.NewRecorder()
	// Add "bug" instead of valid token
	req := httptest.NewRequest("GET", "/api/confirm/"+"bug", nil)
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify it's not confirmed in DB
	err = testDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.False(t, confirmed, "subscription should not be confirmed after request")
}
