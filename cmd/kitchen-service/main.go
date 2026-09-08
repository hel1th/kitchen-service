package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/hel1th/kitchen-service/internal/api"
	"github.com/hel1th/kitchen-service/internal/api/gen"
	menu_api "github.com/hel1th/kitchen-service/internal/module/menu/api"
	menu_repo "github.com/hel1th/kitchen-service/internal/module/menu/repository"
	menu_uc "github.com/hel1th/kitchen-service/internal/module/menu/usecase"
	order_api "github.com/hel1th/kitchen-service/internal/module/order/api"
	order_infra "github.com/hel1th/kitchen-service/internal/module/order/infra"
	order_repo "github.com/hel1th/kitchen-service/internal/module/order/repository"
	order_uc "github.com/hel1th/kitchen-service/internal/module/order/usecase"
	rest_api "github.com/hel1th/kitchen-service/internal/module/restaurant/api"
	rest_repo "github.com/hel1th/kitchen-service/internal/module/restaurant/repository"
	rest_uc "github.com/hel1th/kitchen-service/internal/module/restaurant/usecase"
	"github.com/hel1th/kitchen-service/internal/shared/database"
	"github.com/hel1th/kitchen-service/internal/shared/middleware"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatalf("service startup failed: %v", err)
	}
}

func run(ctx context.Context) error {
	log.Println("Starting Kitchen Service...")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSLMODE"),
	)

	pool, err := database.NewPool(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Repositories
	dishRepo := menu_repo.NewDishRepo(pool)
	categoryRepo := menu_repo.NewCategoryRepo(pool)
	restRepo := rest_repo.NewRestaurantRepo(pool)
	cartRepo := order_repo.NewCartRepo(pool)
	orderRepo := order_repo.NewOrderRepo(pool)
	userRepo := order_repo.NewUserRepo(pool)

	// Usecases
	dishUC := menu_uc.NewDishUsecase(dishRepo, categoryRepo)
	categoryUC := menu_uc.NewCategoryUsecase(categoryRepo)
	restUC := rest_uc.NewRestaurantUsecase(restRepo)

	webhookSender := order_infra.NewHTTPWebhookSender(restUC)
	orderUC := order_uc.NewOrderUsecase(cartRepo, orderRepo, dishUC, dishUC, restUC, webhookSender)

	// API Handlers
	menuHandler := menu_api.NewMenuHandler(dishUC, categoryUC)
	orderHandler := order_api.NewOrderHandler(orderUC)
	restHandler := rest_api.NewRestaurantHandler(restUC)

	// API Composition Root
	server := api.NewServer(menuHandler, orderHandler, restHandler)

	// Router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Auth middleware
	r.Use(middleware.ExtractUserID(userRepo))

	// Register API v1 handlers
	gen.HandlerFromMuxWithBaseURL(server, r, "/api/v1")

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HTTP_HOST")

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: r,
	}

	go func() {
		log.Printf("Listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Kitchen Service gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
