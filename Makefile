# ============================================================================
# Cofre — developer task runner
# Uses Docker Compose v2 (`docker compose`, not the legacy `docker-compose`).
# Run `make` or `make help` to list available targets.
# ============================================================================

COMPOSE := docker compose
DB_USER ?= finance_sh
DB_NAME ?= finance_sh

.DEFAULT_GOAL := help

.PHONY: help up down restart logs ps build seed db-shell \
        backend-dev frontend-dev backend-build frontend-build \
        tidy lint test swagger clean fresh backup restore \

help: ## Show this help
	@echo "finance.sh — make targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ---- Stack lifecycle (Docker Compose) --------------------------------------
up: ## Start the whole stack in the background
	$(COMPOSE) up -d

down: ## Stop and remove containers (keeps volumes)
	$(COMPOSE) down

restart: ## Restart all services
	$(COMPOSE) restart

logs: ## Tail logs from all services
	$(COMPOSE) logs -f

ps: ## Show service status
	$(COMPOSE) ps

build: ## Build all images
	$(COMPOSE) build

# ---- Data / shells ----------------------------------------------------------
seed: ## Seed demo data (restarts backend with SEED=true)
	SEED=true $(COMPOSE) up -d --force-recreate --no-deps backend

db-shell: ## Open a psql shell on postgres
	$(COMPOSE) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

# ---- Local development (no Docker for the app) ------------------------------
backend-dev: ## Run the Go API locally (needs only postgres up)
	$(COMPOSE) up -d postgres
	cd backend && go run ./cmd/api

frontend-dev: ## Run the Vite dev server locally (frontend SPA)
	cd frontend && npm run dev

backend-build: ## Compile the backend binary
	cd backend && go build ./...

frontend-build: ## Build the frontend static bundle
	cd frontend && npm run build

# ---- Quality ----------------------------------------------------------------
tidy: ## go mod tidy
	cd backend && go mod tidy

lint: ## Lint backend (go vet) and frontend (eslint)
	cd backend && go vet ./...
	cd frontend && npm run lint

test: ## Run backend tests
	cd backend && go test ./...

swagger: ## Reminder: the OpenAPI spec is hand-written
	@echo "OpenAPI is hand-written at backend/docs/openapi.yaml — edit it directly."
	@echo "Swagger UI is served by the backend at http://localhost:8080/swagger"

# ---- Backups (encrypted) ----------------------------------------------------
backup: ## Encrypted pg_dump (needs BACKUP_PASSPHRASE; see scripts/backup.sh)
	./scripts/backup.sh

restore: ## Restore an encrypted dump: make restore FILE=backups/<file>.sql.gpg
	./scripts/restore.sh "$(FILE)"

# ---- Cleanup / reset --------------------------------------------------------
clean: ## Stop stack, remove volumes, prune dangling volumes
	$(COMPOSE) down -v
	docker volume prune -f

fresh: clean build up seed ## Full reset: clean, rebuild, start, seed
	@echo "Fresh stack is up (2 containers: postgres + app)."
	@echo "  App:  http://127.0.0.1:8090  (HTTP puro — ponha seu reverse proxy na frente p/ TLS)"
	@echo "  API:  http://127.0.0.1:8090/api/v1  (same-origin, servida pelo mesmo binário)"
