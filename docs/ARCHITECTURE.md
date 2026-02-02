# Architecture

This document describes the architectural decisions, patterns, and design principles used in this service.

## Design Principles

### 1. Simplicity Over Cleverness

The codebase prioritizes clear, straightforward code over clever abstractions. Each component has a single, well-defined responsibility.

### 2. Standard Library First

The service uses only Go's standard library, demonstrating that production-ready services don't require external frameworks. This:
- Reduces dependency management overhead
- Minimizes security surface area
- Ensures long-term maintainability
- Leverages well-tested, stable APIs

### 3. Explicit Over Implicit

All configuration, timeouts, and error handling are explicit. There are no hidden defaults or magic behaviors.

### 4. Context Everywhere

Request context flows through the entire stack, from HTTP handler to service layer, enabling:
- Request cancellation propagation
- Timeout enforcement
- Distributed tracing
- Structured logging with correlation

## Project Structure

```
.
├── main.go           # Entry point, lifecycle management
├── config.go         # Configuration loading
├── logger.go         # Structured logging setup
├── server.go         # HTTP server and routing
├── handlers.go       # HTTP request handlers
├── service.go        # Business logic layer
├── middleware.go     # HTTP middleware
└── docs/
    ├── README.md
    └── ARCHITECTURE.md
```

### Why Flat Structure?

The service uses a flat package structure (no `/cmd`, `/internal`, `/pkg`) because:
- The service is simple enough that directories add cognitive overhead
- All code is in a single `main` package
- Easy to refactor later if the service grows
- Follows Go's principle of starting simple

For larger projects, consider:
```
/cmd/server/          # Application entry points
/internal/            # Private application code
/pkg/                 # Public library code
```

## Component Architecture

### Configuration Layer

**File**: `config.go`

Environment-based configuration following 12-factor app principles:
- All config from environment variables
- Sensible defaults for local development
- Type-safe duration parsing
- Single source of truth for configuration

**Pattern**: Struct-based config loaded at startup, passed to components.

### Logging Layer

**File**: `logger.go`

Structured logging using `log/slog`:
- JSON format in production (machine-parseable)
- Text format in development (human-readable)
- Configurable log levels
- Context-aware logging includes trace IDs

**Pattern**: Single logger instance created at startup, passed to all components.

### HTTP Layer

**File**: `server.go`

HTTP server configuration and routing:
- Explicit timeout configuration
- Middleware chain composition
- Route registration
- Server lifecycle management

**Pattern**: Factory function returns configured `*http.Server`.

### Handler Layer

**File**: `handlers.go`

HTTP request handlers:
- Accept dependencies via closure
- Extract context from request
- Call service layer with context
- Handle errors explicitly
- Return JSON responses

**Pattern**: Higher-order functions return `http.HandlerFunc`.

```go
func ExampleHandler(logger *slog.Logger, svc *Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        // Handler logic
    }
}
```

### Service Layer

**File**: `service.go`

Business logic layer:
- Accept `context.Context` as first parameter
- Check context cancellation before expensive operations
- Return explicit errors
- Use context-aware logging

**Pattern**: Service struct holds dependencies, methods accept context.

```go
func (s *Service) GetExample(ctx context.Context, name string) (map[string]interface{}, error) {
    // Check context
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Business logic
}
```

### Middleware Layer

**File**: `middleware.go`

HTTP middleware for cross-cutting concerns:

1. **TraceIDMiddleware**: Generates unique trace ID, adds to context and response
2. **RecoveryMiddleware**: Catches panics, logs, returns 500
3. **LoggingMiddleware**: Logs requests with method, path, status, duration, trace ID

**Pattern**: Middleware chain using higher-order functions.

```go
var handler http.Handler = mux
handler = LoggingMiddleware(logger)(handler)
handler = RecoveryMiddleware(logger)(handler)
handler = TraceIDMiddleware(handler)
```

**Order matters**: Applied in reverse (TraceID → Recovery → Logging → Handler).

## Key Patterns

### Context Propagation

Request context flows through the entire stack:

```
HTTP Request
  → Middleware (adds trace ID to context)
    → Handler (extracts context)
      → Service (receives context, checks cancellation)
        → Returns result or error
      ← Handler (logs with trace ID from context)
    ← Response (includes X-Trace-ID header)
```

### Graceful Shutdown

Shutdown sequence ensures no dropped requests:

```
1. Receive SIGINT/SIGTERM signal
2. Stop accepting new connections
3. Wait for in-flight requests (up to ShutdownTimeout)
4. Close server
5. Exit cleanly
```

**Implementation**: `signal.NotifyContext()` + `server.Shutdown()`.

### Error Handling

Explicit error handling at every layer:
- Service layer returns errors
- Handlers log and convert to HTTP responses
- Middleware catches panics
- Never silently swallow errors

