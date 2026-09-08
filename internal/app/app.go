package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hel1th/kitchen-service/internal/api"
	"github.com/hel1th/kitchen-service/internal/api/gen"
	"github.com/hel1th/kitchen-service/internal/config"
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

type App struct {
	cfg  *config.Config
	pool *pgxpool.Pool
	srv  *http.Server
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	pool, err := database.NewPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

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

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: r,
	}

	return &App{
		cfg:  cfg,
		pool: pool,
		srv:  srv,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.pool.Close()

	go func() {
		log.Printf("Listening on %s", a.srv.Addr)
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Kitchen Service gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
