package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hel1th/kitchen-service/internal/app"
	"github.com/hel1th/kitchen-service/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	log.Println("Starting Kitchen Service...")
	if err := application.Run(ctx); err != nil {
		log.Fatalf("service shutdown failed: %v", err)
	}
}
