package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"weather_forecast_sub/internal/domain"
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
	t.Run(
		"Successful weather request with fallback to second client",
		testSuccessfulWeatherRequestFallbackToSecondClient,
	)
	t.Run("Empty city parameter", testEmptyCityParameter)
	t.Run("City not found", testCityNotFound)
}

func setupTestRouter(h *handlers.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/weather", h.WeatherHandler.GetWeather)
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

	// Setup mock primary client (will return successful response)
	primaryMock := mockClients.NewMockWeatherClient(ctrl)
	primaryMock.EXPECT().
		GetAPICurrentWeather(gomock.Any(), "Kyiv").
		Return(&domain.WeatherResponse{
			Temperature: 25.4,
			Humidity:    70,
			Description: "Sunny",
		}, nil)

	// Setup mock fallback client (should NOT be called)
	fallbackMock := mockClients.NewMockWeatherClient(ctrl)
	fallbackMock.EXPECT().
		GetAPICurrentWeather(gomock.Any(), "Kyiv").
		Times(0) // assert it was not used

	chainClient := clients.NewChainWeatherClient([]clients.WeatherClient{primaryMock, fallbackMock})

	// Setup service and handler
	weatherService := service.NewWeatherService(chainClient)
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

func testSuccessfulWeatherRequestFallbackToSecondClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock clients
	primaryMock := mockClients.NewMockWeatherClient(ctrl)
	fallbackMock := mockClients.NewMockWeatherClient(ctrl)

	// Expect primary to be called and return an error
	primaryMock.EXPECT().
		GetAPICurrentWeather(gomock.Any(), "Kyiv").
		Return(nil, errors.New("primary failed"))

	// Expect fallback to be called and return valid data
	fallbackMock.EXPECT().
		GetAPICurrentWeather(gomock.Any(), "Kyiv").
		Return(&domain.WeatherResponse{
			Temperature: 18.5,
			Humidity:    60,
			Description: "Partly cloudy",
		}, nil)

	// Create chain client and inject into service
	chainClient := clients.NewChainWeatherClient([]clients.WeatherClient{primaryMock, fallbackMock})
	weatherService := service.NewWeatherService(chainClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	// Setup router
	router := setupTestRouter(h)

	// Perform request
	w := performRequest(router, "/api/weather?city=Kyiv")

	// Assert fallback result
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(
		t,
		`{"temperature":18.5,"humidity":60,"description":"Partly cloudy"}`,
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

	// Mock primary client returns ErrCityNotFound
	primaryMock := mockClients.NewMockWeatherClient(ctrl)
	primaryMock.EXPECT().
		GetAPICurrentWeather(gomock.Any(), url.QueryEscape("Київ")).
		Return(nil, customErrors.ErrCityNotFound)

	// Mock fallback client also returns ErrCityNotFound
	fallbackMock := mockClients.NewMockWeatherClient(ctrl)
	fallbackMock.EXPECT().
		GetAPICurrentWeather(gomock.Any(), url.QueryEscape("Київ")).
		Return(nil, customErrors.ErrCityNotFound)

	chainClient := clients.NewChainWeatherClient([]clients.WeatherClient{primaryMock, fallbackMock})

	// Setup service and handler
	weatherService := service.NewWeatherService(chainClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	// Setup router
	router := setupTestRouter(h)

	// Execute request
	w := performRequest(router, "/api/weather?city=Київ")

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, strings.TrimSpace(w.Body.String()))
}
