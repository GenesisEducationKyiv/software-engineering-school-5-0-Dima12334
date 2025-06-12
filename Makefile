help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Up docker containers without build
	docker-compose --env-file .env.dev up

up-with-build: ## Build and up docker containers
	docker-compose --env-file .env.dev up --build

down: ## Down docker containers
	docker-compose down

migration: ## Create a new SQL migration. Usage: make migration <name>
	@migrate create -ext sql -dir ./migrations -seq $(filter-out $@,$(MAKECMDGOALS))

migrate-up: ## Apply database migrations
	@docker-compose exec app ./bin/migrate up

migrate-down: ## Rollback last database migration
	@docker-compose exec app ./bin/migrate down

test: ## Run all tests
	go test -v ./...

swag: ## Generate Swagger docs
	swag init -g internal/app/app.go

lint: ## Run golangci-lint with remote config (first you need to install this linter on you, look Readme.md)
	bash -c 'golangci-lint run --config <(curl -sSfL https://raw.githubusercontent.com/fabl3ss/genesis-se-school-linter/refs/heads/main/.golangci.yaml)'
