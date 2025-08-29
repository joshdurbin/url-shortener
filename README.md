# URL Shortener

A modern, production-ready URL shortening service built with Go, featuring clean architecture, comprehensive testing, and observability.

## Features

- **Clean Architecture**: Well-structured codebase with clear separation of concerns
- **In-Memory Caching**: Thread-safe in-memory cache with background synchronization
- **Type-Safe Database**: SQLite backend with sqlc-generated type-safe queries
- **RESTful API**: HTTP server with comprehensive endpoints
- **CLI Client**: Command-line interface for easy interaction
- **Observability**: Comprehensive Prometheus metrics and health checks
- **Testing**: Extensive unit and integration test coverage
- **Multiple Shortening Algorithms**: Choose between MD5 hash, Base62 counter, Base62 random, or NanoID
- **Counter-based Generation**: In-memory counter cache with jump-ahead allocation and async database writeback

## Quick Start

### Prerequisites

- Go 1.19 or later
- Make (for development commands)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd url-shortener

# Install dependencies and set up development environment
make dev-setup

# Build the application
make build
```

### Running the Server

```bash
# Using example configuration
make run-example

# Or build and run manually
make build
./url-shortener server
```

The server will start on `http://localhost:8080` by default.

### Using the CLI Client

```bash
# Create a short URL
go run ./cmd/server client create "https://example.com"

# Get URL information
go run ./cmd/server client get <short_code>

# List all URLs
go run ./cmd/server client list

# Delete a URL
go run ./cmd/server client delete <short_code>
```

## API Usage

### Create Short URL
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

### Access Short URL
```bash
curl http://localhost:8080/{short_code}
# Returns 302 redirect to original URL
```

### Get URL Information
```bash
curl http://localhost:8080/api/urls/{short_code}
```

### List All URLs
```bash
curl http://localhost:8080/api/urls
```

### Delete URL
```bash
curl -X DELETE http://localhost:8080/api/urls/{short_code}
```

## Configuration

### YAML Configuration

Create a `config.yaml` file (use `config.example.yaml` as a template):

```yaml
server:
  port: "8080"
  server_url: "http://localhost:8080"
database:
  path: "urls.db"
cache:
  type: "memory"
  sync_interval: "5s"
metrics:
  enabled: true
  port: "9090"
  endpoint: "/metrics"
```

### CLI Configuration

The application is configured via CLI flags:

```bash
# Server options
--port, -p                 Server port (default: "8080")
--server-url              Server URL (default: "http://localhost:8080")
--db-path                 Database file path (default: "urls.db")
--sync-interval           Cache sync interval (default: 5s)

# Metrics options
--metrics-enabled         Enable Prometheus metrics (default: true)
--metrics-port            Metrics server port (default: "9090")
--metrics-endpoint        Metrics endpoint path (default: "/metrics")

# Shortener algorithm options
--shortener-type          Algorithm type: "md5", "base62_counter", "base62_random", "nanoid" (default: "md5")
--shortener-length        Generated code length for applicable algorithms
--shortener-counter-step  Jump-ahead step size for base62_counter algorithm (default: 100)
```

## Development

### Available Commands

```bash
make build                # Build the application
make test                 # Run all tests
make test-unit           # Run unit tests only
make test-integration    # Run integration tests only
make test-coverage       # Generate coverage report
make lint                # Lint code
make fmt                 # Format code
make clean               # Clean build artifacts
```

### Architecture

```
url-shortener/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/          # Domain models and entities
│   ├── repository/      # Data access layer (SQLite with sqlc)
│   ├── cache/           # Caching layer (Memory implementation)
│   ├── service/         # Business logic layer
│   ├── shortener/       # URL shortening algorithms and generators
│   ├── metrics/         # Prometheus metrics collection
│   └── transport/       # Transport layer (HTTP server/client)
├── db/
│   ├── migrations/      # SQL migration files
│   ├── queries/         # SQL queries for sqlc
│   └── sqlc/           # Generated sqlc code
└── tests/              # Integration tests
```

### URL Generation Algorithms

The application supports multiple shortening algorithms through a pluggable generator interface:

1. **MD5 Hash** (default): Uses MD5 hash of URL + timestamp with base36-encoded timestamp
   - Format: `{timestamp_base36}{hash_first_4_chars}`
   - Deterministic and collision-resistant
   - No database dependencies for generation

2. **Base62 Counter**: Sequential counter-based encoding with jump-ahead allocation
   - Uses base62 encoding (0-9, A-Z, a-z)
   - Configurable length with padding
   - In-memory cache with async database writeback
   - Thread-safe with automatic counter synchronization
   - Best performance for high-throughput scenarios

3. **Base62 Random**: Cryptographically random base62 codes
   - Configurable length (default: 6 characters)
   - No database dependencies for generation
   - Collision detection and retry logic

4. **NanoID**: URL-safe random identifier generation
   - Uses alphabet: `_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
   - Configurable length (default: 21 characters)
   - Cryptographically strong randomness

Configure via CLI flags:
```bash
--shortener-type           # Algorithm type: "md5", "base62_counter", "base62_random", "nanoid"
--shortener-length         # Generated code length (where applicable)
--shortener-counter-step   # Counter jump-ahead step size for base62_counter
```

## Database

### Schema
- Uses sqlc for type-safe SQL queries
- Migration files in `db/migrations/`
- Query files in `db/queries/`
- Generated code in `db/sqlc/`

### Tables
- `urls` table with columns: id, short_code, original_url, created_at, last_used_at, usage_count
- `counters` table with columns: key, value, updated_at (for counter-based shortening algorithms)

## Monitoring

### Prometheus Metrics

When metrics are enabled (default), the application exposes comprehensive metrics at `http://localhost:9090/metrics`:

- **URL Operations**: Created, retrieved, deleted URLs with error counters
- **Cache Performance**: Hit/miss ratios, size, synchronization metrics
- **HTTP Performance**: Request counts, duration histograms by endpoint
- **Database Performance**: Query counts, duration, error rates
- **System Metrics**: Uptime, memory usage

### Health Check

```bash
curl http://localhost:9090/health
```

## Cache Implementation

### Memory Cache
- Thread-safe in-memory caching for URL data
- Background synchronization with database
- No external dependencies
- Automatic cache initialization on startup

### Counter Cache
- In-memory counter cache for Base62 counter algorithm
- Jump-ahead allocation to reduce database contention
- Async writeback to database for persistence
- Thread-safe with proper synchronization
- Configurable step size for batch counter allocation

## Testing

### Unit Tests
```bash
make test-unit
```
- Comprehensive mocks for all interfaces
- High coverage of business logic
- Fast execution for development

### Integration Tests
```bash
make test-integration
```
- End-to-end testing across all layers
- Database and cache integration
- HTTP API testing

### Coverage Report
```bash
make test-coverage
# Generates coverage.html
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linting: `make lint`
6. Submit a pull request

## License

[Add your license information here]