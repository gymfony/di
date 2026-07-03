.PHONY: test test-cover lint fmt check

.DEFAULT_GOAL := help

test:
	go test -race -count=1 ./...

test-cover:
	go test -count=1 -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

lint:
	golangci-lint run --timeout=5m

fmt:
	go fmt ./...

check:
	go vet ./...
	golangci-lint run --timeout=5m
	go test -race -count=1 ./...

help:
	@echo "Available commands:"
	@echo "  make test       - running tests"
	@echo "  make test-cover - running linter"
	@echo "  make fmt        - running automatically formats Go source code according to the official community standards"
	@echo "  make check      - running tests and linter"
