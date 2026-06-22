.PHONY: dev down down-clean prod down-prod logs ps help

COMPOSE_DEV  := USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose -f docker-compose.dev.yml
COMPOSE_PROD := docker compose -f docker-compose.prod.yml

dev: ## Start dev stack
	$(COMPOSE_DEV) up --build -d

down: ## Stop dev stack
	$(COMPOSE_DEV) down

down-clean: ## Stop dev stack and remove volumes
	$(COMPOSE_DEV) down -v --remove-orphans

prod: ## Build and start prod stack
	$(COMPOSE_PROD) up --build -d

down-prod: ## Stop prod stack
	$(COMPOSE_PROD) down

logs: ## Show all logs (dev)
	$(COMPOSE_DEV) logs -f

ps: ## Show container status (dev)
	$(COMPOSE_DEV) ps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
