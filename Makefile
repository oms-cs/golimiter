.PHONY: build generate test clean install run help

# Variables
BINARY_NAME=golimiter
BUILD_DIR=./cmd/ratelimiter
PROTO_DIR=./api/proto
GEN_DIR=./gen/pb

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build the binary"
	@echo "  generate  - Generate protobuf code"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install dependencies"
	@echo "  run       - Run the service"

# Install dependencies
install:
	go mod download
	go mod tidy

# Generate protobuf code
generate:
	@echo "Generating protobuf code..."
	@mkdir -p $(GEN_DIR)
	protoc --go_out=$(GEN_DIR) --go-grpc_out=$(GEN_DIR) $(PROTO_DIR)/*.proto

# Build the binary
build: generate
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run integration tests (requires Redis)
test-integration:
	@echo "Running integration tests..."
	go test -tags=integration ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -rf $(GEN_DIR)/*

# Run the service
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Development target (install, generate, build)
dev: install generate build
	@echo "Development setup complete"
