package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"weather_forecast_sub/internal/domain"
	customErrors "weather_forecast_sub/pkg/errors"
	"weather_forecast_sub/pkg/logger"
)

const (
	visualCrossingCityNotFoundMessage = "Bad API Request:Invalid location parameter value."
	visualCrossingTimeout             = 10 * time.Second
)

type VisualCrossingClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewVisualCrossingClient(apiKey string) *VisualCrossingClient {
	return &VisualCrossingClient{
		APIKey:     apiKey,
		BaseURL:    "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline",
		HTTPClient: &http.Client{Timeout: visualCrossingTimeout},
	}
}

type visualCrossingResponse struct {
	CurrentConditions struct {
		Temp       float32 `json:"temp"`
		Humidity   float32 `json:"humidity"`
		Conditions string  `json:"conditions"`
	} `json:"currentConditions"`
	Days []struct {
		Datetime string `json:"datetime"`
		Hours    []struct {
			Datetime   string  `json:"datetime"`
			Temp       float32 `json:"temp"`
			Humidity   float32 `json:"humidity"`
			Conditions string  `json:"conditions"`
		} `json:"hours"`
	} `json:"days"`
}

func (c *VisualCrossingClient) processErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("error reading VisualCrossing response body: %s", err)
		return customErrors.ErrWeatherDataError
	}
	bodyStr := string(body)

	if resp.StatusCode == http.StatusBadRequest && bodyStr == visualCrossingCityNotFoundMessage {
		return customErrors.ErrCityNotFound
	}

	logger.Errorf("VisualCrossing API error. Status code: %d, Message: %s", resp.StatusCode, bodyStr)
	return customErrors.ErrWeatherDataError
}

func (c *VisualCrossingClient) GetAPICurrentWeather(
	ctx context.Context, city string,
) (*domain.WeatherResponse, error) {
	url := fmt.Sprintf(
		"%s/%s/today?unitGroup=metric&include=current&key=%s", c.BaseURL, city, c.APIKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("failed to create VisualCrossing request: %s", err)
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Errorf("error making request to VisualCrossing API: %s", err)
		return nil, err
	}
	defer closeBody(resp.Body, &err)

	if resp.StatusCode != http.StatusOK {
		err = c.processErrorResponse(resp)
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("failed to read VisualCrossing response body: %s", err)
		return nil, customErrors.ErrWeatherDataError
	}

	logger.Infof("VisualCrossing API success response for city %s: %s", city, string(respBody))

	var result visualCrossingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Errorf("error decoding VisualCrossing current weather: %s", err)
		return nil, customErrors.ErrWeatherDataError
	}

	return &domain.WeatherResponse{
		Temperature: result.CurrentConditions.Temp,
		Humidity:    result.CurrentConditions.Humidity,
		Description: result.CurrentConditions.Conditions,
	}, nil
}

func (c *VisualCrossingClient) GetAPIDayWeather(
	ctx context.Context, city string,
) (*domain.DayWeatherResponse, error) {
	url := fmt.Sprintf(
		"%s/%s/today?unitGroup=metric&include=hours&key=%s&contentType=json", c.BaseURL, city, c.APIKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("failed to create VisualCrossing request: %s", err)
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Errorf("error making request to VisualCrossing API: %s", err)
		return nil, err
	}
	defer closeBody(resp.Body, &err)

	if resp.StatusCode != http.StatusOK {
		err = c.processErrorResponse(resp)
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("failed to read VisualCrossing response body: %s", err)
		return nil, customErrors.ErrWeatherDataError
	}

	logger.Infof("VisualCrossing API success response for city %s: %s", city, string(respBody))

	var result visualCrossingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Errorf("error decoding VisualCrossing forecast: %s", err)
		return nil, customErrors.ErrWeatherDataError
	}

	targetHours := map[string]*domain.WeatherResponse{
		"07:00:00": {},
		"10:00:00": {},
		"13:00:00": {},
		"16:00:00": {},
		"19:00:00": {},
		"22:00:00": {},
	}

	for _, hour := range result.Days[0].Hours {
		if target, ok := targetHours[hour.Datetime]; ok {
			target.Temperature = hour.Temp
			target.Humidity = hour.Humidity
			target.Description = hour.Conditions
		}
	}

	return &domain.DayWeatherResponse{
		SevenAM: *targetHours["07:00:00"],
		TenAM:   *targetHours["10:00:00"],
		OnePM:   *targetHours["13:00:00"],
		FourPM:  *targetHours["16:00:00"],
		SevenPM: *targetHours["19:00:00"],
		TenPM:   *targetHours["22:00:00"],
	}, nil
}
