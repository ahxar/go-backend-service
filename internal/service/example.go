package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/safar/go-backend-service/internal/model"
)

// ExampleService defines business logic for example operations
type ExampleService interface {
	ProcessExample(ctx context.Context, name string) (*model.ExampleResponse, error)
}

// ProcessExample processes an example request with business logic
func (s *Service) ProcessExample(ctx context.Context, name string) (*model.ExampleResponse, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Log with context (includes trace ID from middleware)
	s.logger.InfoContext(ctx, "processing example request",
		slog.String("name", name),
	)

	// Call repository layer for data access
	data, err := s.repo.GetData(ctx, name)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get data",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("data access error: %w", err)
	}

	// Business logic here
	// For demo purposes, we simulate some processing time
	time.Sleep(100 * time.Millisecond)

	// Check context again before returning
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Build response
	response := &model.ExampleResponse{
		Message:   fmt.Sprintf("Hello, %s!", name),
		Timestamp: time.Now().UTC(),
		Processed: true,
	}

	// Log successful processing
	s.logger.InfoContext(ctx, "example request processed",
		slog.String("name", name),
		slog.Any("data", data),
	)

	return response, nil
}
