.DEFAULT_GOAL := all

BIN_DIR ?= bin
CLI_BIN ?= $(BIN_DIR)/sorobango
PKG_DIRS ?= ./...

.PHONY: all
all: fmt vet lint test build

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	go build -v -o $(CLI_BIN) ./cmd/sorobango

.PHONY: test
test:
	go test -race -v $(PKG_DIRS)

.PHONY: test-cover
test-cover:
	go test -race -coverprofile=coverage.out -covermode=atomic $(PKG_DIRS)
	go tool cover -func=coverage.out

.PHONY: lint
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not installed; running go vet only" && go vet $(PKG_DIRS))
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || true

.PHONY: fmt
fmt:
	gofmt -w -s .

.PHONY: vet
vet:
	go vet $(PKG_DIRS)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) coverage.out
