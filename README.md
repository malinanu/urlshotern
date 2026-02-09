# URL Shortener

A high-performance URL shortener service built with Go, featuring Redis caching, PostgreSQL persistence, and comprehensive analytics.

> **🚀 Need to Deploy? Having Issues?**
> - **Oracle Server Deployment:** See [START_HERE.md](START_HERE.md) for quick deployment guide
> - **Using Portainer?** See [PORTAINER_DEPLOYMENT_GUIDE.md](PORTAINER_DEPLOYMENT_GUIDE.md)
> - **Troubleshooting 404 errors?** See [DEPLOYMENT_FIX.md](DEPLOYMENT_FIX.md)
> - **Visual diagrams?** See [DEPLOYMENT_FLOW.md](DEPLOYMENT_FLOW.md)

> **Latest Update (2026-02-07)**: Security vulnerabilities fixed! All Docker and frontend dependencies updated. See [SECURITY_UPDATE.md](SECURITY_UPDATE.md) for details.

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

- Go 1.24 or later
- Node.js 18 or later (for frontend)
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

## Database Migrations

The project uses a custom migration system to manage database schema changes safely.

```bash
# Check migration status
make migrate-status

# Apply a migration
make migrate-up VERSION=001

# Rollback a migration
make migrate-down VERSION=005

# Rollback to a specific version
make migrate-rollback-to VERSION=003

# Preview rollback without executing
make migrate-dry-run DIRECTION=down VERSION=005

# Or use scripts directly
./scripts/migrate.sh status        # Linux/macOS
.\scripts\migrate.ps1 status       # Windows PowerShell
```

See [docs/MIGRATIONS.md](docs/MIGRATIONS.md) for the complete migration guide and [docs/ROLLBACK_PROCEDURES.md](docs/ROLLBACK_PROCEDURES.md) for rollback procedures.

## Deployment

### 🚀 Quick Deployment Guides

**Choose your deployment method:**

1. **Oracle Cloud Server (Portainer)** ⭐ Most Common
   - Start here: [START_HERE.md](START_HERE.md)
   - Portainer UI guide: [PORTAINER_DEPLOYMENT_GUIDE.md](PORTAINER_DEPLOYMENT_GUIDE.md)
   - Visual diagrams: [DEPLOYMENT_FLOW.md](DEPLOYMENT_FLOW.md)

2. **Troubleshooting Deployment Issues**
   - 404 errors on auth endpoints? → [DEPLOYMENT_FIX.md](DEPLOYMENT_FIX.md)
   - Quick reference card: [QUICK_FIX_CARD.md](QUICK_FIX_CARD.md)
   - Complete implementation details: [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)

3. **Testing Your Deployment**
   ```bash
   # Windows PowerShell
   .\test-auth-endpoints.ps1 "http://YOUR_SERVER:8082"

   # Linux/macOS
   ./test-auth-endpoints.sh http://YOUR_SERVER:8082
   ```

### Docker Compose (Recommended)

```bash
# Production deployment
make docker-run

# Or manually
docker-compose -f docker/docker-compose.yml up -d

# Oracle Cloud deployment
docker-compose -f docker-compose.oracle.yml up -d
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
- Regular security updates via Snyk scanning
- Automated vulnerability detection and patching

### Recent Security Updates

**February 2026 Security Update:**
- ✅ Updated Go to 1.24.1 (Docker base images)
- ✅ Fixed Alpine Linux vulnerabilities (BusyBox, OpenSSL)
- ✅ Updated Next.js to 16.1.6 (security patches)
- ✅ Updated React to 19.2.3 (security patches)
- ✅ Updated all frontend dependencies to latest secure versions
- ✅ Fixed CI/CD pipeline Go version mismatch

See [SECURITY_UPDATE.md](SECURITY_UPDATE.md) for complete details.

## Monitoring

Health check endpoints:
- `/health` - Basic health check
- `/health/db-pool` - Database connection pool statistics
- `/health/goroutines` - Goroutine error statistics and monitoring

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Documentation

- [CLAUDE.md](CLAUDE.md) - Development guidelines and common commands
- [SECURITY_UPDATE.md](SECURITY_UPDATE.md) - Latest security updates and fixes
- [docs/MIGRATIONS.md](docs/MIGRATIONS.md) - Database migration guide
- [docs/ROLLBACK_PROCEDURES.md](docs/ROLLBACK_PROCEDURES.md) - Rollback procedures
- [docker/README_PGBOUNCER.md](docker/README_PGBOUNCER.md) - Connection pooling guide

## Support

For issues and questions, please use the GitHub issue tracker.