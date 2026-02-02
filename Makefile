.PHONY: help build test lint clean run

# Variables
APP_NAME=server
CMD_PATH=./cmd/server
BUILD_DIR=./bin
VERSION?=$(shell git describe --tags --always --dirty)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: swagger-gen ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -ldflags="-w -s -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

build-all: ## Build for all platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PATH)
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PATH)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "Multi-platform build complete"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w -local github.com/ahxar/go-backend-service .

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean

run: swagger-gen ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run $(CMD_PATH)

dev: swagger-gen ## Run in development mode with air (hot reload)
	@echo "Running in development mode..."
	@air

mod-download: ## Download Go modules
	@echo "Downloading modules..."
	@go mod download

mod-tidy: ## Tidy Go modules
	@echo "Tidying modules..."
	@go mod tidy

mod-verify: ## Verify Go modules
	@echo "Verifying modules..."
	@go mod verify

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/cosmtrek/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

check: lint vet test ## Run all checks

ci: mod-verify check swagger-gen build ## Run CI pipeline locally

swagger-gen: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go -o docs
	@echo "Swagger documentation generated in docs/"

swagger-install: ## Install Swagger CLI tool
	@echo "Installing Swagger CLI tool..."
	@go install github.com/swaggo/swag/cmd/swag@latest

swagger-fmt: ## Format Swagger comments
	@echo "Formatting Swagger comments..."
	@swag fmt

.DEFAULT_GOAL := help
