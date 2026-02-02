package repository

import (
	"context"
)

// HealthRepository defines methods for health checks
type HealthRepository interface {
	CheckHealth(ctx context.Context) error
	CheckReady(ctx context.Context) error
}

// CheckHealth performs health check on data layer
// In a real application, this would check database connectivity
func (r *Repository) CheckHealth(ctx context.Context) error {
	// Example: return r.db.PingContext(ctx)
	return nil
}

// CheckReady performs readiness check on data layer
// In a real application, this would verify database migrations, etc.
func (r *Repository) CheckReady(ctx context.Context) error {
	// Example: return r.db.PingContext(ctx)
	return nil
}
