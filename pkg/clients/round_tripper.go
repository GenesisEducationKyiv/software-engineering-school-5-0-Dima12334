package clients

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"weather_forecast_sub/pkg/logger"
)

type LoggingRoundTripper struct {
	Transport  http.RoundTripper
	ClientName string
}

func NewLoggingRoundTripper(clientName string) *LoggingRoundTripper {
	return &LoggingRoundTripper{
		Transport:  http.DefaultTransport,
		ClientName: clientName,
	}
}

func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var reqBodyBytes []byte
	if req.Body != nil {
		reqBodyBytes, err := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes)) // Reset for real transport
		if err != nil {
			logger.Errorf("[HTTP Client: %s] failed to read request body: %s", lrt.ClientName, err)
			return nil, err
		}
	}

	start := time.Now()
	resp, err := lrt.Transport.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		logger.Errorf("[HTTP Client: %s] request failed: %s", lrt.ClientName, err)
		return resp, err
	}

	var respBodyBytes []byte
	if resp.Body != nil {
		respBodyBytes, err = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes)) // Reset for client usage
		if err != nil {
			logger.Errorf("[HTTP Client: %s] failed to read response body: %s", lrt.ClientName, err)
			return resp, err
		}
	}

	reqURL, err := replaceAPIKey(req.URL.String())
	if err != nil {
		logger.Errorf("[HTTP Client: %s]: %s", lrt.ClientName, err)
		return resp, err
	}

	logger.Infof(
		`[HTTP Client: %s]
Request URL: %s
Status Code: %d
Duration: %s
Request Body: %s
Response Body: %s`,
		lrt.ClientName,
		reqURL,
		resp.StatusCode,
		duration,
		truncateIfNeeded(string(reqBodyBytes)),
		truncateIfNeeded(string(respBodyBytes)),
	)

	return resp, nil
}

// replaceAPIKey used to replace API key from url in logs.
func replaceAPIKey(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	query := parsedURL.Query()
	query.Set("key", "API_KEY")
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

func truncateIfNeeded(body string) string {
	const maxLen = 1000
	if len(body) > maxLen {
		return body[:maxLen] + "...[truncated]"
	}
	return body
}
