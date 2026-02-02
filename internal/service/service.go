package service

import (
	"log/slog"

	"github.com/ahxar/go-backend-service/internal/repository"
)

// Service contains business logic and dependencies
type Service struct {
	logger *slog.Logger
	repo   *repository.Repository
}

// New creates a new Service instance
func New(logger *slog.Logger, repo *repository.Repository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}
