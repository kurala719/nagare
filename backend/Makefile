.PHONY: build run test fmt vet

APP_NAME := nagare-web-server

build:
	go build -tags=jsoniter -o bin/$(APP_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...
