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
2. Create `.env.dev` files in the `ms-notification` and `ms-weather-subscription` directories and fill them with variables as in `.env.dev.example`
3. Create general docker network:
```
docker network create microservices-net
```
4. Build and up docker containers:
- Build and start the notification service:
```
cd notification && make up-with-build
```
- Return to the root directory:
```
cd ..
```
- Build and start the weather subscription service:
```
cd ms-weather-subscription && make up-with-build
```
5. Apply migrations in the `ms-weather-subscription` service:
```
make migrate-up
```
6. Done. Use the App.<br>
You can open http://localhost:8080/subscribe page and fill out the form.<br>
After that you will receive email with a confirmation link, and after confirmation you will start receiving weather updates.<br>
You can unsubscribe from the newsletter at any time by using the unsubscribe link in email.<br>

## Optionally you can run these commands for each package
1. Run tests (for more detailed information looks at `Makefile`)
```
make help
```
2. Run linter (make sure you have installed golangci-lint):
- Installation
```
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
```
- Run
```
make lint
```

