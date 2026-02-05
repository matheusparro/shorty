# Docker commands
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

# App commands
run:
	go run ./cmd/api-server

# Database connection
DB_URL := postgres://shorty:shorty@localhost:5432/shorty?sslmode=disable
MIGRATIONS_DIR := ./migrations

# Migration commands
.PHONY: migrate-up migrate-down migrate-down-all migrate-version migrate-create migrate-force

migrate-up: ## Run all pending migrations
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down: ## Rollback last migration
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-down-all: ## Rollback all migrations (DANGER!)
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down -all

migrate-version: ## Show current migration version
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-create: ## Create new migration (use: make migrate-create NAME=add_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=add_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-force: ## Force set migration version (use: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)

# Combined commands
db-reset: down up migrate-up ## Reset database (destroy and recreate with migrations)
	@echo "✅ Database reset complete!"

dev: up migrate-up run ## Start docker, run migrations, and start app

help: ## Show available commands
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'