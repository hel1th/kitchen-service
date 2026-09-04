.PHONY: build run lint test test-coverage docker-up docker-down migrate-up migrate-down diagrams

APP_NAME=kitchen-service
RESTAURANT_NAME=restaurant-simulator

build:
	go build -o bin/$(APP_NAME) ./cmd/kitchen-service
	go build -o bin/$(RESTAURANT_NAME) ./cmd/restaurant-simulator

run:
	go run ./cmd/kitchen-service

run-restaurant:
	go run ./cmd/restaurant-simulator

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

diagrams:
	java -jar tools/plantuml-mit-1.2026.7.jar docs/diagrams/**/*.puml
