package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ahxar/go-backend-service/internal/service"
)

// Handler contains HTTP handlers and dependencies
type Handler struct {
	logger  *slog.Logger
	service *service.Service
}

// New creates a new Handler instance
func New(logger *slog.Logger, svc *service.Service) *Handler {
	return &Handler{
		logger:  logger,
		service: svc,
	}
}

// writeJSON writes a JSON response with the given status code
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we fail to encode, there's not much we can do
		// The status code has already been written
		h.logger.Error("failed to encode response",
			slog.String("error", err.Error()),
		)
	}
}
