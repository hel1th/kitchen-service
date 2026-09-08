package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

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

type TestEnv struct {
	Server    *httptest.Server
	DBPool    *pgxpool.Pool
	Container *postgres.PostgresContainer
}

func SetupE2E(t *testing.T) *TestEnv {
	ctx := context.Background()

	dbName := "kitchen_db"
	dbUser := "postgres"
	dbPassword := "postgres"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run migrations using goose
	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose: failed to set dialect: %v", err)
	}

	if err := goose.Up(db, "../../migrations"); err != nil {
		t.Fatalf("goose: failed to run migrations: %v", err)
	}

	// Wait a bit to ensure migrations are fully committed (rarely needed, but safe)
	time.Sleep(100 * time.Millisecond)

	pool, err := database.NewPool(ctx, connStr)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	// Initialize repositories
	dishRepo := menu_repo.NewDishRepo(pool)
	categoryRepo := menu_repo.NewCategoryRepo(pool)
	restRepo := rest_repo.NewRestaurantRepo(pool)
	cartRepo := order_repo.NewCartRepo(pool)
	orderRepo := order_repo.NewOrderRepo(pool)
	userRepo := order_repo.NewUserRepo(pool)

	// Initialize usecases
	dishUC := menu_uc.NewDishUsecase(dishRepo, categoryRepo)
	categoryUC := menu_uc.NewCategoryUsecase(categoryRepo)
	restUC := rest_uc.NewRestaurantUsecase(restRepo)

	webhookSender := order_infra.NewHTTPWebhookSender(restUC)
	orderUC := order_uc.NewOrderUsecase(cartRepo, orderRepo, dishUC, dishUC, restUC, webhookSender)

	// Initialize API handlers
	menuHandler := menu_api.NewMenuHandler(dishUC, categoryUC)
	orderHandler := order_api.NewOrderHandler(orderUC)
	restHandler := rest_api.NewRestaurantHandler(restUC)

	server := api.NewServer(menuHandler, orderHandler, restHandler)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.ExtractUserID(userRepo))

	gen.HandlerFromMuxWithBaseURL(server, r, "/api/v1")

	ts := httptest.NewServer(r)

	t.Cleanup(func() {
		ts.Close()
	})

	return &TestEnv{
		Server:    ts,
		DBPool:    pool,
		Container: postgresContainer,
	}
}

// Request helpers

func MakeRequest(
	t *testing.T,
	ts *httptest.Server,
	method, path string,
	userID int64,
	body []byte,
) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, ts.URL+"/api/v1"+path, bodyReader)
	require.NoError(t, err)

	if userID > 0 {
		req.Header.Set("X-User-Id", fmt.Sprintf("%d", userID))
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}
