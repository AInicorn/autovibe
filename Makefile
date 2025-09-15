# Autovibe Project Makefile

.PHONY: help vibe build run clean

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

vibe: ## Show project information beautifully rendered
	@if command -v glow >/dev/null 2>&1; then \
		glow AIREADME.md; \
	elif command -v mdcat >/dev/null 2>&1; then \
		mdcat AIREADME.md; \
	elif command -v pandoc >/dev/null 2>&1; then \
		pandoc AIREADME.md -t ansi; \
	else \
		echo "💡 For better rendering, install glow: sudo apt update && sudo apt install glow"; \
		echo ""; \
		cat AIREADME.md; \
	fi

build: ## Build all services
	@echo "Building database service..."
	cd database-service && make build

run: ## Run the database service
	@echo "Starting database service..."
	cd database-service && make run

clean: ## Clean all build artifacts
	@echo "Cleaning all services..."
	cd database-service && make clean

test: ## Run all tests
	@echo "Running tests..."
	cd database-service && make test