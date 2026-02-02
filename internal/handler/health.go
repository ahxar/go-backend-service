package handler

import (
	"log/slog"
	"net/http"

	"github.com/ahxar/go-backend-service/internal/model"
)

// Health handles health check requests
// @Summary Health check
// @Description Check if the service is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Failure 503 {object} model.ErrorResponse
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.service.CheckHealth(ctx); err != nil {
		h.logger.ErrorContext(ctx, "health check failed",
			slog.String("error", err.Error()),
		)
		h.writeJSON(w, http.StatusServiceUnavailable, &model.ErrorResponse{
			Error: "service unhealthy",
		})
		return
	}

	h.writeJSON(w, http.StatusOK, &model.HealthResponse{
		Status: "healthy",
	})
}

// Ready handles readiness check requests
// @Summary Readiness check
// @Description Check if the service is ready to handle traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} model.ReadyResponse
// @Failure 503 {object} model.ErrorResponse
// @Router /ready [get]
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.service.CheckReady(ctx); err != nil {
		h.logger.ErrorContext(ctx, "readiness check failed",
			slog.String("error", err.Error()),
		)
		h.writeJSON(w, http.StatusServiceUnavailable, &model.ErrorResponse{
			Error: "service not ready",
		})
		return
	}

	h.writeJSON(w, http.StatusOK, &model.ReadyResponse{
		Status: "ready",
	})
}
