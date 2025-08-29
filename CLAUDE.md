# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A modern, well-structured URL shortening service written in Go with clean architecture, featuring SQLite database backend and in-memory caching. The application provides both server and client functionality with comprehensive testing.

## Architecture

The application follows clean architecture principles with clear separation of concerns:

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
│   └── transport/       # Transport layer (HTTP server/client)
├── db/
│   ├── migrations/      # SQL migration files
│   ├── queries/         # SQL queries for sqlc
│   └── sqlc/           # Generated sqlc code
└── tests/              # Integration tests
```

### Key Components

- **Repository Layer**: SQLite with sqlc-generated type-safe queries
- **Cache Layer**: Memory cache implementation with background sync
- **Service Layer**: Core business logic with proper error handling
- **Shortener Layer**: Pluggable URL shortening algorithms with generator interface
- **Transport Layer**: HTTP server with RESTful API and CLI client
- **Configuration**: CLI argument-based configuration

### URL Generation

The application supports multiple shortening algorithms through a pluggable generator interface:

1. **MD5 Hash** (default): Uses MD5 hash of URL + timestamp with base36-encoded timestamp
   - Format: `{timestamp_base36}{hash_first_4_chars}`
   - Deterministic and collision-resistant

2. **Base62 Counter**: Sequential counter-based encoding with jump-ahead allocation
   - Thread-safe with in-memory counter cache
   - Async writeback to database for persistence
   - Best performance for high-throughput scenarios

3. **Base62 Random**: Cryptographically random base62 codes
   - Configurable length (default: 6 characters)
   - No database dependencies for generation

4. **NanoID**: URL-safe random identifier generation
   - Configurable length (default: 21 characters)
   - Cryptographically strong randomness

## Development Commands

### Build and Test
```bash
make build                               # Build the application
make install                             # Install binary to GOPATH/bin
make test                                # Run all tests
make test-unit                           # Run unit tests only
make test-integration                    # Run integration tests only
make test-coverage                       # Generate coverage report
make bench                               # Run benchmarks
make clean                               # Clean build artifacts
```

### Run Application
```bash
make run-server                          # Build and run server
make run-example                         # Run with example config
```

### Development Setup
```bash
make dev-setup                           # Install tools, tidy deps, generate code
make install-tools                       # Install development tools
make generate                            # Generate sqlc code
make fmt                                 # Format code
make lint                               # Lint code (requires golangci-lint)
make tidy                               # Tidy dependencies
```

### Direct Go Commands
```bash
# Start server with default settings
go run ./cmd/server server

# Start server with custom settings
go run ./cmd/server server --port 9000 --db-path custom.db

# Client commands
go run ./cmd/server client create "https://example.com"
go run ./cmd/server client get <short_code>
go run ./cmd/server client list
go run ./cmd/server client delete <short_code>
```

### Server Configuration Options

```bash
# Server command flags
--port, -p                 Server port (default: "8080")
--server-url              Server URL for client communication (default: "http://localhost:8080")
--db-path                 Database file path (default: "urls.db")
--sync-interval           Cache sync interval (default: 5s)
--shortener-type          Algorithm type: "md5", "base62_counter", "base62_random", "nanoid" (default: "md5")
--shortener-length        Generated code length for applicable algorithms
--shortener-counter-step  Jump-ahead step size for base62_counter algorithm (default: 100)
```

## Configuration

Configuration is provided via CLI arguments as shown above. All options have sensible defaults for development and production use.

## API Endpoints

- `POST /api/urls` - Create short URL
- `GET /api/urls` - List all URLs  
- `GET /api/urls/{code}` - Get URL info
- `DELETE /api/urls/{code}` - Delete URL
- `GET /{code}` - Redirect to original URL

## Database

### Schema
- Uses sqlc for type-safe SQL queries
- Migration files in `db/migrations/`
- Query files in `db/queries/`
- Generated code in `db/sqlc/`

### Tables
- `urls` table with columns: id, short_code, original_url, created_at, last_used_at, usage_count
- `counters` table with columns: key, value, updated_at (for counter-based shortening algorithms)

## Testing

### Unit Tests
- Comprehensive mocks for all interfaces
- Test files alongside source code (`*_test.go`)
- Use `make test-unit` to run

### Integration Tests
- End-to-end testing in `tests/integration/`
- Tests database, cache, and HTTP layers together
- Use `make test-integration` to run

### Test Coverage
- Generate HTML coverage reports with `make test-coverage`

## Cache Implementation

### Memory Cache
- Thread-safe in-memory caching for URL data
- Background sync to database
- No external dependencies
- Automatic cache initialization on startup

### Counter Cache
- In-memory counter cache for Base62 counter algorithm
- Jump-ahead allocation to reduce database contention
- Async writeback to database for persistence
- Thread-safe with proper synchronization
- Configurable step size for batch counter allocation


## Development Tips

- Use interfaces extensively for testability
- All database operations use context for cancellation
- Background cache sync runs independently
- Graceful shutdown handles cleanup properly
- Configuration validation prevents startup with invalid config
