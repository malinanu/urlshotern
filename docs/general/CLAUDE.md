# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a high-performance URL shortener service built with Go, featuring Redis caching, PostgreSQL persistence, and comprehensive analytics. The system uses Snowflake ID generation with Base62 encoding for short codes.

## Common Development Commands

```bash
# Development with Docker (Recommended - Hot Reload Enabled)
docker-compose -f docker/docker-compose.local.yml up --build  # First time only
docker-compose -f docker/docker-compose.local.yml up          # Daily development

# Access points:
# - Frontend (Main):  http://localhost:3000  (Next.js with HMR)
# - Backend API:      http://localhost:8080  (Go with air hot reload)
# - Adminer (DB UI):  http://localhost:8081  (Database management)

# Docker commands
make docker-build       # Build production Docker image
make docker-run         # Start production stack
make docker-stop        # Stop all services

# Development setup (Alternative - without Docker)
.\scripts\setup-local.ps1  # Windows
make dev                   # Unix/Linux/macOS

# Build the application
make build

# Run the application locally (without Docker)
go run cmd/server/main.go
# or
make run

# Run tests
make test                # All tests
make test-unit          # Unit tests only
make test-integration   # Integration tests only
make test-coverage      # With coverage report

# Code quality
make lint               # Run linters
make format             # Format code

# Database migrations
make migrate-status                    # Check migration status
make migrate-up VERSION=001            # Apply migration
make migrate-down VERSION=005          # Rollback migration
make migrate-rollback-to VERSION=003   # Rollback to version
make migrate-dry-run DIRECTION=down VERSION=005  # Preview rollback

# Or use scripts directly (cross-platform)
./scripts/migrate.sh status            # Linux/macOS
.\scripts\migrate.ps1 status           # Windows PowerShell

# Database maintenance (add missing tables, verify schema)
.\scripts\apply-missing-tables.ps1 -DryRun   # Preview changes (Windows)
.\scripts\apply-missing-tables.ps1           # Apply missing tables (Windows)
.\scripts\apply-missing-tables.ps1 -Verify   # Verify database schema (Windows)
./scripts/apply-missing-tables.sh --dry-run  # Preview changes (Linux/macOS)
./scripts/apply-missing-tables.sh            # Apply missing tables (Linux/macOS)
./scripts/apply-missing-tables.sh --verify   # Verify database schema (Linux/macOS)
```

## Architecture Overview

```
cmd/server/main.go              # Application entry point
internal/
├── handlers/handlers.go        # HTTP request handlers (Gin)
├── services/
│   ├── shortener.go           # Core URL shortening logic
│   └── analytics.go           # Analytics and reporting
├── storage/
│   ├── postgres.go            # PostgreSQL database layer
│   └── redis.go               # Redis caching layer
├── models/url.go              # Data models and DTOs
└── utils/
    ├── base62.go              # Base62 encoding/decoding
    └── snowflake.go           # Snowflake ID generation
configs/
├── config.go                  # Configuration management
└── database.sql               # Database schema
```

## Key Implementation Details

- **ID Generation**: Uses Sony's Snowflake implementation for distributed unique IDs
- **Encoding**: Base62 encoding for short, URL-safe codes
- **Caching**: Redis for fast lookups with TTL-based expiration
- **Database**: PostgreSQL with proper indexing for performance
- **Connection Pooling**: PgBouncer for efficient database connection management
- **HTTP Framework**: Gin for high-performance HTTP routing
- **Testing**: Unit tests with testify, integration tests with httptest

## API Endpoints

- `POST /api/v1/shorten` - Create short URL
- `GET /{shortCode}` - Redirect to original URL
- `GET /api/v1/analytics/{shortCode}` - Get analytics data
- `GET /health` - Health check
- `GET /health/db-pool` - Database connection pool statistics
- `GET /health/goroutines` - Goroutine error statistics and monitoring

## Environment Configuration

Copy `.env.example` to `.env` and configure:
- Database connection (PostgreSQL)
- Database connection pool settings (max connections, idle connections, timeouts)
- Redis connection
- Server settings (port, base URL)
- Snowflake node ID for distributed deployment

See `docker/README_PGBOUNCER.md` for detailed connection pooling configuration and monitoring.

## Testing Strategy

- Unit tests for utilities (Base62, Snowflake)
- Service layer tests with mocked dependencies
- Integration tests for HTTP handlers
- Benchmark tests for performance validation

## Development Workflow

1. Start services: `.\scripts\setup-local.ps1` (Windows) or `make dev`
2. Run tests: `make test`
3. Start server: `go run cmd/server/main.go`
4. Test API with curl or Postman
5. Check database via Adminer at http://localhost:8081

## Production Deployment

Use `docker/docker-compose.yml` for production with:
- Nginx reverse proxy with rate limiting
- SSL termination
- Health checks
- Proper environment variable configuration