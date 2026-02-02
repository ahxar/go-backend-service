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
├── cmd/
│   └── server/
│       └── main.go              # Entry point, lifecycle management
├── internal/
│   ├── handler/                 # HTTP handlers (request/response)
│   │   ├── handler.go           # Handler struct and JSON utilities
│   │   ├── health.go            # Health check endpoints
│   │   └── example.go           # Example endpoint
│   ├── service/                 # Business logic layer
│   │   ├── service.go           # Service struct and constructor
│   │   ├── health.go            # Health check logic
│   │   └── example.go           # Example business logic
│   ├── repository/              # Data access layer
│   │   ├── repository.go        # Repository struct and constructor
│   │   ├── health.go            # Health data operations
│   │   └── example.go           # Example data operations
│   ├── middleware/              # HTTP middleware
│   │   └── middleware.go        # TraceID, Recovery, Logging
│   ├── server/                  # HTTP server setup
│   │   └── server.go            # Server configuration and routing
│   ├── config/                  # Configuration
│   │   └── config.go            # Environment-based config loading
│   └── model/                   # Domain models
│       └── example.go           # Data structures
├── pkg/
│   └── logger/                  # Reusable logger package
│       └── logger.go            # Structured logging setup
└── docs/
    ├── ARCHITECTURE.md
    ├── graceful-shutdown.puml
    └── request-lifecycle.puml
```

### Why This Structure?

The service uses a clean, layered architecture with clear separation of concerns:

**`cmd/server/`** - Application entry point
- Orchestrates initialization and lifecycle
- Single responsibility: start and stop the application

**`internal/`** - Private application code
- Cannot be imported by external packages
- Ensures encapsulation of implementation details
- Organized by layer and concern

**`pkg/`** - Reusable library code
- Can be imported by external packages
- Contains utilities that could be shared across services

This structure provides:
- Clear separation between layers (handler → service → repository)
- Easy navigation (each package has a single responsibility)
- Testability (each layer can be tested in isolation)
- Scalability (easy to add new features within the established structure)

## Component Architecture

### Configuration Layer

**Package**: `internal/config`
**File**: `config.go`

Environment-based configuration following 12-factor app principles:
- All config from environment variables
- Sensible defaults for local development
- Type-safe duration parsing
- Single source of truth for configuration

**Pattern**: Struct-based config loaded at startup, passed to components.

```go
type Config struct {
    Port            string
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    IdleTimeout     time.Duration
    ShutdownTimeout time.Duration
    LogLevel        string
    Environment     string
}
```

### Logging Layer

**Package**: `pkg/logger`
**File**: `logger.go`

Structured logging using `log/slog`:
- JSON format in production (machine-parseable)
- Text format in development (human-readable)
- Configurable log levels
- Context-aware logging includes trace IDs

**Pattern**: Single logger instance created at startup, passed to all components.

```go
// Usage
log := logger.New(cfg.Environment, cfg.LogLevel)
log.InfoContext(ctx, "message", slog.String("key", "value"))
```

### HTTP Layer

**Package**: `internal/server`
**File**: `server.go`

HTTP server configuration and routing:
- Explicit timeout configuration
- Middleware chain composition
- Route registration
- Server lifecycle management

**Pattern**: Factory function returns configured `*http.Server`.

```go
func New(cfg *config.Config, logger *slog.Logger, h *handler.Handler) *http.Server {
    mux := http.NewServeMux()

    // Register routes
    mux.HandleFunc("GET /health", h.Health)
    mux.HandleFunc("GET /ready", h.Ready)
    mux.HandleFunc("GET /api/example", h.Example)

    // Apply middleware chain
    var httpHandler http.Handler = mux
    httpHandler = middleware.Logging(logger)(httpHandler)
    httpHandler = middleware.Recovery(logger)(httpHandler)
    httpHandler = middleware.TraceID(httpHandler)

    return &http.Server{
        Addr:         fmt.Sprintf(":%s", cfg.Port),
        Handler:      httpHandler,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
        IdleTimeout:  cfg.IdleTimeout,
    }
}
```

### Handler Layer

**Package**: `internal/handler`
**Files**: `handler.go`, `health.go`, `example.go`

HTTP request handlers:
- Struct-based handler with dependencies
- Extract context from request
- Call service layer with context
- Handle errors explicitly
- Return JSON responses

**Pattern**: Handler struct holds dependencies, methods implement `http.HandlerFunc`.

```go
type Handler struct {
    logger  *slog.Logger
    service *service.Service
}

