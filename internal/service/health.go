package service

import (
	"context"
)

// HealthService defines business logic for health checks
type HealthService interface {
	CheckHealth(ctx context.Context) error
	CheckReady(ctx context.Context) error
}

// CheckHealth performs comprehensive health check
func (s *Service) CheckHealth(ctx context.Context) error {
	// Check repository layer health
	if err := s.repo.CheckHealth(ctx); err != nil {
		return err
	}

	// Add additional health checks here
	// - External service connectivity
	// - Cache availability
	// - etc.

	return nil
}

// CheckReady performs comprehensive readiness check
func (s *Service) CheckReady(ctx context.Context) error {
	// Check repository layer readiness
	if err := s.repo.CheckReady(ctx); err != nil {
		return err
	}

	// Add additional readiness checks here
	// - Database migrations complete
	// - Required data seeded
	// - etc.

	return nil
}
