APP=medappoint
PKG=github.com/justanamir/medappoint
PORT?=8080

.PHONY: run tidy build lint test

run:
	cd cmd/server && go run .

tidy:
	go mod tidy

build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$$(date +%Y.%m.%d)-dev" -o bin/$(APP) ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run || true