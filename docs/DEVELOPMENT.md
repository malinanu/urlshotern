# Development Guide

This guide explains how to set up and run the URL Shortener application in development mode.

## Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local backend development)
- Node.js 18+ and npm (for local frontend development)
- PostgreSQL 15+ (if running locally without Docker)
- Redis 7+ (if running locally without Docker)

## Development Modes

The project supports two development modes, each optimized for different workflows:

### Mode 1: Docker-Only (Monolithic) 🐳

**Best for**: Backend development, testing the full stack, production-like environment

This mode runs everything in Docker containers with the frontend pre-built and served as static files by the Go backend.

#### Quick Start

```bash
# Windows
.\scripts\setup-local.ps1

# Linux/macOS
make dev-docker
```

#### What It Does

1. Builds the Docker image with Node.js + Go
2. Installs all dependencies (Go modules + npm packages)
3. Builds the Next.js frontend to static files
4. Copies static files to `/app/static/`
5. Starts the Go backend with hot-reloading via `air`
6. Runs database migrations automatically
7. Serves frontend from `http://localhost:8080`

#### When to Use

- ✅ Testing the complete production-like stack
- ✅ Backend-focused development
- ✅ CI/CD pipeline validation
- ✅ Quick demos without multiple terminal windows
- ✅ When you want everything to "just work" with one command

#### Workflow

```bash
# Start everything
docker-compose -f docker/docker-compose.local.yml up --build

# Access the application
open http://localhost:8080

# Access database UI (Adminer)
open http://localhost:8081

# View logs
docker-compose -f docker/docker-compose.local.yml logs -f app

# Stop everything
docker-compose -f docker/docker-compose.local.yml down

# Clean up (including volumes)
docker-compose -f docker/docker-compose.local.yml down -v
```

#### Important Notes

- Frontend changes require rebuilding the Docker image: `docker-compose up --build`
- Go code changes are hot-reloaded automatically via `air`
- Database schema changes require running migrations manually

---

### Mode 2: Separate Dev Servers (Modern) ⚡ **RECOMMENDED**

**Best for**: Frontend development, rapid iteration, modern Next.js workflow

This mode runs the Next.js dev server separately with hot module replacement (HMR) for instant feedback on frontend changes.

#### Quick Start

```bash
# Terminal 1: Start backend services
make dev-backend

# Terminal 2: Start frontend dev server
make dev-frontend
```

#### What It Does

1. **Backend** (Terminal 1):
   - Runs Go backend on `http://localhost:8080`
   - Connects to PostgreSQL and Redis (in Docker or local)
   - Serves API endpoints

2. **Frontend** (Terminal 2):
   - Runs Next.js dev server on `http://localhost:3000`
   - Enables hot module replacement (HMR)
   - Provides instant feedback for UI changes
   - Uses React Fast Refresh for component updates

#### When to Use

- ✅ **Active frontend development** (UI, components, styling)
- ✅ Rapid iteration with instant hot-reloading
- ✅ Debugging frontend with React DevTools
- ✅ Testing responsive design and interactions
- ✅ Working with Next.js-specific features

#### Detailed Setup

##### Option A: Backend in Docker (Recommended)

```bash
# Terminal 1: Start database and Redis
docker-compose -f docker/docker-compose.local.yml up postgres redis

# Terminal 2: Run Go backend locally
go run cmd/server/main.go

# Terminal 3: Run frontend dev server
cd frontend
npm run dev
```

##### Option B: Backend Locally (Fastest)

```bash
# Ensure PostgreSQL and Redis are running (locally or Docker)

# Terminal 1: Run Go backend
go run cmd/server/main.go

# Terminal 2: Run frontend dev server
cd frontend
npm run dev
```

#### Access Points

- Frontend (Next.js): `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Database (Adminer): `http://localhost:8081` (if using Docker)

#### How It Works

1. Frontend makes API calls to `http://localhost:8080` (configured in `.env.development`)
2. Go backend has CORS enabled for `http://localhost:3000` in development mode
3. Frontend dev server proxies API requests to the backend
4. Changes to React components hot-reload instantly without page refresh

#### Important Notes

- Frontend runs on port **3000**, backend on port **8080**
- CORS is automatically configured for development
- Environment variables are loaded from `frontend/.env.development`
- API calls use `NEXT_PUBLIC_API_URL=http://localhost:8080`

---

## Comparison: Which Mode to Use?

| Feature | Mode 1 (Docker-Only) | Mode 2 (Separate Servers) |
|---------|---------------------|---------------------------|
| **Setup Complexity** | Simple (one command) | Moderate (multiple terminals) |
| **Frontend Hot-Reload** | ❌ No (rebuild required) | ✅ Yes (instant HMR) |
| **Backend Hot-Reload** | ✅ Yes (via air) | ✅ Yes (via air or go run) |
| **Production-Like** | ✅ Yes | ⚠️ Partially |
| **Frontend Dev Speed** | 🐌 Slow (rebuild required) | ⚡ Fast (instant) |
| **Best For** | Backend work, full stack testing | Frontend work, UI development |
| **Startup Time** | Medium (3-5 min first build) | Fast (30 sec) |
| **Resource Usage** | Higher (Docker overhead) | Lower (native processes) |

