.PHONY: build test lint clean tidy test-e2e

# Build targets
build:
	go build -v ./...

# Test targets
test:
	go test -v ./...

test-e2e:
	DG_E2E=1 OPENAI_MOCK=1 go test -v -tags=integration ./test/e2e/...

# Code quality
lint:
	go vet ./...
	go fmt ./...

# Dependency management
tidy:
	go mod tidy

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out

# Default target
all: lint test build 