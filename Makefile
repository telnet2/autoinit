.PHONY: help test test-race test-cover bench lint fmt vet staticcheck gosec clean build examples deps

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Test targets
test: ## Run all tests
	go test -v ./...

test-race: ## Run tests with race detector
	go test -race ./...

test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests
	go test -tags=integration ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Code quality targets
lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

staticcheck: ## Run staticcheck
	staticcheck ./...

gosec: ## Run gosec security scanner
	gosec ./...

# Build targets
build: ## Build the module
	go build -v ./...

examples: ## Build all examples
	@if [ -d "examples" ]; then \
		for dir in examples/*/; do \
			if [ -f "$$dir/main.go" ]; then \
				echo "Building example in $$dir"; \
				cd "$$dir" && go build -v . && cd - > /dev/null; \
			fi \
		done \
	fi

# Dependency targets
deps: ## Download and verify dependencies
	go mod download
	go mod verify

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

# Development targets
dev-setup: ## Set up development environment
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/cmd/goimports@latest

# CI simulation
ci: deps vet staticcheck lint test-race test-cover bench build examples ## Run all CI checks locally

# Cleanup
clean: ## Clean build artifacts
	go clean ./...
	rm -f coverage.out coverage.html

# Documentation
docs: ## Generate documentation
	go doc -all > docs/api.txt

# Release preparation
pre-commit: fmt vet lint test ## Run pre-commit checks
	@echo "✅ Pre-commit checks passed"

# Git hooks
install-hooks: ## Install git hooks
	@echo "#!/bin/sh" > .git/hooks/pre-commit
	@echo "make pre-commit" >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed"