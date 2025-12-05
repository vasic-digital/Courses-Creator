# Course Creator Makefile

# Variables
DOCKER_COMPOSE := docker-compose
GO := go

.PHONY: help build run test clean docker-build docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the Go application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  lint         - Run linter"
	@echo "  format       - Format code"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker images"
	@echo "  docker-up    - Start services with Docker Compose"
	@echo "  docker-down  - Stop services"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  dev-setup    - Set up development environment"

# Build application
build:
	$(GO) build -o bin/core-processor ./core-processor

# Run application
run:
	$(GO) run ./core-processor

# Run tests
test:
	$(GO) test -v ./...

# Run tests with coverage
test-cover:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	$(GO) vet ./...
	golangci-lint run

# Format code
format:
	$(GO) fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker operations
docker-build:
	$(DOCKER_COMPOSE) build

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

# Database migrations
migrate-up:
	$(GO) run ./core-processor/cmd/migrate/main.go -action up

migrate-down:
	$(GO) run ./core-processor/cmd/migrate/main.go -action down

# Development setup
dev-setup:
	cp .env.example .env
	@echo "Please edit .env file with your configuration"
	$(GO) mod download
	npm install --prefix creator-app
	npm install --prefix player-app

# Install SSL certificates for development
ssl-setup:
	openssl req -x509 -newkey rsa:4096 -keyout nginx/ssl/key.pem -out nginx/ssl/cert.pem -days 365 -nodes -subj "/C=US/ST=CA/L=SF/O=CourseCreator/CN=localhost"