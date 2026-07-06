.PHONY: dev down down-clean prod down-prod community pro down-prod logs ps help swagger swagger-copy podman-build podman-install podman-up podman-down podman-logs migrate-create migrate-status

COMPOSE_DEV  := USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose -f docker-compose.dev.yml
COMPOSE_PROD := docker compose -f docker-compose.prod.yml

PODMAN_QUADLET_DIR := $(HOME)/.config/containers/systemd
SSC_PUBLIC_DOMAIN  ?= servasec.local
PRO_REPO_DIR       ?= ../servasec-pro

dev: ## Start dev stack
	$(COMPOSE_DEV) up --build -d

down: ## Stop dev stack
	$(COMPOSE_DEV) down

down-clean: ## Stop dev stack and remove volumes
	$(COMPOSE_DEV) down -v --remove-orphans

prod: community ## (alias) Build and start community prod stack

community: ## Build and start community prod stack (free features only)
	$(COMPOSE_PROD) up --build -d

pro: ## Build and start enterprise prod stack (requires servasec-pro repo)
	cp $(PRO_REPO_DIR)/backend/pro/*.go backend/pro/
	BUILD_TAGS=pro $(COMPOSE_PROD) up --build -d

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
	cp backend/docs/swagger.json ../servasec-docs/openapi/swagger.json
	$(MAKE) -C ../servasec-docs gen-api

podman-build: ## Build all Podman images
	podman build -t servasec-backend:latest -f backend/Dockerfile --target prod backend/
	podman build -t servasec-frontend:latest -f frontend/Dockerfile --target prod frontend/
	podman build -t servasec-caddy:latest \
		-f caddy/Dockerfile \
		--build-arg SSC_PUBLIC_DOMAIN=$(SSC_PUBLIC_DOMAIN) \
		caddy/

podman-install: ## Install Quadlet files for current user
	@mkdir -p $(PODMAN_QUADLET_DIR)
	cp deploy/podman/* $(PODMAN_QUADLET_DIR)/
	@echo "Files installed to $(PODMAN_QUADLET_DIR)"
	@echo "Edit secrets in $(PODMAN_QUADLET_DIR)/servasec-backend.container (JWT_SECRET, REFRESH_SECRET, CSRF_SECRET, SSC_ADMIN_PASSWORD)"
	@echo "Then run: make podman-up"

podman-up: podman-build podman-install ## Build, install and start all Quadlet units
	systemctl --user daemon-reload
	systemctl --user start servasec-caddy.service
	@echo "Started. Check status with: systemctl --user status servasec-*"

podman-down: ## Stop all Quadlet units
	systemctl --user stop servasec-caddy.service servasec-frontend.service servasec-backend.service servasec-db.service 2>/dev/null || true
	systemctl --user daemon-reload

podman-logs: ## Tail logs from all servasec units
	journalctl --user -u servasec-caddy -u servasec-frontend -u servasec-backend -u servasec-db -f

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
