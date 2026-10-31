package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-crud-backend/internal/config"
	"go-crud-backend/internal/database"
	"go-crud-backend/internal/handler"
	"go-crud-backend/internal/repository"
	"go-crud-backend/internal/service"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting application on port %s...", cfg.Port)

	// 1. Initialize PostgreSQL
	pgDB, err := database.NewPostgres(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()
	log.Println("Connected to PostgreSQL successfully")

	// Run auto migrations
	if err := database.MigratePostgres(pgDB); err != nil {
		log.Fatalf("Failed to execute database migrations: %v", err)
	}
	log.Println("Database migrations applied")

	// 2. Initialize Redis
	redisClient, err := database.NewRedis(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis successfully")

	// 3. Initialize layers
	repo := repository.NewPostgresProductRepository(pgDB)
	svc := service.NewProductService(repo, redisClient, cfg.CacheTTL)
	productHandler := handler.NewProductHandler(svc, pgDB, redisClient)

	// 4. Register HTTP routes
	mux := http.NewServeMux()
	productHandler.RegisterRoutes(mux)

	// Apply middlewares: CORS -> Logging -> Recover -> Mux
	var appHandler http.Handler = mux
	appHandler = handler.LoggingMiddleware(appHandler)
	appHandler = handler.RecoverMiddleware(appHandler)
	appHandler = handler.CORSMiddleware(appHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      appHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Graceful shutdown setup
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Signal %v received, initiating graceful shutdown...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Could not stop server gracefully: %v", err)
			if err := srv.Close(); err != nil {
				log.Fatalf("Could not force close server: %v", err)
			}
		}
		log.Println("Server gracefully stopped")
	}
}
