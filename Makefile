# Makefile for GitHub Contribution City

BINARY_NAME=github-contribution-city
CMD_PATH=cmd/city/main.go
BIN_DIR=bin
DIST_DIR=dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0-dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TIME)

.PHONY: all build clean test bench lint coverage build-all run stop docker-build docker-run help

all: test build

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## Build binary for current platform
	@echo "🔨 Building $(BINARY_NAME) ($(VERSION))..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "✅ Binary built at $(BIN_DIR)/$(BINARY_NAME)"

test: ## Run unit and integration tests
	@echo "🧪 Running unit and integration tests..."
	go test -v -race -coverprofile=coverage.out ./...

bench: ## Run performance benchmark tests
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem ./...

coverage: test ## Generate HTML test coverage report
	@echo "📊 Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated at coverage.html"

lint: ## Run go vet and static checks
	@echo "🔍 Running code verification..."
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "ℹ️  golangci-lint not installed locally, running go fmt check"; \
		test -z "$$(gofmt -l . | grep -v '^vendor/')" || (echo "Unformatted files found:" && gofmt -l . && exit 1); \
	fi

build-all: ## Cross-compile binaries for Linux, macOS, and Windows
	@echo "🌐 Cross-compiling for Linux, macOS, Windows (amd64, arm64)..."
	@mkdir -p $(DIST_DIR)
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_PATH)
	# Darwin / macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe $(CMD_PATH)
	@echo "✅ Cross-compilation complete in $(DIST_DIR)/"

run: build ## Start application in detached server mode
	@./scripts/run.sh

stop: ## Stop background server
	@./scripts/stop.sh

clean: ## Clean binaries and temporary files
	@./scripts/cleanup.sh
	@rm -rf $(DIST_DIR) coverage.out coverage.html

docker-build: ## Build container image
	docker build -t $(BINARY_NAME):latest .

docker-run: ## Run container image
	docker run -d -p 8088:8088 --name $(BINARY_NAME) $(BINARY_NAME):latest
