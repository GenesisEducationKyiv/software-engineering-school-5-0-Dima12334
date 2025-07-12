package clients

import (
	"common/logger"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"ms-weather-subscription/internal/domain"
	customErrors "ms-weather-subscription/pkg/errors"
	"net/http"
	"reflect"
	"time"
)

const (
	visualCrossingCityNotFoundMessage = "Bad API Request:Invalid location parameter value."
	visualCrossingTimeout             = 10 * time.Second
)

type VisualCrossingClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	next       WeatherClient
}

func NewVisualCrossingClient(apiKey string) *VisualCrossingClient {
	return &VisualCrossingClient{
		apiKey:  apiKey,
		baseURL: "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline",
		httpClient: &http.Client{
			Timeout:   visualCrossingTimeout,
			Transport: NewLoggingRoundTripper("VisualCrossingClient"),
		},
	}
}

// WithClient mostly used for testing purposes to inject a custom HTTP client.
func (c *VisualCrossingClient) WithClient(client *http.Client) *VisualCrossingClient {
	c.httpClient = client
	return c
}

// WithBaseURL mostly used for testing purposes to inject a custom base URL.
func (c *VisualCrossingClient) WithBaseURL(baseURL string) *VisualCrossingClient {
	c.baseURL = baseURL
	return c
}

func (c *VisualCrossingClient) setNext(next ChainWeatherProvider) {
	c.next = next
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
		return customErrors.ErrWeatherDataError
	}
	bodyStr := string(body)

	if resp.StatusCode == http.StatusBadRequest && bodyStr == visualCrossingCityNotFoundMessage {
		return customErrors.ErrCityNotFound
	}

	return customErrors.ErrWeatherDataError
}

func (c *VisualCrossingClient) GetAPICurrentWeather(
	ctx context.Context, city string,
) (*domain.WeatherResponse, error) {
	requestURL := fmt.Sprintf(
		"%s/%s/today?unitGroup=metric&include=current&key=%s", c.baseURL, city, c.apiKey,
	)

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
				"VisualCrossingClient.GetAPICurrentWeather() error: %s.\n"+
					"Passing request to next weather client in chain: %s",
				err,
				nextClientName,
			)
			return c.next.GetAPICurrentWeather(ctx, city)
		}
		return nil, err
	}

	var result visualCrossingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
	requestURL := fmt.Sprintf(
		"%s/%s/today?unitGroup=metric&include=hours&key=%s", c.baseURL, city, c.apiKey,
	)

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
				"VisualCrossingClient.GetAPIDayWeather() error: %s. "+
					"Passing request to next weather client in chain: %s",
				err,
				nextClientName,
			)
			return c.next.GetAPIDayWeather(ctx, city)
		}
		return nil, err
	}

	var result visualCrossingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
