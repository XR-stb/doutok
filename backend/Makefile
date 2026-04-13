.PHONY: all build gateway run test lint clean docker-up docker-down migrate tidy

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_DIR=bin

all: tidy build

tidy:
	$(GOCMD) mod tidy

build: gateway

gateway:
	$(GOBUILD) -o $(BINARY_DIR)/gateway ./cmd/gateway

run: gateway
	./$(BINARY_DIR)/gateway

test:
	$(GOTEST) -v -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BINARY_DIR)

# Docker
docker-up:
	cd ../deploy && docker compose up -d

docker-down:
	cd ../deploy && docker compose down

# Database migration
migrate:
	mysql -h127.0.0.1 -udoutok -pdoutok123 doutok < migration/001_init.sql

# Generate protobuf
proto:
	protoc --go_out=. --go-grpc_out=. ../api/proto/**/*.proto
