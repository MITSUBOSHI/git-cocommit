.PHONY: build test clean fmt lint help install all

all: fmt lint test build

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build    Install cocommit by running install.sh"
	@echo "  test     Run tests"
	@echo "  clean    Remove build artifacts"
	@echo "  fmt      Format the code"
	@echo "  lint     Run linters"
	@echo "  all      Run fmt, lint, test, and build"
	@echo "  help     Show this help message"
	@echo ""

build:
	@echo "Building and installing cocommit..."
	@chmod +x ./install.sh
	@./install.sh

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linters..."
	@go vet ./...
	@if command -v golint > /dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "golint not installed. Skipping."; \
	fi

install: build 