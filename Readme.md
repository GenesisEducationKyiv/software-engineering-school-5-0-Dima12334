# Weather Forecast API

## Overview

The task involves developing a REST API for creating subscriptions on weather updates using Go and PostgreSQL.

## Deploy
The implemented service is deployed on render.com: https://weather-forecast-sub-app.onrender.com/subscribe

## General info

You can use Postman [collection](https://www.postman.com/dimchik32/workspace/weather-subscription-service/collection/25524341-d34e28e2-0887-4300-9329-37dd06732ab4?action=share&creator=25524341
) to interact with API.

Swagger documentation for API: https://weather-forecast-sub-app.onrender.com/swagger/index.html<br>
Local version: http://localhost:8080/swagger/index.html

## Usage

1. Clone this repository
```
https://github.com/Dima12334/weather_forecast_sub.git
```
2. Create `.env.dev` file in the root directory and fill it with variables as in `.env.dev.example`
3. Build and up docker containers:
```
make up-with-build
```
4. Apply migrations:
```
make migrate-up
```
5. Done. Use the App.<br>
You can open http://localhost:8080/subscribe page and fill out the form.<br>
After that you will receive email with a confirmation link, and after confirmation you will start receiving weather updates.<br>
You can unsubscribe from the newsletter at any time by using the unsubscribe link in email.<br>

## Optionally you can
1. Run tests:
```
make test
```
2. Install and run linter:
- Installation
```
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
```
- Run
```
make lint
```
