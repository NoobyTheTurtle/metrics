AGENT_DIR = ./cmd/agent
SERVER_DIR = ./cmd/server
AGENT_BIN = $(AGENT_DIR)/agent
SERVER_BIN = $(SERVER_DIR)/server

.DEFAULT_GOAL := help

.PHONY: test
test:
	@echo "Running tests without cache..."
	@go test -v -count=1 ./...

.PHONY: test-cover
test-cover:
	@echo "Running tests with coverage..."
	@go test -cover -count=1 ./...

.PHONY: build-agent
build-agent:
	@echo "Building agent..."
	@go build -o $(AGENT_BIN) $(AGENT_DIR)
	@echo "Binary created at: $(AGENT_BIN)"

.PHONY: build-server
build-server:
	@echo "Building server..."
	@go build -o $(SERVER_BIN) $(SERVER_DIR)
	@echo "Binary created at: $(SERVER_BIN)"

.PHONY: build-all
build-all: build-agent build-server
	@echo "All projects built"

.PHONY: run-agent
run-agent: build-agent
	@echo "Running agent..."
	@$(AGENT_BIN)

.PHONY: run-server
run-server: build-server
	@echo "Running server..."
	@$(SERVER_BIN)

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(AGENT_BIN)
	@rm -f $(SERVER_BIN)
	@go clean

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make test         - Run tests without cache"
	@echo "  make test-cover   - Run tests with coverage report"
	@echo "  make build-agent  - Build agent"
	@echo "  make build-server - Build server"
	@echo "  make build-all    - Build all projects"
	@echo "  make run-agent    - Run agent"
	@echo "  make run-server   - Run server"
	@echo "  make clean        - Clean binary files and reports"