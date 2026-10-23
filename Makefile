# Command-line and environment variable settings for the admin commands.
GOFLAGS ?= -trimpath
STATICCHECK := go run honnef.co/go/tools/cmd/staticcheck@latest
GOVULNCHECK := go run golang.org/x/vuln/cmd/govulncheck@latest

.DEFAULT_GOAL := help

## help: Print a list of available commands with their descriptions.
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_\/-]+:.*?##/ { printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## audit: Run all tests, verify modules, check .go file formatting, and run go vet, staticcheck and govulncheck.
audit: test
	go mod verify
	test -z "$$(gofmt -l .)" || (gofmt -l . && echo "code is not gofmt-clean" && exit 1)
	go vet ./...
	$(STATICCHECK) ./...
	$(GOVULNCHECK) ./...

## test: Run all tests, with race detection and VCS stamping enabled.
test:
	go test $(GOFLAGS) -race ./...

## test/cover: Run all tests with coverage, save results to /tmp/coverage.out, and display an HTML report.
test/cover:
	go test $(GOFLAGS) -race -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
	go tool cover -func=/tmp/coverage.out

## upgradeable: List direct module dependencies that have upgrades available and optionally update them.
upgradeable:
	go list -u -m -f '{{if and (not .Indirect) .Update}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' all

## tidy: Run go mod tidy -v, apply go fix, and format all Go files with go fmt.
tidy:
	go mod tidy -v
	go fix ./...
	gofmt -w .

## build: Build the nomoreboiler application and output the binary to /tmp/bin.
build:
	go build $(GOFLAGS) -o /tmp/bin/nomoreboiler ./cmd/nomoreboiler

## run: Build and run the nomoreboiler application.
run: build
	/tmp/bin/nomoreboiler

## run/live: Build and run the nomoreboiler application with live reloading, rebuilding on file changes.
run/live:
	@if ! command -v air >/dev/null 2>&1; then \
		echo "air is not installed; install it with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi
	air

## gate: Generate example codebases to /tmp/nmb-gate and build and vet each one.
gate:
	@rm -rf /tmp/nmb-gate
	@NMB_GATE_DIR=/tmp/nmb-gate go test -count=1 -run TestGateGenerate ./internal/generator/
	@set -e; for dir in /tmp/nmb-gate/*/; do \
		echo "== verifying $$(basename $$dir) =="; \
		(cd $$dir && go build ./... && go vet ./... && go test ./...); \
	done; \
	echo "gate: all generated codebases build, vet and test cleanly"

.PHONY: help audit test test/cover upgradeable tidy build run run/live gate matrix

## matrix: Generate the full permutation matrix to /tmp/nmb-matrix and build, vet and test every codebase.
matrix:
	@rm -rf /tmp/nmb-matrix
	@NMB_MATRIX_DIR=/tmp/nmb-matrix go test -count=1 -run TestMatrixGenerate ./internal/generator/
	@set -e; for dir in /tmp/nmb-matrix/*/; do \
		echo "== verifying $$(basename $$dir) =="; \
		(cd $$dir && go build ./... && go vet ./... && go test ./...); \
	done; \
	echo "matrix: all generated codebases build, vet and test cleanly"
