# Production-Ready Go HTTP Service

A production-ready Go HTTP service built with clean 3-layer architecture using only the Go standard library. Designed to scale with proper separation of concerns, comprehensive testing, and CI/CD automation.

## Features

- **3-Layer Architecture** - Clean separation: Handler → Service → Repository
- **Zero External Dependencies** - Built entirely with Go 1.21+ standard library
- **Scalable Structure** - Organized for growth with `/cmd`, `/internal`, `/pkg`
- **Graceful Shutdown** - Handles SIGINT/SIGTERM signals and drains in-flight requests
- **Structured Logging** - JSON logging in production using `log/slog`
- **Request Tracing** - Unique trace ID per request for distributed tracing
- **Context Propagation** - Request context flows through entire stack
- **Explicit Timeouts** - Configured at server and request levels
- **Health & Readiness** - Health check endpoints
- **12-Factor Config** - Environment-based configuration
- **CI/CD Ready** - GitHub Actions for testing, linting, and building

## Project Structure

```
/Users/safar/Projects/go-backend-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration loading
│   │   └── config_test.go
│   ├── handler/                 # HTTP layer
│   │   ├── handler.go           # Handler infrastructure
│   │   ├── health.go            # Health check handlers
│   │   ├── example.go           # Example API handlers
│   │   └── handler_test.go
│   ├── service/                 # Business logic layer
│   │   ├── service.go           # Service infrastructure
│   │   ├── health.go            # Health check logic
│   │   ├── example.go           # Example business logic
│   │   └── service_test.go
│   ├── repository/              # Data access layer
│   │   ├── repository.go        # Repository infrastructure
│   │   ├── health.go            # Health check data access
│   │   └── example.go           # Example data access
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware
│   ├── model/
│   │   └── example.go           # Domain models
│   └── server/
│       └── server.go            # HTTP server setup
├── pkg/
│   └── logger/
│       └── logger.go            # Reusable logger package
├── .github/
│   └── workflows/
│       └── ci.yml               # CI pipeline
├── docs/                        # Documentation
├── Makefile                     # Build automation
├── .golangci.yml               # Linter configuration
├── .gitignore
├── .env.example
├── go.mod
└── README.md
```

## Architecture Layers

### Handler Layer (`internal/handler/`)

- HTTP request/response handling
- Request validation
- Response formatting
- Calls service layer

### Service Layer (`internal/service/`)

- Business logic
- Transaction management
- Orchestrates repository calls
- Context-aware operations

### Repository Layer (`internal/repository/`)

- Data access
- Database queries
- External service calls
- Cache operations

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Make (optional, for Makefile commands)

### Running Locally

```bash
# Using go run
go run ./cmd/server

# Using make
make run

# With custom configuration
PORT=9000 LOG_LEVEL=debug go run ./cmd/server
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Using go directly
go build -o server ./cmd/server
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Run all checks
make check
```

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/health
# {"status":"healthy"}
```

### Readiness Check

```bash
curl http://localhost:8080/ready
# {"status":"ready"}
```

### Example Endpoint

```bash
curl http://localhost:8080/api/example?name=World
# {"message":"Hello, World!","timestamp":"2026-02-02T12:34:56Z","processed":true}
```

Every response includes an `X-Trace-ID` header for request correlation.

## Configuration

All configuration via environment variables:

| Variable           | Default       | Description                              |
| ------------------ | ------------- | ---------------------------------------- |
| `PORT`             | `8080`        | HTTP server port                         |
| `READ_TIMEOUT`     | `5s`          | Request read timeout                     |
| `WRITE_TIMEOUT`    | `10s`         | Response write timeout                   |
| `IDLE_TIMEOUT`     | `120s`        | Keep-alive idle timeout                  |
| `SHUTDOWN_TIMEOUT` | `15s`         | Graceful shutdown timeout                |
| `LOG_LEVEL`        | `info`        | Logging level (debug, info, warn, error) |
| `ENVIRONMENT`      | `development` | Environment (development, production)    |

See `.env.example` for a complete configuration template.

## CI/CD Pipeline

### Continuous Integration (`.github/workflows/ci.yml`)

Runs on push and pull requests:

- **Test**: Run tests with race detection and coverage
- **Lint**: golangci-lint with comprehensive checks
- **Build**: Verify binary builds successfully
- **Security**: Gosec security scanner

## Adding New Features

### Add a New Endpoint

1. **Define Model** (`internal/model/`)

```go
type UserRequest struct {
    Name string `json:"name"`
}

type UserResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

2. **Add Repository Method** (`internal/repository/user.go`)

```go
func (r *Repository) CreateUser(ctx context.Context, name string) (*model.UserResponse, error) {
    // Database logic here
    return &model.UserResponse{ID: "1", Name: name}, nil
}
```

3. **Add Service Method** (`internal/service/user.go`)

```go
func (s *Service) CreateUser(ctx context.Context, req *model.UserRequest) (*model.UserResponse, error) {
    // Business logic here
    return s.repo.CreateUser(ctx, req.Name)
}
```

4. **Add Handler** (`internal/handler/user.go`)

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var req model.UserRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeJSON(w, http.StatusBadRequest, &model.ErrorResponse{Error: "invalid request"})
        return
    }

    result, err := h.service.CreateUser(ctx, &req)
    if err != nil {
        h.writeJSON(w, http.StatusInternalServerError, &model.ErrorResponse{Error: err.Error()})
        return
    }

    h.writeJSON(w, http.StatusCreated, result)
}
```

5. **Register Route** (`internal/server/server.go`)

```go
mux.HandleFunc("POST /api/users", h.CreateUser)
```

### Add Database Connection

1. **Update Repository** (`internal/repository/repository.go`)

```go
import "database/sql"

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

2. **Initialize in main** (`cmd/server/main.go`)

```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatal(err)
}
defer db.Close()

repo := repository.New(log, db)
```

## Makefile Commands

```bash
make help              # Show all commands
make build             # Build binary
make build-all         # Build for all platforms
make test              # Run tests
make test-coverage     # Generate coverage report
make lint              # Run linters
make fmt               # Format code
make clean             # Clean build artifacts
make run               # Run application
make check             # Run all checks (lint + vet + test)
make ci                # Run full CI pipeline locally
```

## Development Tools

Install recommended tools:

```bash
make install-tools
```

This installs:

- golangci-lint (linting)
- goimports (import formatting)
- air (hot reload for development)

## Testing Strategy

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test layer interactions
- **Coverage**: Target >70% code coverage
- **Race Detection**: All tests run with `-race` flag
- **Mocking**: Use interfaces for dependency injection

## Best Practices

### Organizing Large Services

As your service grows, organize each layer by domain:

```
internal/
├── handler/
│   ├── user/           # User-related handlers
│   │   ├── create.go
│   │   ├── get.go
│   │   ├── update.go
│   │   └── delete.go
│   └── product/        # Product-related handlers
│       ├── create.go
│       └── list.go
├── service/
│   ├── user/           # User business logic
│   └── product/        # Product business logic
└── repository/
    ├── user/           # User data access
    └── product/        # Product data access
```

### Context Usage

Always pass context as the first parameter:

```go
func (s *Service) ProcessData(ctx context.Context, data string) error {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Your logic here
}
```

### Error Handling

Return wrapped errors for better debugging:

```go
if err := s.repo.Save(ctx, data); err != nil {
    return fmt.Errorf("failed to save data: %w", err)
}
```

## Performance Considerations

- Connection pooling at repository layer
- Request timeouts prevent resource exhaustion
- Context cancellation stops work early
- Structured logging with appropriate levels
- Efficient JSON encoding/decoding

## Security

- No hardcoded credentials
- Environment-based configuration
- Timeouts on all operations
- Panic recovery middleware
- Input validation at handler layer

## Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - Detailed architecture decisions
- [API Documentation](docs/README.md) - Complete API reference
- [Request Lifecycle](docs/request-lifecycle.puml) - Request flow diagram
- [Graceful Shutdown](docs/graceful-shutdown.puml) - Shutdown sequence

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make check`
5. Submit a pull request

## License

MIT

## Project Status

✅ Production-ready 3-layer architecture
✅ Comprehensive test coverage
✅ CI/CD automation with GitHub Actions
✅ Health check endpoints
✅ Zero external dependencies
✅ Extensive documentation
