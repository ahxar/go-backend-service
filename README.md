# Go HTTP Service

> A production-ready HTTP service template built with Go's standard library. Clean architecture, zero dependencies, and ready to scale.

## 🚀 Quick Start

```bash
# Run the service
go run ./cmd/server

# The server starts on http://localhost:8080
# Try it: curl http://localhost:8080/health
```

That's it! The service is running with:
- ✅ Health check endpoint
- ✅ Request tracing
- ✅ Structured logging
- ✅ Graceful shutdown

## 📦 What You Get

This template provides a solid foundation for building HTTP services in Go:

- **Clean Architecture** - 3-layer separation (Handler → Service → Repository)
- **No Dependencies** - Uses only Go 1.21+ standard library
- **Production Ready** - Includes logging, tracing, health checks, and graceful shutdown
- **Well Tested** - Comprehensive test coverage with examples
- **CI/CD** - GitHub Actions workflow for testing and linting
- **Scalable Structure** - Organized to grow from small to large projects

## 📖 Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Configuration](#configuration)
- [Development](#development)
- [Architecture](#architecture)
- [Extending](#extending)

## 💻 Installation

**Prerequisites:**
- Go 1.21 or higher
- Make (optional, but recommended)

**Get started:**

```bash
# Clone or download this repository
cd go-backend-service

# Run the service
make run
# or
go run ./cmd/server
```

## 🎯 Usage

### Running the Service

```bash
# Development mode
make run

# With custom port
PORT=9000 make run

# With debug logging
LOG_LEVEL=debug make run

# Production mode
ENVIRONMENT=production LOG_LEVEL=info make run
```

### Building

```bash
# Build for your platform
make build
# Creates: ./bin/server

# Build for all platforms (Linux, macOS, Windows)
make build-all

# Run the binary
./bin/server
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run linter
make lint

# Run everything (lint + vet + test)
make check
```

## 🌐 API Endpoints

### Health Check
Check if the service is alive.

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{"status":"healthy"}
```

### Readiness Check
Check if the service is ready to handle traffic.

```bash
curl http://localhost:8080/ready
```

**Response:**
```json
{"status":"ready"}
```

### Example Endpoint
A sample endpoint demonstrating the full request lifecycle.

```bash
curl http://localhost:8080/api/example?name=World
```

**Response:**
```json
{
  "message": "Hello, World!",
  "timestamp": "2026-02-02T12:34:56Z",
  "processed": true
}
```

**Note:** Every response includes an `X-Trace-ID` header for request correlation across logs.

## ⚙️ Configuration

Configure the service using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `ENVIRONMENT` | `development` | Environment: development, production |
| `READ_TIMEOUT` | `5s` | Maximum time to read requests |
| `WRITE_TIMEOUT` | `10s` | Maximum time to write responses |
| `IDLE_TIMEOUT` | `120s` | Keep-alive timeout |
| `SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown timeout |

**Example:**
```bash
# Create a .env file (optional)
cp .env.example .env

# Edit values, then run
export $(cat .env | xargs) && go run ./cmd/server
```

## 🛠️ Development

### Project Structure

```
cmd/server/           # Application entry point
internal/
  ├── handler/        # HTTP handlers (request/response)
  ├── service/        # Business logic
  ├── repository/     # Data access (databases, APIs)
  ├── middleware/     # HTTP middleware
  ├── model/          # Data models
  ├── config/         # Configuration
  └── server/         # Server setup
pkg/logger/           # Reusable logger
```

### Development Tools

Install recommended tools for development:

```bash
make install-tools
```

This installs:
- **golangci-lint** - Fast Go linters runner
- **goimports** - Automatic import formatting
- **air** - Live reload for development

### Available Commands

```bash
make help              # Show all available commands
make build             # Build the application
make test              # Run tests with coverage
make lint              # Run linters
make fmt               # Format code
make run               # Run the application
make clean             # Clean build artifacts
make check             # Run all checks
make ci                # Run CI pipeline locally
```

## 🏗️ Architecture

This service follows a **3-layer architecture** for clean separation of concerns:

### 1. Handler Layer (`internal/handler/`)
Handles HTTP requests and responses.

**Responsibilities:**
- Parse and validate requests
- Format responses
- Handle HTTP-specific concerns
- Call the service layer

### 2. Service Layer (`internal/service/`)
Contains business logic.

**Responsibilities:**
- Implement business rules
- Orchestrate operations
- Call the repository layer
- Return results or errors

### 3. Repository Layer (`internal/repository/`)
Manages data access.

**Responsibilities:**
- Database queries
- External API calls
- Cache operations
- Data persistence

**Request Flow:**
```
HTTP Request → Handler → Service → Repository → Database
HTTP Response ← Handler ← Service ← Repository ← Database
```

[See detailed architecture documentation →](docs/ARCHITECTURE.md)

## 🔧 Extending

### Adding a New Endpoint

Follow these steps to add a new API endpoint:

**1. Define your model** in `internal/model/`:
```go
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

**2. Add repository method** in `internal/repository/`:
```go
func (r *Repository) CreateUser(ctx context.Context, name string) (*model.User, error) {
    // Database logic here
    return &model.User{ID: "1", Name: name}, nil
}
```

**3. Add service method** in `internal/service/`:
```go
func (s *Service) CreateUser(ctx context.Context, name string) (*model.User, error) {
    // Business logic here
    return s.repo.CreateUser(ctx, name)
}
```

**4. Add handler** in `internal/handler/`:
```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Parse request, call service, return response
}
```

**5. Register route** in `internal/server/server.go`:
```go
mux.HandleFunc("POST /api/users", h.CreateUser)
```

### Adding a Database

**1. Update repository** to accept database connection:
```go
type Repository struct {
    logger *slog.Logger
    db     *sql.DB
}
```

**2. Initialize in main** (`cmd/server/main.go`):
```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatal(err)
}
defer db.Close()

repo := repository.New(log, db)
```

### Growing Your Service

As your service grows, organize code by domain:

**Small service** (< 10 endpoints):
- Keep each layer in one file per domain

**Medium service** (10-50 endpoints):
- Split handlers into separate files
- Keep service and repository combined

**Large service** (> 50 endpoints):
- Create subdirectories per domain
- Split all layers by feature

Example structure for large services:
```
internal/
  ├── handler/
  │   ├── user/
  │   │   ├── create.go
  │   │   ├── get.go
  │   │   └── update.go
  │   └── product/
  └── service/
      ├── user/
      └── product/
```

## 🧪 Testing

The service includes comprehensive tests:

```bash
# Run all tests
make test

# Run tests with race detection
go test -race ./...

# Generate coverage report
make test-coverage
open coverage.html
```

**Testing strategy:**
- **Unit tests** - Test individual components
- **Integration tests** - Test layer interactions
- **Table-driven tests** - Test multiple scenarios
- **Coverage target** - Aim for >70%

## 🔒 Security

The service includes security best practices:

- ✅ No hardcoded credentials
- ✅ Environment-based configuration
- ✅ Request timeouts prevent resource exhaustion
- ✅ Panic recovery middleware
- ✅ Input validation at handler layer
- ✅ Structured logging (no sensitive data)

## 🚦 CI/CD

The project includes a GitHub Actions workflow for:

- **Testing** - Run tests with race detection
- **Linting** - Check code quality with golangci-lint
- **Building** - Verify the service builds
- **Security** - Scan for vulnerabilities with Gosec

See `.github/workflows/ci.yml` for details.

## 📚 Additional Documentation

- [Architecture Details](docs/ARCHITECTURE.md) - In-depth architecture decisions
- [Request Lifecycle](docs/request-lifecycle.puml) - Visual request flow diagram
- [Graceful Shutdown](docs/graceful-shutdown.puml) - Shutdown sequence diagram

## 🤝 Contributing

Contributions are welcome! Here's how:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Make your changes
4. Run tests (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing`)
7. Open a Pull Request

## 📝 License

MIT

---

**Built with ❤️ using Go's standard library**

### Project Status

✅ Production-ready
✅ Well-tested
✅ Zero dependencies
✅ Actively maintained
