.PHONY: build test lint fmt run swagger deps clean docker-build help

BINARY_NAME := server
BUILD_DIR   := bin
GO          := go
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -X main.Version=$(VERSION)

# Build the server binary
build:
	$(GO) build -v -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run tests with race detector and coverage
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run linters
lint:
	golangci-lint run ./...

# Format code
fmt:
	gofmt -w .
	goimports -w .

# Run the server locally
run:
	$(GO) run -ldflags "$(LDFLAGS)" ./cmd/server -config configs/config.yaml

# Regenerate swagger docs into docs/
swagger:
	$(GO) run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g cmd/server/main.go -o docs --parseInternal

# Download and tidy dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out

# Build the docker image
docker-build:
	docker build -t golang-ddd-template:latest .

help:
	@echo "Available targets:"
	@echo "  build         - Build the server binary"
	@echo "  test          - Run tests with -race and coverage"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format code"
	@echo "  run           - Run the server locally"
	@echo "  swagger       - Regenerate swagger docs"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build the docker image"