func (h *Handler) Example(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    name := r.URL.Query().Get("name")

    result, err := h.service.GetExample(ctx, name)
    if err != nil {
        h.writeJSON(w, http.StatusInternalServerError,
            map[string]string{"error": "internal server error"})
        return
    }

    h.writeJSON(w, http.StatusOK, result)
}
```

### Service Layer

**Package**: `internal/service`
**Files**: `service.go`, `health.go`, `example.go`

Business logic layer:
- Accept `context.Context` as first parameter
- Check context cancellation before expensive operations
- Call repository layer for data access
- Return explicit errors
- Use context-aware logging

**Pattern**: Service struct holds dependencies (logger, repository), methods accept context.

```go
type Service struct {
    logger *slog.Logger
    repo   *repository.Repository
}

func (s *Service) GetExample(ctx context.Context, name string) (*model.ExampleResponse, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    s.logger.InfoContext(ctx, "processing example request",
        slog.String("name", name))

    // Call repository for data
    data, err := s.repo.GetExampleData(ctx, name)
    if err != nil {
        return nil, err
    }

    // Business logic
    return data, nil
}
```

### Repository Layer

**Package**: `internal/repository`
**Files**: `repository.go`, `health.go`, `example.go`

Data access layer:
- Handles all data persistence and retrieval
- In a real app, would contain database queries, API calls, caching
- Accept `context.Context` for cancellation and timeouts
- Returns domain models from `internal/model`

**Pattern**: Repository struct holds dependencies (logger, database connections), methods accept context.

```go
type Repository struct {
    logger *slog.Logger
    // In production: db *sql.DB
}

func (r *Repository) GetExampleData(ctx context.Context, name string) (*model.ExampleResponse, error) {
    // Database query with context
    // Currently returns mock data for demonstration
    return &model.ExampleResponse{
        Message:   fmt.Sprintf("Hello, %s!", name),
        Timestamp: time.Now(),
        Processed: true,
    }, nil
}
```

### Middleware Layer

**Package**: `internal/middleware`
**File**: `middleware.go`

HTTP middleware for cross-cutting concerns:

1. **TraceID**: Generates unique trace ID using `crypto/rand`, adds to context and response header
2. **Recovery**: Catches panics, logs with context, returns 500 with JSON error
3. **Logging**: Logs requests with method, path, status, duration, trace ID

**Pattern**: Middleware chain using higher-order functions.

```go
var httpHandler http.Handler = mux
httpHandler = middleware.Logging(logger)(httpHandler)
httpHandler = middleware.Recovery(logger)(httpHandler)
httpHandler = middleware.TraceID(httpHandler)
```

**Order matters**: Applied in reverse (TraceID → Recovery → Logging → Handler).

**Key features**:
- TraceID uses cryptographically random IDs via `crypto/rand`
- Recovery properly handles error response writing with error checking
- Logging captures status code using a wrapped `responseWriter`
- All middleware is context-aware and includes trace IDs in logs

## Key Patterns

### Context Propagation

Request context flows through the entire stack:

```
HTTP Request
  → TraceID Middleware (generates ID, adds to context & response header)
    → Recovery Middleware (defers panic recovery with context)
      → Logging Middleware (captures start time)
        → Handler (extracts context, query params)
          → Service (receives context, checks cancellation, business logic)
            → Repository (receives context, data access)
            ← Returns data or error
          ← Service returns result or error
        ← Handler returns JSON response
      ← Logging logs with duration, status, trace ID
    ← Recovery catches any panics
  ← Response (includes X-Trace-ID header)
```

**Key points:**
- Context flows from `http.Request.Context()` through all layers
- Trace ID is extracted using `middleware.GetTraceID(ctx)`
- All logging includes trace ID automatically via `logger.InfoContext(ctx, ...)`
- Context cancellation propagates from client disconnect through all layers

### Graceful Shutdown

Shutdown sequence ensures no dropped requests:

```
1. Server starts in goroutine via srv.ListenAndServe()
2. Main goroutine blocks on <-ctx.Done()
3. User sends SIGINT (Ctrl+C) or SIGTERM
4. signal.NotifyContext closes context
5. Main goroutine unblocks, logs "shutdown signal received"
6. Create shutdown context with timeout (default: 15s)
7. Call srv.Shutdown(shutdownCtx)
8. Server stops accepting new connections
9. Server waits for in-flight requests to complete
10. Server closes, main goroutine exits with code 0
```

**Implementation**:
```go
// Create signal context
ctx, stop := signal.NotifyContext(context.Background(),
    os.Interrupt, syscall.SIGTERM)
