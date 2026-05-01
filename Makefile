.PHONY: help check setup build dev docker prod test clean install-deps

# Colors
GREEN := \033[0;32m
BLUE := \033[0;34m
NC := \033[0m # No Color

help:
	@echo "$(BLUE)HDF5 Agent - Available Commands:$(NC)"
	@echo ""
	@echo "  $(GREEN)make check$(NC)         - Check prerequisites"
	@echo "  $(GREEN)make setup$(NC)         - Setup dependencies"
	@echo "  $(GREEN)make install-deps$(NC)  - Install system dependencies"
	@echo "  $(GREEN)make build$(NC)         - Build backend and frontend"
	@echo "  $(GREEN)make dev$(NC)           - Run development servers"
	@echo "  $(GREEN)make docker$(NC)        - Run with Docker Compose"
	@echo "  $(GREEN)make prod$(NC)          - Build production Docker image"
	@echo "  $(GREEN)make test$(NC)          - Run tests"
	@echo "  $(GREEN)make clean$(NC)         - Clean build artifacts"
	@echo ""

check:
	@./run.sh check

setup:
	@./run.sh setup

build:
	@./run.sh build

dev:
	@./run.sh dev

docker:
	@./run.sh docker

prod:
	@./run.sh prod

test:
	@go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -f hdf5-agent
	@rm -rf frontend/dist
	@rm -rf frontend/node_modules
	@go clean
	@echo "Done!"

install-deps:
	@echo "Installing system dependencies..."
	@which brew > /dev/null && brew install hdf5 || \
	(apt-get update && apt-get install -y libhdf5-dev) || \
	(yum install -y hdf5-devel)
	@echo "Done!"

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@cd frontend && npm run lint || true

lint:
	@echo "Linting Go code..."
	@go vet ./...
	@echo "Linting TypeScript/React..."
	@cd frontend && npm run lint || true

run-test-server:
	@go run scripts/create_test_data.go
	@echo "Test data created at: test_data.h5"

.DEFAULT_GOAL := help
