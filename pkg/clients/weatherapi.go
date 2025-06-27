package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
	"weather_forecast_sub/internal/domain"
	customErrors "weather_forecast_sub/pkg/errors"
	"weather_forecast_sub/pkg/logger"
)

const (
	weatherAPICityNotFoundCode = 1006
	weatherAPIClientTimeout    = 10 * time.Second
)

type WeatherAPIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	next       WeatherClient
}

func NewWeatherAPIClient(apiKey string) *WeatherAPIClient {
	return &WeatherAPIClient{
		apiKey:  apiKey,
		baseURL: "https://api.weatherapi.com/v1",
		httpClient: &http.Client{
			Timeout:   weatherAPIClientTimeout,
			Transport: NewLoggingRoundTripper("WeatherAPIClient"),
		},
	}
}

// WithClient mostly used for testing purposes to inject a custom HTTP client.
func (c *WeatherAPIClient) WithClient(client *http.Client) *WeatherAPIClient {
	c.httpClient = client
	return c
}

// WithBaseURL mostly used for testing purposes to inject a custom base URL.
func (c *WeatherAPIClient) WithBaseURL(baseURL string) *WeatherAPIClient {
	c.baseURL = baseURL
	return c
}

func (c *WeatherAPIClient) setNext(next ChainWeatherProvider) {
	c.next = next
}

type weatherAPIErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type currentWeatherAPIResponse struct {
	Current struct {
		TempC     float32 `json:"temp_c"`
		Humidity  float32 `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

type dayWeatherAPIResponse struct {
	Forecast struct {
		ForecastDay []struct {
			Hour []struct {
				Time      string  `json:"time"` // "2025-05-17 07:00"
				TempC     float32 `json:"temp_c"`
				Humidity  float32 `json:"humidity"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func (c *WeatherAPIClient) processErrorResponse(resp *http.Response) error {
	var apiErr weatherAPIErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
		logger.Errorf("error parsing WeatherAPI error response: %s", err)
		return customErrors.ErrWeatherDataError
	}

	if resp.StatusCode == http.StatusBadRequest && apiErr.Error.Code == weatherAPICityNotFoundCode {
		return customErrors.ErrCityNotFound
	}

	return customErrors.ErrWeatherDataError
}

func (c *WeatherAPIClient) GetAPICurrentWeather(
	ctx context.Context, city string,
) (*domain.WeatherResponse, error) {
	requestURL := fmt.Sprintf("%s/current.json?key=%s&q=%s", c.baseURL, c.apiKey, city)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp.Body, &err)

	if resp.StatusCode != http.StatusOK {
		err = c.processErrorResponse(resp)

		if c.next != nil {
			nextClientName := reflect.TypeOf(c.next).Elem().Name()
			logger.Warnf(
				"WeatherAPIClient.GetAPICurrentWeather() error: %s. "+
					"Passing request to next weather client in chain: %s",
				err,
				nextClientName,
			)
			return c.next.GetAPICurrentWeather(ctx, city)
		}
		return nil, err
	}

	var result currentWeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf("error parsing WeatherAPI response: %s", err.Error())
		return nil, customErrors.ErrWeatherDataError
	}

	return &domain.WeatherResponse{
		Temperature: result.Current.TempC,
		Humidity:    result.Current.Humidity,
		Description: result.Current.Condition.Text,
	}, nil
}

func (c *WeatherAPIClient) GetAPIDayWeather(
	ctx context.Context, city string,
) (*domain.DayWeatherResponse, error) {
	requestURL := fmt.Sprintf("%s/forecast.json?key=%s&q=%s&days=1", c.baseURL, c.apiKey, city)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp.Body, &err)

	if resp.StatusCode != http.StatusOK {
		err = c.processErrorResponse(resp)

		if c.next != nil {
			nextClientName := reflect.TypeOf(c.next).Elem().Name()
			logger.Warnf(
				"WeatherAPIClient.GetAPIDayWeather() error: %s. "+
					"Passing request to next weather client in chain: %s",
				err,
				nextClientName,
			)
			return c.next.GetAPIDayWeather(ctx, city)
		}
		return nil, err
	}

	var result dayWeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf("error parsing WeatherAPI response: %s", err.Error())
		return nil, customErrors.ErrWeatherDataError
	}

	// Map of required times
	targetHours := map[string]*domain.WeatherResponse{
		"07:00": {},
		"10:00": {},
		"13:00": {},
		"16:00": {},
		"19:00": {},
		"22:00": {},
	}

	const dateTimeSplitStrLength = 2 // "2025-05-17 07:00" -> ["2025-05-17", "07:00"]
	for _, hourData := range result.Forecast.ForecastDay[0].Hour {
		dateTimeParts := strings.Split(hourData.Time, " ")
		if len(dateTimeParts) != dateTimeSplitStrLength {
			continue
		}
		timePart := dateTimeParts[1]

		if target, ok := targetHours[timePart]; ok {
			target.Temperature = hourData.TempC
			target.Humidity = hourData.Humidity
			target.Description = hourData.Condition.Text
		}
	}

	return &domain.DayWeatherResponse{
		SevenAM: *targetHours["07:00"],
		TenAM:   *targetHours["10:00"],
		OnePM:   *targetHours["13:00"],
		FourPM:  *targetHours["16:00"],
		SevenPM: *targetHours["19:00"],
		TenPM:   *targetHours["22:00"],
	}, nil
}
