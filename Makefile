APP=medappoint
PORT?=8080
VERSION?=dev

.PHONY: run tidy build test lint

run:
	go run ./cmd/server

tidy:
	go mod tidy

# Cross-platform friendly: pass a VERSION when you want a custom stamp, e.g.:
#   make build VERSION=2025.08.18-dev
build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/$(APP) ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run || true
