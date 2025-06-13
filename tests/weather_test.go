package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"weather_forecast_sub/internal/handlers"
	"weather_forecast_sub/internal/service"
	"weather_forecast_sub/pkg/clients"
	mockClients "weather_forecast_sub/pkg/clients/mocks"
	customErrors "weather_forecast_sub/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWeather(t *testing.T) {
	t.Run("Successful weather request", testSuccessfulWeatherRequest)
	t.Run("Empty city parameter", testEmptyCityParameter)
	t.Run("City not found", testCityNotFound)
}

func setupTestRouter(h *handlers.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/weather", h.GetWeather)
	return router
}

func performRequest(router *gin.Engine, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func testSuccessfulWeatherRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock client
	mockClient := mockClients.NewMockWeatherClient(ctrl)
	mockClient.EXPECT().
		GetAPICurrentWeather(gomock.Any(), "Kyiv").
		Return(&clients.WeatherResponse{
			Temperature: 25.4,
			Humidity:    70,
			Description: "Sunny",
		}, nil)

	// Setup service and handler
	weatherService := service.NewWeatherService(mockClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	// Setup router
	router := setupTestRouter(h)

	// Execute request
	w := performRequest(router, "/api/weather?city=Kyiv")

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(
		t,
		`{"temperature":25.4,"humidity":70,"description":"Sunny"}`,
		strings.TrimSpace(w.Body.String()),
	)
}

func testEmptyCityParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock client (no expectations needed)
	mockClient := mockClients.NewMockWeatherClient(ctrl)

	// Setup service and handler
	weatherService := service.NewWeatherService(mockClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	// Setup router
	router := setupTestRouter(h)

	// Execute request
	w := performRequest(router, "/api/weather")

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, strings.TrimSpace(w.Body.String()))
}

func testCityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock client
	mockClient := mockClients.NewMockWeatherClient(ctrl)
	mockClient.EXPECT().
		GetAPICurrentWeather(gomock.Any(), url.QueryEscape("Київ")).
		Return(nil, customErrors.ErrCityNotFound)

	// Setup service and handler
	weatherService := service.NewWeatherService(mockClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	// Setup router
	router := setupTestRouter(h)

	// Execute request
	w := performRequest(router, "/api/weather?city=Київ")

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, strings.TrimSpace(w.Body.String()))
}