### Dependency Injection

Dependencies passed via function parameters:
- No global state
- Easy to test with mocks
- Clear dependency graph
- Explicit initialization order

## Timeouts

Multiple timeout layers for defense in depth:

| Timeout | Purpose | Default |
|---------|---------|---------|
| ReadTimeout | Max time to read request | 5s |
| WriteTimeout | Max time to write response | 10s |
| IdleTimeout | Max keep-alive idle time | 120s |
| ShutdownTimeout | Max graceful shutdown time | 15s |

**All configurable via environment variables**.

## Extensibility Points

### Adding Database

1. Add driver to `go.mod`
2. Create connection pool in `main.go`
3. Pass `*sql.DB` to `Service` constructor
4. Update `ReadyCheck()` to ping database
5. Use context for query timeouts

```go
type Service struct {
    logger *slog.Logger
    db     *sql.DB
}

func (s *Service) ReadyCheck(ctx context.Context) error {
    return s.db.PingContext(ctx)
}
```

### Adding Metrics

1. Add Prometheus client library
2. Create metrics registry in `main.go`
3. Add metrics middleware
4. Expose `/metrics` endpoint
5. Track request duration, status codes, errors

### Adding Authentication

1. Create JWT/OAuth verification logic
2. Add authentication middleware
3. Extract user from token, add to context
4. Update handlers to check authenticated user

### Adding Background Jobs

1. Create worker pool in `main.go`
2. Use same signal context for coordinated shutdown
3. Wait for workers to finish in shutdown sequence

## Testing Strategy

### Unit Tests

Test individual components in isolation:
- `config_test.go`: Config loading with various env vars
- `handlers_test.go`: Handler logic with `httptest`
- `service_test.go`: Business logic, context cancellation
- `middleware_test.go`: Middleware behavior

### Integration Tests

Test complete request flow:
- Start real HTTP server on random port
- Make HTTP requests
- Verify responses, headers, logs
- Test shutdown sequence

### Test Patterns

```go
// Handler testing
req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)
assert.Equal(t, http.StatusOK, rec.Code)

// Context cancellation
ctx, cancel := context.WithCancel(context.Background())
cancel()
_, err := svc.DoSomething(ctx)
assert.Equal(t, context.Canceled, err)
```

## Performance Considerations

### Connection Pooling

The HTTP server automatically manages connection pooling via `IdleTimeout`. Adjust based on load patterns.

### Request Buffering

`ReadTimeout` and `WriteTimeout` prevent slow clients from holding connections indefinitely.

### Graceful Degradation

Context cancellation allows long-running operations to stop early when clients disconnect.

## Security Considerations

### Input Validation

Validate all user input in handlers before passing to service layer.

### Error Messages

Never expose internal errors to clients. Log detailed errors, return generic messages.

### Timeouts

All timeouts prevent resource exhaustion attacks.

### Trace IDs

Use cryptographically random trace IDs via `crypto/rand`, not predictable sequences.

## Operational Excellence

### Observability

- Structured JSON logs in production
- Trace IDs for request correlation
- Health and readiness endpoints
- HTTP access logs with duration

### Configuration

- All config via environment variables
- No hardcoded values
- Sensible defaults for local dev
- Production values in deployment configs

### Deployment

- Stateless service (horizontally scalable)
- Health check endpoints
- Graceful shutdown for zero-downtime deploys
- Small binary size (stdlib only)

## Trade-offs

### Standard Library Only

**Pros**:
- No dependency management
- Smaller attack surface
- Long-term stability

**Cons**:
- More verbose routing (vs. frameworks)
- Manual middleware composition
- No built-in validation

**Decision**: Simplicity and stability outweigh convenience for this service size.

### Flat Package Structure

**Pros**:
- Simpler navigation
- Less cognitive overhead
- Easy to refactor later

**Cons**:
- All code in `main` package
- Can't import between packages

**Decision**: Appropriate for service of this size, refactor if it grows.

### Environment-Based Config

**Pros**:
- 12-factor compliant
- Easy to deploy
- No config files to manage

**Cons**:
- Less discoverable than config files
- Type conversion required

**Decision**: Industry standard for cloud-native services.

## Future Considerations

As the service grows, consider:

1. **Package structure**: Move to `/cmd`, `/internal`, `/pkg`
2. **Wire dependency injection**: Generate DI code
3. **OpenAPI spec**: Document API with OpenAPI 3.0
4. **Prometheus metrics**: Track request rates, errors, latency
5. **Distributed tracing**: Add OpenTelemetry support
6. **Database**: Add connection pooling and migrations
7. **Caching**: Add Redis for performance
8. **Rate limiting**: Protect against abuse

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [12-Factor App](https://12factor.net/)
