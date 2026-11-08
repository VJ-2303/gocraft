GOFLAGS ?= -trimpath -p 4
STATICCHECK := go run honnef.co/go/tools/cmd/staticcheck@latest
GOVULNCHECK := go run golang.org/x/vuln/cmd/govulncheck@latest
MATRIX_JOBS ?= 4
MATRIX_DIR ?= /tmp/gocraft-matrix
MATRIX_SHARD ?=
MATRIX_SHARDS ?= 4
GATE_DIR ?= /tmp/gocraft-gate

.DEFAULT_GOAL := help

help: ## Print available commands
	@awk 'BEGIN {FS = ":.*?## "; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_\/-]+:.*?## / { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 } END { printf "\n" }' $(MAKEFILE_LIST)

audit: test ## Run tests, verify modules, format check, vet, staticcheck, govulncheck
	go mod verify
	test -z "$$(gofmt -l .)" || (gofmt -l . && echo "code is not gofmt-clean" && exit 1)
	go vet ./...
	$(STATICCHECK) ./...
	$(GOVULNCHECK) ./...

test: ## Run all tests with race detector
	go test $(GOFLAGS) -race ./...

test/cover: ## Run tests with coverage report
	go test $(GOFLAGS) -race -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
	go tool cover -func=/tmp/coverage.out

upgradeable: ## List direct module dependencies with available upgrades
	go list -u -m -f '{{if and (not .Indirect) .Update}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' all

tidy: ## Run go mod tidy, go fix, and gofmt
	go mod tidy -v
	go fix ./...
	gofmt -w .

build: ## Build the gocraft binary
	go build $(GOFLAGS) -o /tmp/bin/gocraft ./cmd/gocraft

run: build ## Build and run the gocraft binary
	/tmp/bin/gocraft

run/live: ## Run gocraft with live reload via air
	@if ! command -v air >/dev/null 2>&1; then \
		echo "air is not installed; install it with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi
	air

gate: ## Generate and verify 13 representative codebases
	@rm -rf $(GATE_DIR)
	@GOCRAFT_GATE_DIR=$(GATE_DIR) go test -count=1 -run TestGateGenerate ./internal/generator/
	@set -e; \
	names=$$(ls -1 $(GATE_DIR)); \
	if [ -z "$$names" ]; then echo "no codebases in $(GATE_DIR)"; exit 1; fi; \
	echo "$$names" | xargs -P $(MATRIX_JOBS) -I{} sh -c 'set -e; \
		name="{}"; \
		echo "== verifying $$name =="; \
		cd "$(GATE_DIR)/{}" && go build ./... && go vet ./... && go test ./...;'; \
	echo "gate: all generated codebases build, vet and test cleanly"

matrix: ## Generate and audit 299 pairwise permutation codebases
	@rm -rf $(MATRIX_DIR)
	@GOCRAFT_MATRIX_DIR=$(MATRIX_DIR) go test -count=1 -run TestMatrixGenerate ./internal/generator/
	@$(MAKE) matrix-audit
	@echo "matrix: all generated codebases are gofmt-clean, build, vet, test cleanly and pass the file and comment audit"

matrix-shard: ## Generate and audit one shard of the pairwise matrix (SHARD=n MATRIX_SHARDS=n)
	@test -n "$(MATRIX_SHARD)" || (echo "MATRIX_SHARD is required (1-based shard number)" && exit 1)
	@test -n "$(MATRIX_SHARDS)" || (echo "MATRIX_SHARDS is required (total shard count)" && exit 1)
	@rm -rf $(MATRIX_DIR)
	@GOCRAFT_MATRIX_DIR=$(MATRIX_DIR) GOCRAFT_MATRIX_SHARD=$(MATRIX_SHARD) GOCRAFT_MATRIX_SHARDS=$(MATRIX_SHARDS) go test -count=1 -run TestMatrixGenerate ./internal/generator/
	@$(MAKE) matrix-audit
	@echo "matrix shard $(MATRIX_SHARD)/$(MATRIX_SHARDS): generated and audited cleanly"

matrix-audit: ## Audit every codebase in MATRIX_DIR (gofmt, stray files, comment hygiene)
	@set -e; \
	names=$$(ls -1 $(MATRIX_DIR)); \
	if [ -z "$$names" ]; then echo "no codebases in $(MATRIX_DIR)"; exit 1; fi; \
	echo "auditing $$(echo "$$names" | grep -c .) codebases in $(MATRIX_DIR)"; \
	echo "$$names" | xargs -P $(MATRIX_JOBS) -I{} sh -c 'set -e; \
		name="{}"; dir="$(MATRIX_DIR)/{}"; \
		echo "== auditing $$name =="; \
		cd "$$dir" && go build ./... && go vet ./... && go test ./...; \
		fmt=$$(gofmt -l . 2>/dev/null); \
		if [ -n "$$fmt" ]; then echo "not gofmt-clean: $$name: $$fmt"; exit 1; fi; \
		strays=$$(find . -name "*.sqlite" -o -name "*.log" -o -name ".air.toml" -o -name "tmp" -o -name "nohup.out" -o -name ".DS_Store" | grep -v certManager || true); \
		if [ -n "$$strays" ]; then echo "stray files in $$name: $$strays"; exit 1; fi; \
		junk=$$(grep -rn "TODO\|FIXME\|XXX\|goes here" . --include="*.go" || true); \
		if [ -n "$$junk" ]; then echo "comment hygiene in $$name: $$junk"; exit 1; fi;'

.PHONY: help audit test test/cover upgradeable tidy build run run/live gate matrix matrix-shard matrix-audit
