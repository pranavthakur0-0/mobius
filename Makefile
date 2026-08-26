# ==============================================================================
# Mobius Build System
# ==============================================================================

BINARY_NAME=mobius
BUILD_DIR=bin
CMD_DIR=./cmd/mobius

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

.PHONY: all build run test test-race test-coverage fmt vet lint clean tidy help

all: fmt vet lint test build

## build: Compile the mobius CLI binary
build:
	@mkdir -p $(BUILD_DIR)
	@echo "==> Building $(BUILD_DIR)/$(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "==> Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## run: Run Mobius interactive REPL
run:
	@$(GOCMD) run $(CMD_DIR)/main.go

## test: Run unit test suite
test:
	@echo "==> Running tests..."
	$(GOTEST) -v ./...

## test-race: Run tests with race detection enabled
test-race:
	@echo "==> Running tests with race detector..."
	$(GOTEST) -race -v ./...

## test-coverage: Run tests and generate coverage profile
test-coverage:
	@echo "==> Generating coverage profile..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "==> Coverage report written to coverage.html"

## fmt: Format all Go files using standard gofmt
fmt:
	@echo "==> Formatting Go files..."
	@$(GOFMT) -w -s .

## vet: Run go vet on packages
vet:
	@echo "==> Running go vet..."
	$(GOVET) ./...

## lint: Run golangci-lint static analysis
lint:
	@echo "==> Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, running go vet..."; \
		$(GOVET) ./...; \
	fi

## tidy: Tidy and verify Go modules
tidy:
	@echo "==> Tidying go.mod..."
	$(GOMOD) tidy
	$(GOMOD) verify

## clean: Remove binary artifacts and coverage output
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(BINARY_NAME) coverage.out coverage.html .mobius/artifacts/* .mobius/events/*
	@echo "==> Clean complete."

## help: Display this help message
help:
	@echo "Mobius Development Commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
