package handler

import (
	"log/slog"
	"net/http"

	"github.com/ahxar/go-backend-service/internal/model"
)

// Example handles example API requests
// @Summary Example endpoint
// @Description A sample endpoint demonstrating the full request lifecycle
// @Tags example
// @Accept json
// @Produce json
// @Param name query string false "Name to greet" default(World)
// @Success 200 {object} model.ExampleResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/example [get]
func (h *Handler) Example(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract query parameter
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}

	// Call service layer
	result, err := h.service.ProcessExample(ctx, name)
	if err != nil {
		h.logger.ErrorContext(ctx, "service error",
			slog.String("error", err.Error()),
			slog.String("name", name),
		)
		h.writeJSON(w, http.StatusInternalServerError, &model.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	// Return successful response
	h.writeJSON(w, http.StatusOK, result)
}
