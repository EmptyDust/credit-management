.PHONY: help test-refactor test-backend test-e2e test-frontend up-refactor down-refactor logs-refactor dev-backend clean

help:
	@echo "Available targets:"
	@echo "  make test-refactor  - Run backend unit tests, backend e2e tests, and frontend build"
	@echo "  make test-backend   - Run modular monolith backend tests"
	@echo "  make test-e2e       - Run modular monolith backend e2e tests"
	@echo "  make test-frontend  - Build the frontend"
	@echo "  make up-refactor    - Start postgres, redis, backend, and frontend"
	@echo "  make down-refactor  - Stop the refactor stack"
	@echo "  make logs-refactor  - Tail logs for the refactor stack"
	@echo "  make dev-backend    - Run the modular monolith backend locally"
	@echo "  make clean          - Clean generic test artifacts"

test-refactor: test-backend test-e2e test-frontend

test-backend:
	cd backend && go test ./...

test-e2e:
	cd backend && go test ./tests/e2e/... -v

test-frontend:
	cd frontend && npm exec --yes pnpm -- build

up-refactor:
	docker compose up -d postgres redis backend frontend

down-refactor:
	docker compose down

logs-refactor:
	docker compose logs -f postgres redis backend frontend

dev-backend:
	cd backend && go run ./cmd/server

clean:
	rm -rf coverage/
	find . -name "*.test" -delete
	find . -name "*.out" -delete
