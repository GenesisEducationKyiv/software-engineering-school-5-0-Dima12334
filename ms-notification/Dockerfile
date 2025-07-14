# ----------- Build stage ------------
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY ./go.work ./
COPY ./go.work.sum ./

COPY ./pkg ./pkg
COPY ./ms-notification ./ms-notification

RUN go work sync && go mod download all

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

EXPOSE 8000

CMD ["./bin/app"]
