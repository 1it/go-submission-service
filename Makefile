.PHONY: run docker-run stop build docker-build test test-e2e clean help

# Default target
all: help

# Show help information
help:
	@echo "Go Submission Service - Make targets"
	@echo ""
	@echo "Development:"
	@echo "  run           - Run the service locally (go run)"
	@echo "  docker-run    - Run with Docker Compose (service + MailHog)"
	@echo "  stop          - Stop Docker Compose services"
	@echo ""
	@echo "Build:"
	@echo "  build         - Build the Go binary locally"
	@echo "  docker-build  - Build the Docker image"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run unit tests"
	@echo "  test-e2e      - Run end-to-end tests (starts compose, runs e2e-test.sh)"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean         - Stop containers and remove volumes"

# Run the service locally
run:
	@go mod tidy
	@go run main.go

# Run with Docker Compose (recommended for quick start with MailHog)
docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up --build

# Stop Docker Compose services
stop:
	@echo "Stopping services..."
	@docker-compose down

# Build the Go binary locally
build:
	@go build -o go-submission-service .

# Build the Docker image (no push)
docker-build:
	@docker-compose build

# Run unit tests
test:
	@go test -v ./...

# Run end-to-end tests
test-e2e: stop
	@echo "Running E2E tests..."
	@docker-compose up -d --build
	@echo "Waiting for services to start..."
	@sleep 10
	@docker-compose ps
	@./e2e-test.sh

# Clean up
clean: stop
	@echo "Cleaning up..."
	@docker-compose down -v --remove-orphans 2>/dev/null || true
	@rm -f go-submission-service
	@echo "Cleanup complete."
