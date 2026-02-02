package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/ahxar/go-backend-service/internal/repository"
)

func setupTestService() *Service {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	repo := repository.New(logger)
	return New(logger, repo)
}

func TestProcessExample(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	result, err := svc.ProcessExample(ctx, "Test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Message != "Hello, Test!" {
		t.Errorf("expected message 'Hello, Test!', got %s", result.Message)
	}

	if !result.Processed {
		t.Error("expected processed to be true")
	}
}

func TestProcessExample_ContextCancellation(t *testing.T) {
	svc := setupTestService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.ProcessExample(ctx, "Test")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestProcessExample_ContextTimeout(t *testing.T) {
	svc := setupTestService()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := svc.ProcessExample(ctx, "Test")
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestCheckHealth(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	if err := svc.CheckHealth(ctx); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckReady(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	if err := svc.CheckReady(ctx); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
