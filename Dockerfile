FROM golang:1.24-alpine

ENV GOOS=linux GOARCH=amd64

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o app ./cmd/app/main.go

CMD ["./app"]
