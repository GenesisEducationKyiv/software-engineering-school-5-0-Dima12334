# ----------- Build stage ------------
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY ./ms-notification/go.mod ./ms-notification/go.sum ./
RUN go mod download

COPY ./ms-notification ./ms-notification

RUN go build -ldflags="-s -w" -o bin/app ./ms-notification/cmd/app/main.go

# ----------- Run stage ------------
FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/app ./bin/app
COPY --from=builder /app/ms-notification/configs ./configs
COPY --from=builder /app/ms-notification/templates ./templates

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./bin/app"]
