.PHONY: build run test fmt vet

APP_NAME := nagare-web-server

build:
	go build -o bin/$(APP_NAME) ./cmd/web_server

run:
	go run ./cmd/web_server

test:
	go test ./... -v

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...
