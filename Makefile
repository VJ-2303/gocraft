GOFLAGS ?= -trimpath
STATICCHECK := go run honnef.co/go/tools/cmd/staticcheck@latest
GOVULNCHECK := go run golang.org/x/vuln/cmd/govulncheck@latest

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

run: build ## Build and run gocraft
	/tmp/bin/gocraft

run/live: ## Run gocraft with live reload via air
	@if ! command -v air >/dev/null 2>&1; then \
		echo "air is not installed; install it with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi
	air

gate: ## Generate and verify 13 representative codebases
	@rm -rf /tmp/gocraft-gate
	@GOCRAFT_GATE_DIR=/tmp/gocraft-gate go test -count=1 -run TestGateGenerate ./internal/generator/
	@set -e; for dir in /tmp/gocraft-gate/*/; do \
		echo "== verifying $$(basename $$dir) =="; \
		(cd $$dir && go build ./... && go vet ./... && go test ./...); \
	done; \
	echo "gate: all generated codebases build, vet and test cleanly"

matrix: ## Generate and audit 29 permutation codebases
	@rm -rf /tmp/gocraft-matrix
	@GOCRAFT_MATRIX_DIR=/tmp/gocraft-matrix go test -count=1 -run TestMatrixGenerate ./internal/generator/
	@set -e; for dir in /tmp/gocraft-matrix/*/; do \
		name=$$(basename $$dir); \
		echo "== auditing $$name =="; \
		(cd $$dir && go build ./... && go vet ./... && go test ./...); \
		fmt=$$(cd $$dir && gofmt -l .); \
		if [ -n "$$fmt" ]; then echo "not gofmt-clean: $$name: $$fmt"; exit 1; fi; \
		strays=$$(find $$dir -name "*.sqlite" -o -name "*.log" -o -name ".air.toml" -o -name "tmp" -o -name "nohup.out" -o -name ".DS_Store" | grep -v certManager || true); \
		if [ -n "$$strays" ]; then echo "stray files: $$strays"; exit 1; fi; \
		junk=$$(grep -rn "TODO\|FIXME\|XXX\|goes here" $$dir --include="*.go" || true); \
		if [ -n "$$junk" ]; then echo "comment hygiene: $$junk"; exit 1; fi; \
	done; \
	echo "matrix: all generated codebases are gofmt-clean, build, vet, test cleanly and pass the file and comment audit"

.PHONY: help audit test test/cover upgradeable tidy build run run/live gate matrix
