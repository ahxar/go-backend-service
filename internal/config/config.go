package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	LogLevel        string
	Environment     string
	// OpenTelemetry configuration
	OtelEnabled        bool
	OtelEndpoint       string
	OtelServiceName    string
	OtelServiceVersion string
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		ReadTimeout:     getEnv("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getEnv("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     getEnv("IDLE_TIMEOUT", 120*time.Second),
		ShutdownTimeout: getEnv("SHUTDOWN_TIMEOUT", 15*time.Second),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		// OpenTelemetry configuration
		OtelEnabled:        getEnv("OTEL_ENABLED", true),
		OtelEndpoint:       getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		OtelServiceName:    getEnv("OTEL_SERVICE_NAME", "go-backend-service"),
		OtelServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
	}
}

// getEnv retrieves an environment variable, parses it based on type, or returns a default value
func getEnv[T any](key string, defaultValue T) T {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result any
	var err error

	// Parse based on the type of defaultValue
	switch any(defaultValue).(type) {
	case string:
		result = value
	case bool:
		result, err = strconv.ParseBool(value)
	case time.Duration:
		result, err = time.ParseDuration(value)
	default:
		return defaultValue
	}

	if err != nil {
		return defaultValue
	}

	return result.(T)
}
