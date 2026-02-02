.PHONY: help test lint clean build build-dev build-linux run docker-build swagger install-swag deps env run-local run-dev run-prod run-watch format check

# --- Project Variables ---
BINARY_NAME := compressor
CMD_API_PATH := ./cmd/api/main.go

# --- Build Variables ---
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date +%FT%T%z)

LDFLAGS := -ldflags "-X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)"
GO_BUILD_FLAGS := -trimpath

# --- Docker Variables ---
DOCKER_TAG := $(COMMIT_HASH)

help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mUsage:\033[0m\n make \033[36m<target>\033[0m\n"} \
	/^[a-zA-Z0-9_-]+:.*?##/ { printf " \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
	/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Environment

deps: ## Install system dependencies (libvips) for Ubuntu/WSL
	@echo "Installing system dependencies..."
	sudo apt update
	sudo apt install -y build-essential libvips libvips-dev pkg-config

env: ## Create .env file from example
	@cp -n .env.example .env || true

##@ Development

run-local: ## Run app in local mode with .env
	@cd $(shell pwd) && ENV_PATH=.env APP_ENV=local go run ./cmd/api/main.go -env-path=.env

run-dev: ## Run app in dev mode
	@go run $(CMD_API_PATH) -env=dev -env-path=.env

run-prod: ## Run app in prod mode
	@go run $(CMD_API_PATH) -env=prod -env-path=.env

run-watch: ## Run with live reload
	@air -c .air.toml

##@ Testing & Quality

test: ## Run unit tests
	@go test -v -race ./...

lint: ## Run golangci-lint
	@golangci-lint run ./...

audit: ## Run vulnerability check
	@go list -u -m all
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

##@ Builds

# ATTENTION! bimg need CGO_ENABLED=1

build: ## Build binary for current OS (default: LOCAL config)
	@echo "Building $(BINARY_NAME) with LOCAL config..."
	@CGO_ENABLED=1 APP_ENV=local \
		go build $(GO_BUILD_FLAGS) $(LDFLAGS) \
		-o bin/$(BINARY_NAME) $(CMD_API_PATH)
	@echo "Build complete: bin/$(BINARY_NAME) (APP_ENV=local → config.local.yaml)"

build-local: ## Explicit local build (alias, LOCAL config)
	@$(MAKE) build

build-dev: ## Build binary for DEV config
	@echo "Building $(BINARY_NAME)-dev with DEV config..."
	@CGO_ENABLED=1 APP_ENV=dev \
		go build $(GO_BUILD_FLAGS) $(LDFLAGS) \
		-o bin/$(BINARY_NAME)-dev $(CMD_API_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)-dev (APP_ENV=dev → config.dev.yaml)"

build-prod: ## Build binary for PROD config
	@echo "Building $(BINARY_NAME)-prod with PROD config..."
	@CGO_ENABLED=1 APP_ENV=prod \
		go build $(GO_BUILD_FLAGS) $(LDFLAGS) \
		-o bin/$(BINARY_NAME)-prod $(CMD_API_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)-prod (APP_ENV=prod → config.prod.yaml)"

clean: ## Remove build artifacts
	@rm -rf bin/

##@ Deployment

docker-build: ## Build docker image (PROD, APP_ENV=prod inside image)
	@echo "Building Docker image $(BINARY_NAME):$(DOCKER_TAG) with PROD config..."
	@docker build \
		-t $(BINARY_NAME):$(DOCKER_TAG) \
		-t $(BINARY_NAME):latest \
		-f deployments/Dockerfile .
##@ Quality Control

format: ## Format code
	@go fmt ./...

check: format lint test audit ## Run all checks before commit
