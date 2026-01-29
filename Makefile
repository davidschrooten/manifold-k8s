.PHONY: build test coverage lint clean install

BINARY_NAME=manifold-k8s
BIN_DIR=bin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) .

test:
	@echo "Running tests..."
	go test -v -race ./...

coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Running linters..."
	go fmt ./...
	go vet ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -f coverage.txt coverage.html

install: build
	@echo "Installing $(BINARY_NAME)..."
	go install .