package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hel1th/kitchen-service/internal/app"
	"github.com/hel1th/kitchen-service/internal/config"
	"github.com/hel1th/kitchen-service/internal/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logg := logger.Setup(cfg.Log.Level)
	slog.SetDefault(logg)

	application, err := app.New(ctx, cfg, logg)
	if err != nil {
		logg.Error("failed to initialize app", slog.Any("error", err))
		os.Exit(1)
	}

	logg.Info("Starting Kitchen Service...")
	if err := application.Run(ctx); err != nil {
		logg.Error("service shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}
}
