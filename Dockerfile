# ----------- Build stage ------------
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o bin/migrate ./cmd/migrate/main.go
RUN go build -ldflags="-s -w" -o bin/app ./cmd/app/main.go

# ----------- Run stage ------------
FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/migrate ./bin/migrate
COPY --from=builder /app/bin/app ./bin/app
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/docs ./docs

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./bin/app"]
