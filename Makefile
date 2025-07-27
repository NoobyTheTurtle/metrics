AGENT_DIR = ./cmd/agent
SERVER_DIR = ./cmd/server
AGENT_BIN = $(AGENT_DIR)/agent
SERVER_BIN = $(SERVER_DIR)/server

DATABASE_DSN ?= postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable

.DEFAULT_GOAL := help

.PHONY: test
test:
	@echo "Running unit tests..."
	@go test -count=1 -short ./...

.PHONY: test-all
test-all:
	@echo "Running all tests with database..."
	@go test -count=1 ./...

.PHONY: test-cover
test-cover:
	@echo "Running tests with coverage..."
	@go test -cover -count=1 ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with detailed coverage report..."
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1
	@go tool cover -func=coverage.out
	@rm -f coverage.out

.PHONY: generate
generate:
	@echo "Running go generate..."
	@go generate ./...

.PHONY: generate-mocks
generate-mocks:
	@echo "Regenerating all mocks..."
	@find ./internal -name "interfaces.go" | while read file; do \
		dir=$$(dirname "$$file"); \
		pkg=$$(basename "$$dir"); \
		echo "Generating mock for $$file -> $$dir/mocks.go"; \
		mockgen -source=$$file -destination=$$dir/mocks.go -package=$$pkg; \
	done
	@echo "Mocks successfully regenerated"

.PHONY: format
format:
	@echo "Formatting Go code with goimports..."
	@find . -name "*.go" -not -path "./.history/*" -not -path "./vendor/*" | xargs goimports -w -local github.com/smanhack/metrics
	@echo "Code formatting completed"

.PHONY: godoc
godoc:
	@echo "Starting godoc server at http://localhost:8082/pkg/github.com/NoobyTheTurtle/metrics/?m=all"
	@godoc -http=:8082

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
	@DATABASE_DSN="$(DATABASE_DSN)" $(SERVER_BIN)

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(AGENT_BIN)
	@rm -f $(SERVER_BIN)
	@go clean

.PHONY: postgres
postgres:
	@echo "Starting PostgreSQL in Docker..."
	@docker run --name metrics-postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_DB=metrics \
		-p 5432:5432 \
		-v $(shell pwd)/tmp/postgres-data:/var/lib/postgresql/data \
		-d \
		postgres:17-alpine

.PHONY: postgres-stop
postgres-stop:
	@echo "Stopping PostgreSQL Docker container..."
	@docker stop metrics-postgres
	@docker rm metrics-postgres

.PHONY: profile-base
profile-base:
	@echo "Generating base memory profile..."
	@go run cmd/profile/main.go base

.PHONY: profile-result
profile-result:
	@echo "Generating result memory profile..."
	@go run cmd/profile/main.go result

.PHONY: profile-compare
profile-compare:
	@echo "Comparing memory profiles..."
	@go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

.PHONY: profile-clean
profile-clean:
	@echo "Cleaning profile files..."
	@rm -f profiles/base.pprof
	@rm -f profiles/result.pprof

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make test             - Run unit tests"
	@echo "  make test-all         - Run all tests with database"
	@echo "  make test-cover       - Run tests with coverage report"
	@echo "  make test-coverage    - Get total test coverage percentage"
	@echo "  make generate         - Run go generate"
	@echo "  make generate-mocks   - Regenerate all mocks"
	@echo "  make format           - Format Go code with goimports"
	@echo "  make godoc            - Start godoc web server at http://localhost:8082"
	@echo "  make build-agent      - Build agent"
	@echo "  make build-server     - Build server"
	@echo "  make build-all        - Build all projects"
	@echo "  make run-agent        - Run agent"
	@echo "  make run-server       - Run server"
	@echo "  make postgres         - Start PostgreSQL in Docker"
	@echo "  make postgres-stop    - Stop and remove PostgreSQL Docker container"
	@echo "  make clean            - Clean binary files and reports"
	@echo "  make profile-base     - Generate base memory profile"
	@echo "  make profile-result   - Generate result memory profile"
	@echo "  make profile-compare  - Compare base.pprof vs result.pprof (shows memory diff)"
	@echo "  make profile-clean    - Clean all profile files"