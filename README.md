# URL Shortener

A high-performance URL shortener service built with Go, featuring Redis caching, PostgreSQL persistence, and comprehensive analytics.

## Features

- ⚡ Fast URL shortening with Base62 encoding
- 🔄 Automatic redirect handling
- 📊 Comprehensive analytics and click tracking
- 💾 Redis caching for optimal performance
- 🐘 PostgreSQL for reliable data persistence
- 🐳 Docker containerization
- 🔒 Production-ready with security headers
- 🌐 Custom short code support
- 📈 RESTful API with JSON responses
- ⚖️ Rate limiting and abuse protection

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL (for production)
- Redis (for caching)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/URLshorter/url-shortener.git
   cd url-shortener
   ```

2. **Start development services**
   ```bash
   # On Windows
   .\scripts\setup-local.ps1
   
   # On Unix/Linux/macOS
   make dev
   ```

3. **Install Go dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

The service will be available at `http://localhost:8080`

## API Endpoints

### Shorten URL
```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://www.example.com",
  "custom_code": "optional-custom-code",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

Response:
```json
{
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://www.example.com",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Redirect
```http
GET /{shortCode}
```
Returns a 301 redirect to the original URL.

### Analytics
```http
GET /api/v1/analytics/{shortCode}?days=30
```

Response:
```json
{
  "short_code": "abc123",
  "original_url": "https://www.example.com",
  "total_clicks": 150,
  "created_at": "2024-01-01T00:00:00Z",
  "daily_clicks": [...],
  "country_stats": [...]
}
```

### Health Check
```http
GET /health
```

## Configuration

Environment variables (see `.env.example`):

```env
# Server
SERVER_PORT=8080
BASE_URL=http://localhost:8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=urlshortener
DB_PASSWORD=password
DB_NAME=urlshortener_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Snowflake ID
NODE_ID=1
```

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage
```

## Deployment

### Docker Compose (Recommended)

```bash
# Production deployment
make docker-run

# Or manually
docker-compose -f docker/docker-compose.yml up -d
```

### Manual Deployment

1. Build the binary:
   ```bash
   make build
   ```

2. Set up PostgreSQL and Redis

3. Configure environment variables

4. Run the binary:
   ```bash
   ./bin/url-shortener
   ```

## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │    │   Nginx     │    │   App       │
│             │───▶│  (Proxy)    │───▶│  (Go)       │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
                                              │
                                              ▼
                                      ┌─────────────┐
                                      │   Redis     │
                                      │  (Cache)    │
                                      └─────────────┘
                                              │
                                              ▼
                                      ┌─────────────┐
                                      │ PostgreSQL  │
                                      │(Persistence)│
                                      └─────────────┘
```

## Performance

- **Shortening**: ~1000 requests/second
- **Redirects**: ~5000 requests/second
- **Cache hit ratio**: 95%+ for popular URLs
- **Latency**: <10ms average response time

## Security Features

- Rate limiting (10 req/s for API, 30 req/s for redirects)
- Input validation and sanitization
- SQL injection prevention
- XSS protection headers
- CSRF protection
- IP-based analytics (anonymized)

## Monitoring

Health check endpoint available at `/health` for monitoring systems.

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues and questions, please use the GitHub issue tracker.