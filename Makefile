.PHONY: dev down down-clean prod down-prod community community-build pro pro-build dev-pro logs ps help swagger swagger-copy migrate-create migrate-status

COMPOSE_DEV  := USER_UID=$(shell id -u) USER_GID=$(shell id -g) GIT_VERSION=$(shell ./scripts/version.sh) docker compose -f docker-compose.dev.yml
COMPOSE_PROD := GIT_VERSION=$(shell ./scripts/version.sh) docker compose -f docker-compose.prod.yml
COMPOSE_BUILD := GIT_VERSION=$(shell ./scripts/version.sh) docker compose -f docker-compose.prod.yml -f docker-compose.build.yml

SSC_PUBLIC_DOMAIN  ?= servasec.local
PRO_REPO_DIR       ?= ../servasec-pro

dev: ## Start dev stack
	$(COMPOSE_DEV) up --build -d

down: ## Stop dev stack
	$(COMPOSE_DEV) down

down-clean: ## Stop dev stack and remove volumes
	$(COMPOSE_DEV) down -v --remove-orphans

prod: community ## (alias) Pull and start community prod stack

community: ## Pull and start community prod images (registry.gitlab.com)
	$(COMPOSE_PROD) up -d

community-build: ## Build and start community prod stack from source
	$(COMPOSE_BUILD) up --build -d

pro: ## Pull and start pro prod stack (requires docker login for the private servasec-pro project)
	SSC_BACKEND_IMAGE=registry.gitlab.com/servasec/servasec-pro/backend $(COMPOSE_PROD) up -d

pro-build: ## Build and start pro prod stack (requires servasec-pro repo)
	cp $(PRO_REPO_DIR)/backend/pro/*.go backend/pro/
	BUILD_TAGS=pro $(COMPOSE_BUILD) up --build -d

dev-pro: ## Start dev stack with pro features (requires servasec-pro repo)
	cp $(PRO_REPO_DIR)/backend/pro/*.go backend/pro/
	BUILD_TAGS=pro $(COMPOSE_DEV) up --build -d

down-prod: ## Stop prod stack
	$(COMPOSE_PROD) down

logs: ## Show all logs (dev)
	$(COMPOSE_DEV) logs -f

ps: ## Show container status (dev)
	$(COMPOSE_DEV) ps

swagger: ## Generate swagger.json from Go annotations
	cd backend && swag init --parseDependency --parseInternal --output docs

swagger-copy: swagger ## Copy swagger.json to servasec-docs and regenerate API docs
	cp backend/docs/swagger.json ../servasec-docs/static/openapi/swagger.json
	$(MAKE) -C ../servasec-docs gen-api

migrate-create: ## Create a new migration: make migrate-create NAME=add_scan_metadata
	@cd backend && bash -c '\
		last=$$(ls -1 migrations/[0-9][0-9][0-9]_*.sql 2>/dev/null | tail -1 | grep -oE "^[0-9]+"); \
		next=$$(printf "%03d" $$(( $${last:-0} + 1 ))); \
		file="migrations/$${next}_$(NAME).sql"; \
		printf -- "-- +goose Up\n\n\n-- +goose Down\n" > "$$file"; \
		echo "Created: backend/$$file"; \
	'

migrate-status: ## Show migration status (requires running stack)
	$(COMPOSE_DEV) exec backend sh -c 'apk add --no-cache postgresql-client 2>/dev/null; PGPASSWORD=$$POSTGRES_PASSWORD psql -h db -U $$POSTGRES_USER -d $$POSTGRES_DB -c "SELECT version_id, is_applied, tstamp FROM goose_db_version ORDER BY version_id"'

migrate-down: ## Rollback last migration (requires running stack)
	@echo "Rollback is not automated. To revert: restore from backup or apply manually."
	@echo "  docker compose exec db pg_dump -U $$$$POSTGRES_USER $$$$POSTGRES_DB > pre_rollback.sql"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
