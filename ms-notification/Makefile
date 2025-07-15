ENV_FILE := .env.dev

help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Up docker containers without build
	docker-compose --env-file $(ENV_FILE) up

up-with-build: ## Build and up docker containers
	docker-compose --env-file $(ENV_FILE) up --build

down: ## Down docker containers
	docker-compose --env-file $(ENV_FILE) down

test-unit: ## Run unit tests only
	go test -v $(shell go list ./... | grep -v 'internal/app\|internal/handlers')

lint: ## Run golangci-lint with remote config (first you need to install this linter on you, look Readme.md)
	@bash -c 'golangci-lint run --config <(curl -sSfL https://raw.githubusercontent.com/fabl3ss/genesis-se-school-linter/refs/heads/main/.golangci.yaml)'
