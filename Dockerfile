FROM golang:1.27.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/kitchen-service ./cmd/kitchen-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/restaurant-simulator ./cmd/restaurant-simulator

FROM alpine:3.20 AS kitchen-service
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/kitchen-service /app/kitchen-service
EXPOSE 8080
ENTRYPOINT ["/app/kitchen-service"]

FROM alpine:3.20 AS restaurant-simulator
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/restaurant-simulator /app/restaurant-simulator
EXPOSE 8081
ENTRYPOINT ["/app/restaurant-simulator"]
