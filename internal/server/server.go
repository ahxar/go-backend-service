package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/safar/go-backend-service/internal/config"
	"github.com/safar/go-backend-service/internal/handler"
	"github.com/safar/go-backend-service/internal/middleware"
)

// New creates and configures the HTTP server
func New(cfg *config.Config, logger *slog.Logger, h *handler.Handler) *http.Server {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /ready", h.Ready)
	mux.HandleFunc("GET /api/example", h.Example)

	// Apply middleware chain: trace ID -> recovery -> logging
	var httpHandler http.Handler = mux
	httpHandler = middleware.Logging(logger)(httpHandler)
	httpHandler = middleware.Recovery(logger)(httpHandler)
	httpHandler = middleware.TraceID(httpHandler)

	// Configure server with explicit timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      httpHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return server
}
