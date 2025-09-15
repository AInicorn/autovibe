# Autovibe Database Service Makefile

.PHONY: help build run test clean generate proto

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the database service
	@echo "Building database service..."
	go build -o bin/database-service ./cmd/database-service

run: ## Run the database service
	@echo "Running database service..."
	go run ./cmd/database-service

test: ## Run tests
	@echo "Running tests..."
	go test ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f database-service

generate: ## Generate Ent ORM code
	@echo "Generating Ent code..."
	go generate ./ent

proto: ## Generate protobuf code (TODO)
	@echo "Protobuf generation not yet implemented"
	@echo "TODO: Add protoc commands when .proto files are ready"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...