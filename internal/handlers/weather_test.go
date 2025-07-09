package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"weather_forecast_sub/internal/handlers"
	"weather_forecast_sub/internal/service"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

func fakeNewWeatherAPIClient(fakePrimaryServer *httptest.Server) *clients.WeatherAPIClient {
	return clients.NewWeatherAPIClient(
		"dummy-key",
	).WithBaseURL(fakePrimaryServer.URL).WithClient(fakePrimaryServer.Client())
}

func fakeNewVisualCrossingClient(fakePrimaryServer *httptest.Server) *clients.VisualCrossingClient {
	return clients.NewVisualCrossingClient(
		"dummy-key",
	).WithBaseURL(fakePrimaryServer.URL).WithClient(fakePrimaryServer.Client())
}

func setupChainWeatherClient(
	primaryServer, fallbackServer *httptest.Server,
) (*clients.ChainWeatherClient, error) {
	primaryClient := fakeNewWeatherAPIClient(primaryServer)
	fallbackClient := fakeNewVisualCrossingClient(fallbackServer)
	chainClient, err := clients.NewChainWeatherClient(
		[]clients.ChainWeatherProvider{primaryClient, fallbackClient},
	)

	return chainClient, err
}

func testSuccessfulWeatherRequest(t *testing.T) {
	// Setup a fake WeatherAPI server that returns a successful weather response
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{
			"current": {
				"temp_c": 25.4,
				"humidity": 70,
				"condition": {
					"text": "Sunny"
				}
			}
		}`))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer primaryServer.Close()

	// Setup a fake VisualCrossing server that would return 500 (should not be called in this test)
	fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("fallback server should NOT be called")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer fallbackServer.Close()

	chainClient, err := setupChainWeatherClient(primaryServer, fallbackServer)
	if err != nil {
		t.Fatalf("failed to create chain weather client: %v", err)
	}

	redisCache := testutils.SetupTestCache(t)
	cachingWeatherClient := clients.NewCachingWeatherClient(chainClient, redisCache)

	weatherService := service.NewWeatherService(cachingWeatherClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	router := setupTestRouter(h)

	w := performRequest(router, "/api/weather?city=Kyiv")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(
		t,
		`{"temperature":25.4,"humidity":70,"description":"Sunny"}`,
		strings.TrimSpace(w.Body.String()),
	)
}

func testSuccessfulWeatherRequestFallbackToSecondClient(t *testing.T) {
	// Setup fake WeatherAPI server that fails
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate internal server error
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"error": "internal error"}`))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer primaryServer.Close()

	// Setup VisualCrossing server that succeeds
	fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{
			"currentConditions": {
				"temp": 18.5,
				"humidity": 60,
				"conditions": "Partly cloudy"
			},
			"days": []
		}`))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer fallbackServer.Close()

	chainClient, err := setupChainWeatherClient(primaryServer, fallbackServer)
	if err != nil {
		t.Fatalf("failed to create chain weather client: %v", err)
	}

	redisCache := testutils.SetupTestCache(t)
	cachingWeatherClient := clients.NewCachingWeatherClient(chainClient, redisCache)

	weatherService := service.NewWeatherService(cachingWeatherClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	router := setupTestRouter(h)

	w := performRequest(router, "/api/weather?city=Kyiv")

	// Assert fallback (VisualCrossing) result
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t,
		`{"temperature":18.5,"humidity":60,"description":"Partly cloudy"}`,
		strings.TrimSpace(w.Body.String()),
	)
}

func testEmptyCityParameter(t *testing.T) {
	// Use dummy client that will never be called
	dummyServer := httptest.NewServer(http.NotFoundHandler())
	defer dummyServer.Close()

	client := fakeNewWeatherAPIClient(dummyServer)

	redisCache := testutils.SetupTestCache(t)
	cachingWeatherClient := clients.NewCachingWeatherClient(client, redisCache)

	weatherService := service.NewWeatherService(cachingWeatherClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	router := setupTestRouter(h)

	w := performRequest(router, "/api/weather") // no city param

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, strings.TrimSpace(w.Body.String()))
}

func testCityNotFound(t *testing.T) {
	// Fake WeatherAPI response (400 + code 1006)
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{
			"error": {
				"code": 1006,
				"message": "No matching location found."
			}
		}`))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer primaryServer.Close()

	// Fake VisualCrossing response (400 + plain-text body)
	fallbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Bad API Request:Invalid location parameter value."))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer fallbackServer.Close()

	chainClient, err := setupChainWeatherClient(primaryServer, fallbackServer)
	if err != nil {
		t.Fatalf("failed to create chain weather client: %v", err)
	}

	redisCache := testutils.SetupTestCache(t)
	cachingWeatherClient := clients.NewCachingWeatherClient(chainClient, redisCache)

	weatherService := service.NewWeatherService(cachingWeatherClient)
	h := handlers.NewHandler(&service.Services{Weather: weatherService})

	router := setupTestRouter(h)

	w := performRequest(router, "/api/weather?city=Київ")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, strings.TrimSpace(w.Body.String()))
}
