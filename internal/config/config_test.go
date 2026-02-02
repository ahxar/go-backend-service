package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	clearEnv()

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}

	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("expected read timeout 5s, got %v", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 10*time.Second {
		t.Errorf("expected write timeout 10s, got %v", cfg.WriteTimeout)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected log level info, got %s", cfg.LogLevel)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected environment development, got %s", cfg.Environment)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	clearEnv()

	if err := os.Setenv("PORT", "9000"); err != nil {
		t.Fatalf("failed to set PORT: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("ENVIRONMENT", "production"); err != nil {
		t.Fatalf("failed to set ENVIRONMENT: %v", err)
	}

	defer clearEnv()

	cfg := Load()

	if cfg.Port != "9000" {
		t.Errorf("expected port 9000, got %s", cfg.Port)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected environment production, got %s", cfg.Environment)
	}
}

func clearEnv() {
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("READ_TIMEOUT")
	_ = os.Unsetenv("WRITE_TIMEOUT")
	_ = os.Unsetenv("IDLE_TIMEOUT")
	_ = os.Unsetenv("SHUTDOWN_TIMEOUT")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("ENVIRONMENT")
}
