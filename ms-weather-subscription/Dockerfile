# ----------- Build stage ------------
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY ./go.work ./
COPY ./go.work.sum ./

COPY ./pkg ./pkg
COPY ./ms-notification ./ms-notification
COPY ./ms-weather-subscription ./ms-weather-subscription
COPY ./api/proto/stubs ./api/proto/stubs

RUN go work sync && go mod download all

RUN go build -ldflags="-s -w" -o bin/migrate ./ms-weather-subscription/cmd/migrate/main.go
RUN go build -ldflags="-s -w" -o bin/app ./ms-weather-subscription/cmd/app/main.go

# ----------- Run stage ------------
FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/migrate ./bin/migrate
COPY --from=builder /app/bin/app ./bin/app
COPY --from=builder /app/ms-weather-subscription/configs ./ms-weather-subscription/configs
COPY --from=builder /app/ms-weather-subscription/templates ./ms-weather-subscription/templates
COPY --from=builder /app/ms-weather-subscription/migrations ./ms-weather-subscription/migrations
COPY --from=builder /app/ms-weather-subscription/docs ./ms-weather-subscription/docs
COPY --from=builder /app/go.work ./go.work
COPY --from=builder /app/go.work.sum ./go.work.sum

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./bin/app"]
