.PHONY: build run test clean deps tidy lint swagger

# Binary name
BINARY_NAME=auth-service

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLIST=$(GOCMD) list
GOLINT=golangci-lint

# Build
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/api/

# Run the application
run:
	$(GOCMD) run ./cmd/api

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

# Download dependencies
deps:
	$(GOMOD) download

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Run linter
lint:
	$(GOLINT) run

# Generate swagger docs
swagger:
	swag init -g cmd/api/main.go

# Install development dependencies
install-deps:
	# Install golangci-lint if not installed
	if ! command -v golangci-lint &> /dev/null; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2; \
	fi
	# Install swag if not installed
	if ! command -v swag &> /dev/null; then \
		$(GOGET) -u github.com/swaggo/swag/cmd/swag@latest; \
	fi

# Help
help:
	@echo "Available commands:"
	@echo "  build     - Build the application"
	@echo "  run       - Run the application"
	@echo "  test      - Run tests"
	@echo "  clean     - Remove build artifacts"
	@echo "  deps      - Download dependencies"
	@echo "  tidy      - Clean up dependencies"
	@echo "  lint      - Run linter"
	@echo "  swagger   - Generate Swagger documentation"
	@echo "  install-deps - Install development dependencies"

.DEFAULT_GOAL := help