defer stop()

// Start server in goroutine
go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Error("server error", slog.String("error", err.Error()))
        os.Exit(1)
    }
}()

// Block until signal
<-ctx.Done()

// Graceful shutdown with timeout
shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
defer cancel()

if err := srv.Shutdown(shutdownCtx); err != nil {
    log.Error("shutdown error", slog.String("error", err.Error()))
    os.Exit(1)
}
```

### Error Handling

Explicit error handling at every layer:
- Service layer returns errors
- Handlers log and convert to HTTP responses
- Middleware catches panics
- Never silently swallow errors

### Dependency Injection

Dependencies passed via constructor functions:
- No global state
- Easy to test with mocks
- Clear dependency graph
- Explicit initialization order

**Initialization flow** in `cmd/server/main.go`:
```go
// 1. Load configuration
cfg := config.Load()

// 2. Initialize logger
log := logger.New(cfg.Environment, cfg.LogLevel)

// 3. Initialize repository (bottom layer)
repo := repository.New(log)

// 4. Initialize service (middle layer)
svc := service.New(log, repo)

// 5. Initialize handler (top layer)
h := handler.New(log, svc)

// 6. Create HTTP server
srv := server.New(cfg, log, h)
```

**Benefits:**
- Clear dependency tree: handler → service → repository
- Each layer only depends on the layer below
- Easy to swap implementations for testing
- Compile-time dependency verification

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

1. Add driver to `go.mod`:
   ```bash
   go get github.com/lib/pq  # PostgreSQL example
   ```

2. Create connection pool in `cmd/server/main.go`:
   ```go
   db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
   if err != nil {
       log.Error("failed to open database", slog.String("error", err.Error()))
       os.Exit(1)
   }
   defer db.Close()

   // Set connection pool settings
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(5)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

3. Pass `*sql.DB` to `Repository` constructor:
   ```go
   repo := repository.New(log, db)
   ```

4. Update `internal/repository/repository.go`:
   ```go
   type Repository struct {
       logger *slog.Logger
       db     *sql.DB
   }

   func New(logger *slog.Logger, db *sql.DB) *Repository {
       return &Repository{
           logger: logger,
           db:     db,
       }
   }
   ```

5. Update `internal/repository/health.go` to ping database:
   ```go
   func (r *Repository) CheckHealth(ctx context.Context) error {
       return r.db.PingContext(ctx)
   }
   ```

6. Use context for all queries:
   ```go
   func (r *Repository) GetUser(ctx context.Context, id string) (*model.User, error) {
       var user model.User
       err := r.db.QueryRowContext(ctx,
           "SELECT id, name, email FROM users WHERE id = $1", id,
       ).Scan(&user.ID, &user.Name, &user.Email)
       return &user, err
   }
   ```

### Adding Metrics

1. Add Prometheus client library
2. Create metrics registry in `main.go`
3. Add metrics middleware
4. Expose `/metrics` endpoint
5. Track request duration, status codes, errors

### Adding Authentication

1. Add JWT library to `go.mod`:
   ```bash
   go get github.com/golang-jwt/jwt/v5
   ```

2. Create authentication middleware in `internal/middleware/middleware.go`:
   ```go
   func Auth(secret string) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               // Extract token from Authorization header
               authHeader := r.Header.Get("Authorization")
               if authHeader == "" {
                   http.Error(w, "missing authorization", http.StatusUnauthorized)
                   return
               }

               // Verify and parse JWT
               token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
                   return []byte(secret), nil
               })
               if err != nil || !token.Valid {
                   http.Error(w, "invalid token", http.StatusUnauthorized)
                   return
               }

               // Extract claims and add to context
               claims := token.Claims.(jwt.MapClaims)
               ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
               next.ServeHTTP(w, r.WithContext(ctx))
           })
       }
   }
   ```

