ENV_FILE := .env.dev
TEST_ENV_FILE := .env.test

help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Up docker containers without build
	docker-compose --env-file $(ENV_FILE) up

up-with-build: ## Build and up docker containers
	docker-compose --env-file $(ENV_FILE) up --build

down: ## Down docker containers
	docker-compose --env-file $(ENV_FILE) down

migration: ## Create a new SQL migration. Usage: make migration <name>
	@migrate create -ext sql -dir ./migrations -seq $(filter-out $@,$(MAKECMDGOALS))

migrate-up: ## Apply database migrations
	@docker-compose --env-file $(ENV_FILE) exec app ./bin/migrate up

migrate-down: ## Rollback last database migration
	@docker-compose --env-file $(ENV_FILE) exec app ./bin/migrate down

test: ## Run all tests
	@bash -c '\
		docker-compose -f docker-compose-test.yaml --env-file $(TEST_ENV_FILE) up -d; \
		trap "docker-compose -f docker-compose-test.yaml --env-file $(TEST_ENV_FILE) stop" EXIT; \
		go test -v ./... \
	'

test-unit: ## Run unit tests only
	go test -v $(shell go list ./... | grep -v 'internal/app\|internal/handlers')

test-integration: ## Run integration tests only (with DB)
	@bash -c '\
		docker-compose -f docker-compose-test.yaml --env-file $(TEST_ENV_FILE) up -d; \
		trap "docker-compose -f docker-compose-test.yaml --env-file $(TEST_ENV_FILE) stop" EXIT; \
		go test -v ./internal/app/... ./internal/handlers/... \
	'

swag: ## Generate Swagger docs
	swag init -g internal/app/app.go

lint: ## Run golangci-lint with remote config (first you need to install this linter on you, look Readme.md)
	@bash -c 'golangci-lint run --config <(curl -sSfL https://raw.githubusercontent.com/fabl3ss/genesis-se-school-linter/refs/heads/main/.golangci.yaml)'
