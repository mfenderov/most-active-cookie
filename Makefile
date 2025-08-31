.PHONY: all build test test-all clean mocks lint fmt install release publish

# Repository configuration
NAME := most-active-cookie
REPO := github.com/mfenderov/${NAME}

# Binary configuration
BINARY_NAME := most-active-cookie
CLI_PATH := ./cmd/most-active-cookie
BUILD_DIR := ./build

# Build flags
BUILD_FLAGS := -ldflags "-s -w"

## all: Quality checks + build (default workflow)
all: lint test build

## build: Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CLI_PATH)

## mocks: Generate mocks using mockery
mocks:
	go tool mockery

## fmt: Auto-fix formatting issues (gofmt, goimports, unused params, etc.)
fmt:
	go tool golangci-lint run --fix

## lint: Run static analysis checks with golangci-lint
lint:
	go tool golangci-lint run

## test: Run unit tests (fast feedback)
test-unit: mocks
	go test -v -race ./src/...

test-integration:
	go test -v ./integration-tests/...
	go test -bench=. -benchmem ./integration-tests/...

## test-all: Run comprehensive tests (unit + integration + performance + benchmarks)
test: mocks test-unit test-integration

## install: Install CLI tool via go install
install:
	@echo "📦 Installing CLI tool..."
	@go install $(REPO)/cmd/$(BINARY_NAME)
	@echo "✅ CLI installed: $(BINARY_NAME)"

## release: Trigger Go module proxy indexing (requires GITHUB_REF_NAME)
release:
	@if [ -z "$(GITHUB_REF_NAME)" ]; then \
		echo "❌ GITHUB_REF_NAME not set. Use: make release GITHUB_REF_NAME=v1.0.0"; \
		exit 1; \
	fi
	@echo "🚀 Triggering Go proxy for $(REPO)@$(GITHUB_REF_NAME)..."
	@curl -f "https://proxy.golang.org/$(REPO)/@v/$(GITHUB_REF_NAME).info" || echo "Module will be available after first go get"
	@echo "✅ Module $(REPO)@$(GITHUB_REF_NAME) is ready for: go get $(REPO)@$(GITHUB_REF_NAME)"

## publish: Complete publishing workflow (build + test + release)
publish: lint test build release
	@echo "🎉 Complete publishing workflow completed!"

## clean: Remove build artifacts and generated files
clean:
	go clean
	rm -rf $(BUILD_DIR)
	rm -f src/*/*_mock.go
