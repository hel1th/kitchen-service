-include .env
export

.PHONY: all build run run-restaurant lint test test-coverage generate docker-up docker-down migrate-up migrate-down migrate-create diagrams clean

APP_NAME=kitchen-service
RESTAURANT_NAME=restaurant-simulator

DB_DSN ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)
MIGRATIONS_DIR ?= migrations

all: build

build:
	go build -o bin/$(APP_NAME) ./cmd/kitchen-service
	go build -o bin/$(RESTAURANT_NAME) ./cmd/restaurant-simulator

run:
	go run ./cmd/kitchen-service

run-restaurant:
	go run ./cmd/restaurant-simulator

generate:
	oapi-codegen -config api/openapi/oapi-codegen.yaml api/openapi/openapi.yaml

lint:
	golangci-lint run ./...

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

diagrams:
	java -DRELATIVE_INCLUDE="." -jar tools/plantuml-mit-1.2026.7.jar -r "docs/diagrams/**.puml"

clean:
	rm -rf bin/ coverage.out coverage.html
