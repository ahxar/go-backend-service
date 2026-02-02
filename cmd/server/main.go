package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/safar/go-backend-service/internal/config"
	"github.com/safar/go-backend-service/internal/handler"
	"github.com/safar/go-backend-service/internal/repository"
	"github.com/safar/go-backend-service/internal/server"
	"github.com/safar/go-backend-service/internal/service"
	"github.com/safar/go-backend-service/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.Environment, cfg.LogLevel)

	log.Info("starting server",
		slog.String("port", cfg.Port),
		slog.String("environment", cfg.Environment),
		slog.String("log_level", cfg.LogLevel),
	)

	// Initialize repository layer
	// In a real app, this would include database connections
	repo := repository.New(log)

	// Initialize service layer
	svc := service.New(log, repo)

	// Initialize handler layer
	h := handler.New(log, svc)

	// Create and configure HTTP server
	srv := server.New(cfg, log, h)

	// Create signal context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	go func() {
		log.Info("server listening",
			slog.String("address", srv.Addr),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}()

	// Block until shutdown signal received
	<-ctx.Done()

	log.Info("shutdown signal received, starting graceful shutdown")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}
