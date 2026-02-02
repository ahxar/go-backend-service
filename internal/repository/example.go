package repository

import (
	"context"
)

// ExampleRepository defines methods for example data access
type ExampleRepository interface {
	GetData(ctx context.Context, id string) (map[string]interface{}, error)
}

// GetData retrieves example data
// In a real application, this would query a database
func (r *Repository) GetData(ctx context.Context, id string) (map[string]interface{}, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Simulate data retrieval
	// In production, this would be:
	// row := r.db.QueryRowContext(ctx, "SELECT * FROM examples WHERE id = $1", id)

	data := map[string]interface{}{
		"id":     id,
		"status": "active",
	}

	return data, nil
}
