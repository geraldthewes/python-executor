.PHONY: help build build-server build-cli test test-unit test-integration lint clean docker-build docker-push run-server install-tools

# Build configuration
BINARY_SERVER := bin/python-executor-server
BINARY_CLI := bin/python-executor
VERSION := v0.4
DOCKER_IMAGE := registry.cluster:5000/python-executor:$(VERSION)
DOCKER_IMAGE_LATEST := registry.cluster:5000/python-executor:latest
GO_BUILD_FLAGS := -ldflags="-s -w"
CGO_ENABLED := 0

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: build-server build-cli ## Build both server and CLI

build-server: ## Build the API server
	@echo "Building server..."
	@mkdir -p bin
	CGO_ENABLED=$(CGO_ENABLED) go build $(GO_BUILD_FLAGS) -o $(BINARY_SERVER) ./cmd/server

build-cli: ## Build the CLI tool
	@echo "Building CLI..."
	@mkdir -p bin
	CGO_ENABLED=$(CGO_ENABLED) go build $(GO_BUILD_FLAGS) -o $(BINARY_CLI) ./cmd/python-executor

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/integration/...

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.txt

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) -f Dockerfile .
	@echo "Tagging Docker image as latest..."
	docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE_LATEST)

docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image to registry..."
	docker push $(DOCKER_IMAGE)
	@echo "Pushing latest tagged Docker image..."
	docker push $(DOCKER_IMAGE_LATEST)


nomad-restart:
	@echo "Restarting Nomad service..."
	@nomad job restart -yes python-executor
	@echo "Nomad service restarted successfully"


run-server: build-server ## Run the server locally
	@echo "Starting server..."
	$(BINARY_SERVER)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.DEFAULT_GOAL := help
