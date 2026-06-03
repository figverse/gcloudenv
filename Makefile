BINARY := gcloudenv
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/figverse/gcloudenv/cmd.version=$(VERSION)

.DEFAULT_GOAL := build

.PHONY: build
build: ## Build the binary into ./$(BINARY)
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) .

.PHONY: install
install: ## Install the binary into GOBIN
	go install -ldflags '$(LDFLAGS)' .

.PHONY: test
test: ## Run unit tests
	go test ./...

.PHONY: cover
cover: ## Run tests with a coverage summary
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

.PHONY: fmt
fmt: ## Format the code
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any file is not gofmt-clean
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "not gofmt-clean:"; echo "$$out"; exit 1; fi

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (must be installed)
	golangci-lint run

.PHONY: check
check: fmt-check vet lint test ## Run all CI checks locally

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
