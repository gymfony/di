.PHONY: test test-cover lint fmt check

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