**Recommendation**:
- Use **Mode 2** for daily frontend development
- Use **Mode 1** for backend work and full-stack testing

---

## Environment Variables

### Backend Environment Variables

Located in `.env` or configured in docker-compose files:

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=urlshortener
DB_PASSWORD=devpassword
DB_NAME=urlshortener_db

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Server
SERVER_PORT=8080
BASE_URL=http://localhost:8080
GO_ENV=development

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_ISSUER=url-shortener

# Polar (Payment Gateway)
POLAR_ACCESS_TOKEN=polar_oat_...
POLAR_WEBHOOK_SECRET=whsec_...
POLAR_ORGANIZATION_ID=...
POLAR_API_URL=https://sandbox-api.polar.sh/v1
```

### Frontend Environment Variables

Located in `frontend/.env.development`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_BASE_URL=http://localhost:3000
```

---

## Common Tasks

### Database Management

```bash
# Check migration status
make migrate-status

# Apply specific migration
make migrate-up VERSION=001

# Rollback migration
make migrate-down VERSION=005

# Access database via Adminer
open http://localhost:8081
# Server: postgres
# User: urlshortener
# Password: devpassword
# Database: urlshortener_db
```

### Testing

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

### Code Quality

```bash
# Format code
make format

# Run linters
make lint

# Run security checks
make security-check
```

### Debugging

#### Debug Backend

```bash
# Run with verbose logging
GO_ENV=development go run cmd/server/main.go

# Use delve debugger
dlv debug cmd/server/main.go
```

#### Debug Frontend

```bash
# Run with debug mode
cd frontend
npm run dev

# Open browser DevTools (F12)
# Use React DevTools extension for component inspection
```

---

## Troubleshooting

### Frontend Not Loading (Mode 1)

**Problem**: Accessing `http://localhost:8080` shows errors or blank page

**Solution**:
```bash
# Rebuild Docker image
docker-compose -f docker/docker-compose.local.yml up --build

# Check if static files exist in container
docker-compose -f docker/docker-compose.local.yml exec app ls -la /app/static

# View entrypoint logs for build errors
docker-compose -f docker/docker-compose.local.yml logs app | grep "Frontend"
```

### CORS Errors (Mode 2)

**Problem**: Frontend can't connect to backend, CORS errors in browser console

**Solution**:
```bash
# Ensure GO_ENV=development is set
export GO_ENV=development  # Linux/macOS
$env:GO_ENV="development"  # Windows PowerShell

# Check backend CORS configuration in internal/routes/routes.go
# Verify frontend is running on http://localhost:3000
# Check NEXT_PUBLIC_API_URL in frontend/.env.development
```

### Database Connection Failed

**Problem**: Backend can't connect to PostgreSQL

**Solution**:
```bash
# Check if PostgreSQL is running
docker-compose -f docker/docker-compose.local.yml ps postgres

# Check database logs
docker-compose -f docker/docker-compose.local.yml logs postgres

# Test connection manually
psql -h localhost -U urlshortener -d urlshortener_db
# Password: devpassword
```

### Port Already in Use

**Problem**: Port 8080, 3000, 5432, or 6379 already in use

**Solution**:
```bash
# Find process using port (Linux/macOS)
lsof -i :8080

# Find process using port (Windows)
netstat -ano | findstr :8080

# Kill the process or change port in docker-compose.yml
```

### Frontend Build Fails in Docker

**Problem**: Docker build fails at `npm run build` step

**Solution**:
```bash
# Check Node.js version in container
docker-compose -f docker/docker-compose.local.yml run app node --version

# Manually build frontend to see errors
cd frontend
npm ci
npm run build

# Check next.config.ts for configuration issues
# Ensure output: 'export' is set
```

### Hot-Reload Not Working (Mode 1)

**Problem**: Go code changes don't reload automatically

**Solution**:
```bash
# Check if air is running
docker-compose -f docker/docker-compose.local.yml exec app ps aux | grep air

# Check .air.toml configuration
# Ensure volumes are mounted correctly in docker-compose.local.yml
```

---

## Best Practices

### Development Workflow

1. **Start of Day**:
   ```bash
   # Mode 2 (Frontend work)
   make dev-backend
   make dev-frontend

   # Mode 1 (Backend work)
   make dev-docker
   ```

2. **Making Changes**:
   - Frontend changes: Edit in `frontend/src/`, hot-reload happens instantly (Mode 2)
   - Backend changes: Edit in `internal/`, `air` reloads automatically (both modes)
   - Database schema: Create migration, run `make migrate-up VERSION=XXX`

3. **Before Committing**:
   ```bash
   make test
   make lint
   make format
   ```

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes, test thoroughly

# Run quality checks
make test
make lint

# Commit with descriptive message
git add .
git commit -m "feat: Add user authentication"

# Push and create PR
git push origin feature/your-feature-name
```

---

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Next.js Documentation](https://nextjs.org/docs)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)

---

## Getting Help

If you encounter issues not covered in this guide:

1. Check the main `README.md` for project overview
2. Review `CLAUDE.md` for architecture details
3. Check Docker logs: `docker-compose logs -f`
4. Search existing GitHub issues
5. Create a new issue with:
   - Development mode used (Mode 1 or Mode 2)
   - Steps to reproduce
   - Error messages and logs
   - Environment details (OS, Docker version, etc.)
