package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/handlers"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/internal/service"
	mockSender "weather_forecast_sub/pkg/email/mocks"
	"weather_forecast_sub/pkg/hash"
	"weather_forecast_sub/testutils"

	"github.com/jmoiron/sqlx"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSubscription(t *testing.T) {
	t.Run("Show subscribe page", testShowSubscribePageMocked)
	t.Run("Successful subscription", testSuccessfulSubscribe)
	t.Run("Invalid request body", testInvalidSubscribeRequestBody)
	t.Run("Duplicate subscription", testDuplicateSubscribe)
	t.Run("Duplicate email subscription with different frequency", testDuplicateEmailSubscribe)
	t.Run("Unsubscribe success", testUnsubscribeSuccess)
	t.Run("Unsubscribe not found", testUnsubscribeNotFound)
	t.Run("Unsubscribe invalid token", testUnsubscribeInvalidToken)
	t.Run("Confirm success", testConfirmSuccess)
	t.Run("Confirm not found", testConfirmNotFound)
	t.Run("Confirm invalid token", testConfirmInvalidToken)
}

type subscriptionTestEnv struct {
	TestDB          *sqlx.DB
	Router          *gin.Engine
	MockEmailSender *mockSender.MockSender
	CleanupFunc     func()
}

func setupTestEnvironment(t *testing.T, ctrl *gomock.Controller) subscriptionTestEnv {
	cfg := testutils.SetupTestConfig(t)
	testDB := testutils.SetupTestDB(t)

	repo := repository.NewSubscriptionRepo(testDB)
	hasher := &hash.SHA256Hasher{}

	mockEmailSender := mockSender.NewMockSender(ctrl)
	emailsService := service.NewEmailsService(mockEmailSender, cfg.Email, cfg.HTTP)

	subService := service.NewSubscriptionService(
		repo,
		hasher,
		emailsService,
	)
	services := &service.Services{Subscriptions: subService}

	handler := handlers.NewHandler(services)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/subscribe", handler.SubscriptionHandler.ShowSubscribePage)
	router.POST("/api/subscribe", handler.SubscriptionHandler.SubscribeEmail)
	router.GET("/api/confirm/:token", handler.SubscriptionHandler.ConfirmEmail)
	router.GET("/api/unsubscribe/:token", handler.SubscriptionHandler.UnsubscribeEmail)

	// Create router with mock template
	router.LoadHTMLGlob(config.GetOriginalPath("templates/**/*.html"))

	cleanup := func() {
		_, err := testDB.Exec(`DELETE FROM subscriptions;`)
		if err != nil {
			t.Fatalf("cleanup failed: could not delete subscriptions data: %v", err)
		}
	}

	return subscriptionTestEnv{
		TestDB:          testDB,
		Router:          router,
		MockEmailSender: mockEmailSender,
		CleanupFunc:     cleanup,
	}
}

func testShowSubscribePageMocked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/subscribe", nil)
	testSettings.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Subscribe to Weather updates")
}

func testSuccessfulSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Mock expectations
	testSettings.MockEmailSender.
		EXPECT().
		Send(gomock.Any()).
		Return(nil)

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBufferString(
		`{"email": "test@example.com", "city": "Kyiv", "frequency": "daily"}`,
	))
	req.Header.Set("Content-Type", "application/json")

	testSettings.Router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	// Check database
	var sub domain.Subscription
	err := testSettings.TestDB.QueryRowx(`
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
	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBufferString(
		`{"email": "test@example.com", "city": "Kyiv", "frequency": "wrong_frequency"}`))
	req.Header.Set("Content-Type", "application/json")

	testSettings.Router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// No database changes expected
	var count int
	err := testSettings.TestDB.QueryRowx(`
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
	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Create existing subscription
	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("existing@example.com", "Kyiv", "daily")
	_, err := testSettings.TestDB.Exec(`
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

	testSettings.Router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusConflict, w.Code)

	// Verify existing record wasn't modified
	var originalToken string
	err = testSettings.TestDB.QueryRowx(`
		SELECT token 
		FROM subscriptions 
		WHERE email = $1
	`, "existing@example.com").Scan(&originalToken)
	assert.NoError(t, err)
	assert.Equal(t, token, originalToken)
}

func testDuplicateEmailSubscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Mock expectations
	testSettings.MockEmailSender.
		EXPECT().
		Send(gomock.Any()).
		Return(nil)

	var createdSubscriptions int

	// Create existing subscription
	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("existing@example.com", "Kyiv", "daily")
	_, err := testSettings.TestDB.Exec(`
		INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
		VALUES ('existing@example.com', 'Kyiv', 'daily', $1, false, NOW())
	`, token)
	assert.NoError(t, err)
	createdSubscriptions++

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/subscribe", bytes.NewBufferString(
		`{"email": "existing@example.com", "city": "Kyiv", "frequency": "hourly"}`, // different frequency
	))
	req.Header.Set("Content-Type", "application/json")

	testSettings.Router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)
	createdSubscriptions++

	// Verify existing record wasn't modified
	var countSubscriptions int
	err = testSettings.TestDB.QueryRowx(`SELECT COUNT(*) FROM subscriptions;`).Scan(&countSubscriptions)
	assert.NoError(t, err)
	assert.Equal(t, createdSubscriptions, countSubscriptions)
}

func testUnsubscribeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Insert subscription to be deleted
	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("unsubscribe@example.com", "Kyiv", "daily")

	_, err := testSettings.TestDB.Exec(`
		INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
		VALUES ('unsubscribe@example.com', 'Kyiv', 'daily', $1, false, NOW())
	`, token)
	assert.NoError(t, err)

	var count int
	err = testSettings.TestDB.QueryRowx(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE token = $1
	`, token).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/unsubscribe/"+token, nil)

	testSettings.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's deleted
	err = testSettings.TestDB.QueryRowx(`
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
	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("non-existen-email@example.com", "Kyiv", "daily")

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/unsubscribe/"+token, nil)

	testSettings.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func testUnsubscribeInvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Insert subscription to be deleted
	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("unsubscribe@example.com", "Kyiv", "daily")

	_, err := testSettings.TestDB.Exec(`
		INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
		VALUES ('unsubscribe@example.com', 'Kyiv', 'daily', $1, false, NOW())
	`, token)
	assert.NoError(t, err)

	var count int
	err = testSettings.TestDB.QueryRowx(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE token = $1
	`, token).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	w := httptest.NewRecorder()
	// Add "bug" instead of valid token
	req := httptest.NewRequest("GET", "/api/unsubscribe/"+"bug", nil)

	testSettings.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify it's not deleted
	err = testSettings.TestDB.QueryRowx(`
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

	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Insert unconfirmed subscription
	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("confirm@example.com", "Kyiv", "daily")

	_, err := testSettings.TestDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('confirm@example.com', 'Kyiv', 'daily', $1, false, NOW())
    `, token)
	assert.NoError(t, err)

	// Verify initial state
	var confirmed bool
	err = testSettings.TestDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.False(t, confirmed, "subscription should be unconfirmed initially")

	// Execute confirmation
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/confirm/"+token, nil)
	testSettings.Router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's confirmed in DB
	err = testSettings.TestDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.True(t, confirmed, "subscription should be confirmed after request")
}

func testConfirmNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("non-existen-email@example.com", "Kyiv", "daily")

	// Execute
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/confirm/"+token, nil)
	testSettings.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func testConfirmInvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testSettings := setupTestEnvironment(t, ctrl)
	defer testSettings.CleanupFunc()

	// Insert unconfirmed subscription
	hasher := &hash.SHA256Hasher{}
	token := hasher.GenerateSubscriptionHash("confirm@example.com", "Kyiv", "daily")

	_, err := testSettings.TestDB.Exec(`
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at)
        VALUES ('confirm@example.com', 'Kyiv', 'daily', $1, false, NOW())
    `, token)
	assert.NoError(t, err)

	// Verify initial state
	var confirmed bool
	err = testSettings.TestDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.False(t, confirmed, "subscription should be unconfirmed initially")

	// Execute confirmation
	w := httptest.NewRecorder()
	// Add "bug" instead of valid token
	req := httptest.NewRequest("GET", "/api/confirm/"+"bug", nil)
	testSettings.Router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify it's not confirmed in DB
	err = testSettings.TestDB.QueryRowx(`
        SELECT confirmed FROM subscriptions WHERE token = $1
    `, token).Scan(&confirmed)
	assert.NoError(t, err)
	assert.False(t, confirmed, "subscription should not be confirmed after request")
}
