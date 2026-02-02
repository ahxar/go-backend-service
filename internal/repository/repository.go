package repository

import (
	"log/slog"
)

// Repository provides data access methods
type Repository struct {
	logger *slog.Logger
}

// New creates a new Repository instance
func New(logger *slog.Logger) *Repository {
	return &Repository{
		logger: logger,
	}
}
