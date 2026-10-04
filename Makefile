# scarabeo - Makefile
# Run "make" or "make help" to list every target.

GO        ?= go
BINARY    ?= scarabeo
BIN_DIR   ?= bin
CMD       := .
DICT      ?= dicts/italia-1a
RACK      ?= 8
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)
BENCHTIME ?= 1s

# Defaults for the run-* targets
CARET     ?= asmmquosd
CROSSING  ?= a
TOP_N     ?= 10
PLAYERS   ?= 2
SEED      ?= 42
MAX_TURNS ?= 0

.DEFAULT_GOAL := help
.PHONY: help all build build-all install version run run-solve run-play run-top \
        serve fmt fmt-check vet lint tidy test test-race test-verbose test-cover \
        bench bench-cpu bench-mem clean

help: ## Show this help
	@echo "scarabeo - available targets:"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

## Build

all: fmt-check vet test build ## Format check, vet, test and build

build: ## Build the binary into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(CMD)

build-all: ## Cross-compile linux/amd64, darwin/arm64 and windows/amd64
	@mkdir -p $(BIN_DIR)
	GOOS=linux   GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-linux-amd64 $(CMD)
	GOOS=darwin  GOARCH=arm64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-darwin-arm64 $(CMD)
	GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-windows-amd64.exe $(CMD)

install: ## Install the binary into GOPATH/bin
	$(GO) install -trimpath -ldflags "$(LDFLAGS)" $(CMD)

version: ## Print the version injected at build time
	@echo $(VERSION)

## Run

run: run-solve ## Alias for run-solve

run-solve: build ## Solve a rack (CARET=... CROSSING=...)
	$(BIN_DIR)/$(BINARY) solve -dict $(DICT) -rack-size $(RACK) -crossing "$(CROSSING)" $(CARET)

run-play: build ## Simulate a game (PLAYERS=... SEED=... MAX_TURNS=...)
	$(BIN_DIR)/$(BINARY) play -dict $(DICT) -rack-size $(RACK) -players $(PLAYERS) -seed $(SEED) -max-turns $(MAX_TURNS)

run-top: build ## Show the best words (TOP_N=...)
	$(BIN_DIR)/$(BINARY) top -dict $(DICT) -rack-size $(RACK) -n $(TOP_N)

serve: ## Serve the HTML board page at http://localhost:8080
	$(GO) run ./cmd/webserve -addr :8080 -dir web

## Quality

fmt: ## Format all Go sources with gofmt
	gofmt -w .

fmt-check: ## Fail if any Go source is not gofmt-clean
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then echo "not gofmt-clean:"; echo "$$out"; exit 1; fi; \
	echo "gofmt clean"

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run golangci-lint (must be installed)
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint is not installed"; exit 1; }
	golangci-lint run

tidy: ## Tidy go.mod and go.sum
	$(GO) mod tidy

## Test

test: ## Run the test suite
	$(GO) test ./...

test-race: ## Run tests with the race detector
	$(GO) test -race ./...

test-verbose: ## Run tests verbosely
	$(GO) test -v ./...

test-cover: ## Run tests and print a coverage summary
	$(GO) test -covermode=atomic -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	@echo "HTML report: go tool cover -html=coverage.out"

## Benchmark

bench: ## Run all benchmarks
	$(GO) test -run '^$$' -bench . -benchmem -benchtime $(BENCHTIME) ./...

bench-cpu: ## Benchmark the engine with a CPU profile
	$(GO) test -run '^$$' -bench . -benchmem -benchtime $(BENCHTIME) -cpuprofile=cpu.prof ./internal/engine
	@echo "Profile: go tool pprof cpu.prof"

bench-mem: ## Benchmark the engine with a memory profile
	$(GO) test -run '^$$' -bench . -benchmem -benchtime $(BENCHTIME) -memprofile=mem.prof ./internal/engine
	@echo "Profile: go tool pprof mem.prof"

## Clean

clean: ## Remove build artifacts, coverage and profiles
	rm -rf $(BIN_DIR) coverage.out coverage.html cpu.prof mem.prof
