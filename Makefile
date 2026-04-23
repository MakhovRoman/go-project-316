CMD_DIR = ./cmd/hexlet-go-crawler
BINARY_NAME = hexlet-go-crawler

.PHONY: build test run lint dev

lint:
	golangci-lint run

test:
	go test ./...

build: lint test
	go build -o bin/${BINARY_NAME} ${CMD_DIR}

run:
	go run ${CMD_DIR}/main.go $(URL)

