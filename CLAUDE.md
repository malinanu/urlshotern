# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a high-performance URL shortener service built with Go, featuring Redis caching, PostgreSQL persistence, and comprehensive analytics. The system uses Snowflake ID generation with Base62 encoding for short codes.

## Common Development Commands

```bash
# Development setup (Windows)
.\scripts\setup-local.ps1

# Development setup (Unix/Linux/macOS)  
make dev

# Build the application
make build

# Run the application locally
go run cmd/server/main.go
# or
make run

# Run tests
make test                # All tests
make test-unit          # Unit tests only
make test-integration   # Integration tests only
make test-coverage      # With coverage report

# Docker commands
make docker-build       # Build Docker image
make docker-run         # Start production stack
make docker-stop        # Stop all services

# Code quality
make lint               # Run linters
make format             # Format code
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
- **HTTP Framework**: Gin for high-performance HTTP routing
- **Testing**: Unit tests with testify, integration tests with httptest

## API Endpoints

- `POST /api/v1/shorten` - Create short URL
- `GET /{shortCode}` - Redirect to original URL  
- `GET /api/v1/analytics/{shortCode}` - Get analytics data
- `GET /health` - Health check

## Environment Configuration

Copy `.env.example` to `.env` and configure:
- Database connection (PostgreSQL)
- Redis connection
- Server settings (port, base URL)
- Snowflake node ID for distributed deployment

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