3. Apply middleware to protected routes in `internal/server/server.go`:
   ```go
   // Public routes
   mux.HandleFunc("GET /health", h.Health)
   mux.HandleFunc("POST /auth/login", h.Login)

   // Protected routes
   protectedMux := http.NewServeMux()
   protectedMux.HandleFunc("GET /api/example", h.Example)
   mux.Handle("/api/", middleware.Auth(cfg.JWTSecret)(protectedMux))
   ```

4. Extract user in handlers:
   ```go
   func (h *Handler) Example(w http.ResponseWriter, r *http.Request) {
       userID := r.Context().Value("user_id").(string)
       // Use userID in business logic
   }
   ```

### Adding Background Jobs

1. Create worker pool in `main.go`
2. Use same signal context for coordinated shutdown
3. Wait for workers to finish in shutdown sequence

## Testing Strategy

### Unit Tests

Test individual components in isolation:
- `internal/config/config_test.go`: Config loading with various env vars
- `internal/handler/handler_test.go`: Handler logic with `httptest`
- `internal/service/service_test.go`: Business logic, context cancellation
- `internal/middleware/middleware_test.go`: Middleware behavior
- `internal/repository/repository_test.go`: Data access with mocked database

### Integration Tests

Test complete request flow:
- Start real HTTP server on random port
- Make HTTP requests
- Verify responses, headers, logs
- Test shutdown sequence

### Test Patterns

```go
// Handler testing with dependencies
func TestHandler_Health(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    repo := repository.New(logger)
    svc := service.New(logger, repo)
    h := handler.New(logger, svc)

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rec := httptest.NewRecorder()

    h.Health(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Body.String(), `"status":"healthy"`)
}

// Service testing with context cancellation
func TestService_GetExample_ContextCanceled(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    repo := repository.New(logger)
    svc := service.New(logger, repo)

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    _, err := svc.GetExample(ctx, "test")
    assert.Equal(t, context.Canceled, err)
}

// Middleware testing
func TestMiddleware_TraceID(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := middleware.GetTraceID(r.Context())
        assert.NotEmpty(t, traceID)
        w.WriteHeader(http.StatusOK)
    })

    wrapped := middleware.TraceID(handler)
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()

    wrapped.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
    assert.NotEmpty(t, rec.Header().Get("X-Trace-ID"))
}
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

### Layered Package Structure

**Pros**:
- Clear separation of concerns (handler → service → repository)
- Each package has single responsibility
- Easy to test layers in isolation
- Enforces proper dependencies (layers only depend on lower layers)
- Scales well as service grows

**Cons**:
- More directories to navigate
- Slightly more boilerplate for small services
- Need to understand layer boundaries

**Decision**: The structure provides clear organization and scales well as features are added. The benefits of separation outweigh the minimal overhead for a service of this size.

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

The service already has a solid foundation. As it grows, consider:

1. ✅ **Package structure**: Already using `/cmd`, `/internal`, `/pkg`
2. **Wire dependency injection**: Generate DI code for complex dependency graphs
3. **OpenAPI spec**: Document API with OpenAPI 3.0 / Swagger
4. **Prometheus metrics**: Track request rates, errors, latency
   ```go
   // Add metrics middleware
   httpHandler = middleware.Metrics()(httpHandler)
   mux.Handle("/metrics", promhttp.Handler())
   ```
5. **Distributed tracing**: Add OpenTelemetry support
   ```go
   // Initialize tracer
   tp := trace.NewTracerProvider(...)
   otel.SetTracerProvider(tp)

   // Add tracing middleware
   httpHandler = otelhttp.NewHandler(httpHandler, "server")
   ```
6. **Database**: Add connection pooling and migrations (see "Adding Database" above)
7. **Caching**: Add Redis for performance
   ```go
   // Initialize Redis client
   rdb := redis.NewClient(&redis.Options{...})

   // Pass to repository
   repo := repository.New(log, db, rdb)
   ```
8. **Rate limiting**: Protect against abuse
   ```go
   // Add rate limiting middleware
   httpHandler = middleware.RateLimit(100)(httpHandler)
   ```
9. **API Versioning**: Add version prefixes to routes
   ```go
   mux.HandleFunc("GET /api/v1/example", h.Example)
   mux.HandleFunc("GET /api/v2/example", h.ExampleV2)
   ```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [12-Factor App](https://12factor.net/)
