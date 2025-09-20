.PHONY: setup start stop dev-backend dev-frontend test build clean

# Development setup
setup:
	@echo "Setting up development environment..."
	cd frontend && npm install
	cd backend && go mod download
	mkdir -p data storage

# Start production services
start:
	docker compose up -d

# Stop services
stop:
	docker compose down

# Start backend development server
dev-backend:
	cd backend && go run cmd/api/main.go

# Start frontend development server
dev-frontend:
	cd frontend && npm run dev

# Run all tests
test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...
	@echo "Running frontend tests..."
	cd frontend && npm test

# Build application
build:
	docker compose build

# Clean build artifacts
clean:
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf data/*.db
	rm -rf storage/*
	docker compose down -v

# Help
help:
	@echo "Available commands:"
	@echo "  make setup         - Install dependencies and create required directories"
	@echo "  make start         - Start production services"
	@echo "  make stop          - Stop services"
	@echo "  make dev-backend   - Start backend development server"
	@echo "  make dev-frontend  - Start frontend development server"
	@echo "  make test          - Run all tests"
	@echo "  make build         - Build Docker images"
	@echo "  make clean         - Remove build artifacts and data"