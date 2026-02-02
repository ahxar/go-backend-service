package otel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config holds OpenTelemetry configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Enabled        bool
}

// Setup initializes OpenTelemetry with tracing and metrics
func Setup(ctx context.Context, cfg Config, logger *slog.Logger) (func(context.Context) error, error) {
	if !cfg.Enabled {
		logger.Info("OpenTelemetry is disabled")
		return func(context.Context) error { return nil }, nil
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Setup trace provider
	traceShutdown, err := setupTraceProvider(ctx, res, cfg.Endpoint, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup trace provider: %w", err)
	}

	// Setup metric provider
	metricShutdown, err := setupMeterProvider(ctx, res, cfg.Endpoint, logger)
	if err != nil {
		traceShutdown(ctx)
		return nil, fmt.Errorf("failed to setup meter provider: %w", err)
	}

	logger.Info("OpenTelemetry initialized",
		slog.String("service", cfg.ServiceName),
		slog.String("endpoint", cfg.Endpoint),
	)

	// Return combined shutdown function
	return func(ctx context.Context) error {
		var err error
		if shutdownErr := traceShutdown(ctx); shutdownErr != nil {
			err = shutdownErr
		}
		if shutdownErr := metricShutdown(ctx); shutdownErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; %w", err, shutdownErr)
			} else {
				err = shutdownErr
			}
		}
		return err
	}, nil
}

func setupTraceProvider(ctx context.Context, res *resource.Resource, endpoint string, logger *slog.Logger) (func(context.Context) error, error) {
	// Strip scheme from endpoint if present (WithEndpoint expects host:port only)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	// Create OTLP trace exporter
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second),
		),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("trace provider initialized")

	return traceProvider.Shutdown, nil
}

func setupMeterProvider(ctx context.Context, res *resource.Resource, endpoint string, logger *slog.Logger) (func(context.Context) error, error) {
	// Strip scheme from endpoint if present (WithEndpoint expects host:port only)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	// Create OTLP metric exporter
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create meter provider
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(10*time.Second),
		)),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	logger.Info("meter provider initialized")

	return meterProvider.Shutdown, nil
}
