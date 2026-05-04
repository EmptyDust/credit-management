.PHONY: help test-refactor test-backend test-e2e test-integration test-frontend test-platform up-refactor down-refactor logs-refactor dev-backend migrate-backend seed-backend clean

help:
	@echo "Available targets:"
	@echo "  make test-refactor  - Run backend unit tests, backend e2e tests, and frontend build"
	@echo "  make test-backend   - Run modular monolith backend tests"
	@echo "  make test-e2e       - Run modular monolith backend e2e tests"
	@echo "  make test-integration - Run persistence integration tests against local postgres and redis"
	@echo "  make test-frontend  - Build the frontend"
	@echo "  make test-platform  - Run focused platform tests for config, storage, health, and redis"
	@echo "  make up-refactor    - Start postgres, redis, backend, and frontend"
	@echo "  make down-refactor  - Stop the refactor stack"
	@echo "  make logs-refactor  - Tail logs for the refactor stack"
	@echo "  make dev-backend    - Run the modular monolith backend locally"
	@echo "  make migrate-backend - Run backend SQL migrations"
	@echo "  make seed-backend   - Seed backend baseline data"
	@echo "  make clean          - Clean generic test artifacts"

test-refactor: test-backend test-e2e test-frontend

test-backend:
	cd backend && go test ./...

test-e2e:
	cd backend && go test ./tests/e2e/... -v

test-integration:
	@if ! nc -z localhost 5433 >/dev/null 2>&1; then docker compose up -d --wait postgres; fi
	@REDIS_ADDR="$${TEST_REDIS_ADDR:-localhost:6379}"; \
	REDIS_HOST="$${REDIS_ADDR%%:*}"; \
	REDIS_PORT="$${REDIS_ADDR##*:}"; \
	if ! nc -z "$$REDIS_HOST" "$$REDIS_PORT" >/dev/null 2>&1; then docker compose up -d --wait redis; fi
	cd backend && TEST_DATABASE_DSN="$${TEST_DATABASE_DSN:-postgres://postgres:password@localhost:5433/credit_management?sslmode=disable}" TEST_DATABASE_SCHEMA="$${TEST_DATABASE_SCHEMA:-integration_runtime}" TEST_REDIS_ADDR="$${TEST_REDIS_ADDR:-localhost:6379}" TEST_REDIS_DB="$${TEST_REDIS_DB:-15}" go test -tags=integration ./tests/integration/... -v

test-frontend:
	cd frontend && npm exec --yes pnpm -- build

test-platform:
	cd backend && go test ./internal/platform/config ./internal/platform/storage ./internal/platform/health ./internal/platform/redis -v

up-refactor:
	docker compose up -d postgres redis backend frontend

down-refactor:
	docker compose down

logs-refactor:
	docker compose logs -f postgres redis backend frontend

dev-backend:
	cd backend && go run ./cmd/server

migrate-backend:
	cd backend && go run ./cmd/migrate

seed-backend:
	cd backend && go run ./cmd/seed

clean:
	rm -rf coverage/
	find . -name "*.test" -delete
	find . -name "*.out" -delete
