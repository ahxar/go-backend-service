package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/safar/go-backend-service/internal/model"
	"github.com/safar/go-backend-service/internal/repository"
	"github.com/safar/go-backend-service/internal/service"
)

func setupTestHandler() *Handler {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	repo := repository.New(logger)
	svc := service.New(logger, repo)
	return New(logger, svc)
}

func TestHealth(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response model.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", response.Status)
	}
}

func TestReady(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/ready", http.NoBody)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response model.ReadyResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Errorf("expected status ready, got %s", response.Status)
	}
}

func TestExample(t *testing.T) {
	h := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/example?name=Test", http.NoBody)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	h.Example(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response model.ExampleResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "Hello, Test!" {
		t.Errorf("expected message 'Hello, Test!', got %s", response.Message)
	}

	if !response.Processed {
		t.Error("expected processed to be true")
	}
}
