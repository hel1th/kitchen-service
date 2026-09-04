package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	run(ctx)
}

func run(ctx context.Context) {
	fmt.Println("Starting Kitchen Service...")
	<-ctx.Done()
	fmt.Println("Shutting down Kitchen Service gracefully...")
}
