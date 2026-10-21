include .env
export

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATIONS_DIR := migrations

.PHONY: help run build tidy fmt vet migrate-create migrate-up migrate-down migrate-status migrate-reset migrate-version

help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: ## Run server
	go run ./cmd/server

build: ## Build binary
	go build -o bin/server ./cmd/server

tidy: ## go mod tidy
	go mod tidy

fmt: ## go fmt
	go fmt ./...

vet: ## go vet
	go vet ./...

generate: ## Create migration e.g. make generate NAME=create_users_table
ifndef NAME
	$(error NAME is required. Usage: make generate NAME=create_users_table)
endif
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

migrate-up: ## Run migrations up
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-down-to: ## Rollback migration to
ifndef ID
	$(error ID is required. Usage: make migrate-down-to ID=<migration_id>)
endif
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down-to $(ID)

migrate-status: ## Migration status
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-version: ## Current migration version
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" version

migrate-reset: ## Rollback all then migrate up
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" reset
