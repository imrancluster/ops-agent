# Ops Agent: developer commands. Run `make help` to list them.
# CI runs exactly `make check`, so a green `make check` locally means a green CI.

GO      ?= go
PKGS    := ./...
BIN_DIR := bin

.DEFAULT_GOAL := help
.PHONY: help fmt fmt-check vet lint test cover vuln tidy check build run clean

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

fmt: ## Format code (gofmt + goimports)
	golangci-lint fmt

fmt-check: ## Fail if any file is not formatted
	@out="$$(golangci-lint fmt --diff)"; \
	if [ -n "$$out" ]; then echo "$$out"; echo "run 'make fmt'"; exit 1; fi

vet: ## Go's built-in static checks
	$(GO) vet $(PKGS)

lint: ## Run all linters from .golangci.yml
	golangci-lint run $(PKGS)

test: ## Run tests with the race detector and coverage
	$(GO) test -race -cover $(PKGS)

cover: ## Open an HTML coverage report
	$(GO) test -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -html=coverage.out

vuln: ## Scan for known vulnerabilities in code paths we actually call
	govulncheck $(PKGS)

tidy: ## Sync go.mod/go.sum with imports
	$(GO) mod tidy

check: fmt-check vet lint test ## Everything CI runs. Must be green before every commit.

build: ## Build the agent binary into ./bin
	$(GO) build -o $(BIN_DIR)/ops-agent ./cmd/agent

run: ## Run the agent, loading .env if present
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; $(GO) run ./cmd/agent

clean: ## Remove build and coverage output
	rm -rf $(BIN_DIR) coverage.out
