package model

import "time"

// ExampleRequest represents an example API request
type ExampleRequest struct {
	Name string `json:"name"`
}

// ExampleResponse represents an example API response
type ExampleResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Processed bool      `json:"processed"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadyResponse represents a readiness check response
type ReadyResponse struct {
	Status string `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
