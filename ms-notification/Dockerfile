# ----------- Build stage ------------
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY ./go.work ./
COPY ./go.work.sum ./

COPY ./pkg ./pkg
COPY ./ms-notification ./ms-notification
COPY ./ms-weather-subscription ./ms-weather-subscription

RUN go work sync && go mod download all

RUN go build -ldflags="-s -w" -o bin/app ./ms-notification/cmd/app/main.go

# ----------- Run stage ------------
FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/app ./bin/app
COPY --from=builder /app/ms-notification/configs ./ms-notification/configs
COPY --from=builder /app/ms-notification/templates ./ms-notification/templates
COPY --from=builder /app/go.work ./go.work
COPY --from=builder /app/go.work.sum ./go.work.sum

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8000

CMD ["./bin/app"]
