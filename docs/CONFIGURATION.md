# Configuration Reference

This document provides a comprehensive reference for all configuration options available in the URL Shortener application.

## Table of Contents

- [Environment Variables](#environment-variables)
  - [Server Configuration](#server-configuration)
  - [Database Configuration](#database-configuration)
  - [Redis Configuration](#redis-configuration)
  - [Authentication & Security](#authentication--security)
  - [Email Configuration](#email-configuration)
  - [CDN Configuration](#cdn-configuration)
  - [Logging Configuration](#logging-configuration)
  - [Polar Integration](#polar-integration)
- [Configuration Files](#configuration-files)
- [Environment-Specific Examples](#environment-specific-examples)
- [Configuration Best Practices](#configuration-best-practices)
- [Validation](#validation)

---

## Environment Variables

All configuration is managed through environment variables. These can be set in:
- `.env` file for local development
- System environment variables
- Docker environment files
- Kubernetes ConfigMaps and Secrets

### Server Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `SERVER_PORT` | string | No | `8080` | Port the server listens on |
| `SERVER_HOST` | string | No | `localhost` | Host the server binds to (use `0.0.0.0` for Docker) |
| `BASE_URL` | string | Yes | `http://localhost:8080` | Base URL for short link generation |
| `ENVIRONMENT` | string | No | `development` | Environment name (`development`, `staging`, `production`) |

**Examples:**
```bash
# Development
SERVER_PORT=8080
SERVER_HOST=localhost
BASE_URL=http://localhost:8080
ENVIRONMENT=development

# Production
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
BASE_URL=https://yourdomain.com
ENVIRONMENT=production
```

---

### Database Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `DB_HOST` | string | Yes | `localhost` | PostgreSQL host address |
| `DB_PORT` | string | Yes | `5432` | PostgreSQL port |
| `DB_USER` | string | Yes | `urlshortener` | Database username |
| `DB_PASSWORD` | string | Yes | `password` | Database password (use secrets management in production) |
| `DB_NAME` | string | Yes | `urlshortener_db` | Database name |
| `DB_MAX_CONNECTIONS` | int | No | `25` | Maximum number of open database connections |
| `DB_SSLMODE` | string | No | `disable` | SSL mode (`disable`, `require`, `verify-ca`, `verify-full`) |

**Security Notes:**
- Never commit `DB_PASSWORD` to version control
- Use `DB_SSLMODE=require` or higher in production
- For AWS RDS, use `DB_SSLMODE=verify-full` with proper CA certificates
- Consider using connection pooling services like PgBouncer for high-traffic applications

**Examples:**
```bash
# Local Development
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=urlshortener_dev
DB_MAX_CONNECTIONS=25
DB_SSLMODE=disable

# Production (AWS RDS)
DB_HOST=mydb.abc123.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=app_user
DB_PASSWORD=${DB_PASSWORD_SECRET}  # From secrets manager
DB_NAME=urlshortener_prod
DB_MAX_CONNECTIONS=100
DB_SSLMODE=verify-full

# Production (Heroku Postgres)
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
```

**Connection String Format:**
```
postgres://username:password@host:port/database?sslmode=require
```

---

### Redis Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `REDIS_HOST` | string | Yes | `localhost` | Redis host address |
| `REDIS_PORT` | string | Yes | `6379` | Redis port |
| `REDIS_PASSWORD` | string | No | `""` | Redis password (empty for no auth) |
| `REDIS_DB` | int | No | `0` | Redis database number (0-15) |
| `REDIS_MAX_RETRIES` | int | No | `3` | Maximum connection retry attempts |
| `REDIS_POOL_SIZE` | int | No | `10` | Connection pool size |

**Examples:**
```bash
# Local Development (no password)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Production (ElastiCache)
REDIS_HOST=my-cluster.abc123.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD_SECRET}
REDIS_DB=0
REDIS_MAX_RETRIES=5
REDIS_POOL_SIZE=50

# Redis Cluster (multiple nodes)
REDIS_CLUSTER_NODES=node1:6379,node2:6379,node3:6379
REDIS_PASSWORD=${REDIS_PASSWORD_SECRET}
```

---

### Authentication & Security

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `JWT_SECRET` | string | Yes | `your-secret-key-change-this-in-production` | Secret key for JWT signing (min 32 chars) |
| `JWT_ISSUER` | string | No | `urlshortener` | JWT issuer claim |
| `JWT_EXPIRY` | duration | No | `15m` | Access token expiry duration |
| `REFRESH_TOKEN_EXPIRY` | duration | No | `7d` | Refresh token expiry duration |
| `GOOGLE_CLIENT_ID` | string | No | `""` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | string | No | `""` | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | string | No | `""` | Google OAuth redirect URL |
| `BCRYPT_COST` | int | No | `10` | Bcrypt hashing cost (10-14 recommended) |

**Security Requirements:**
- `JWT_SECRET` must be at least 32 characters long
- Use cryptographically secure random strings for secrets
- Rotate `JWT_SECRET` regularly (see [SECRETS_ROTATION.md](SECRETS_ROTATION.md))
- Never log or expose secrets in error messages

**Generating Secure Secrets:**
```bash
# Generate 64-character random secret
openssl rand -base64 48

# Or using Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"; "os"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

**Examples:**
```bash
# Development
JWT_SECRET=dev-secret-key-not-for-production
JWT_ISSUER=urlshortener-dev
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
BCRYPT_COST=10

# Production
JWT_SECRET=${JWT_SECRET_FROM_VAULT}  # From secrets manager
JWT_ISSUER=urlshortener-prod
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
BCRYPT_COST=12  # Higher cost for production

# With Google OAuth
GOOGLE_CLIENT_ID=123456789-abcdefg.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
GOOGLE_REDIRECT_URL=https://yourdomain.com/auth/google/callback
```

---

### Email Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `SMTP_HOST` | string | Conditional | `""` | SMTP server host |
| `SMTP_PORT` | int | Conditional | `587` | SMTP server port (587 for TLS, 465 for SSL) |
| `SMTP_USERNAME` | string | Conditional | `""` | SMTP authentication username |
| `SMTP_PASSWORD` | string | Conditional | `""` | SMTP authentication password |
| `FROM_EMAIL` | string | Yes | `noreply@urlshortener.com` | Sender email address |
| `FROM_NAME` | string | Yes | `URL Shortener` | Sender name |
| `BASE_URL` | string | Yes | `http://localhost:3000` | Base URL (already configured above, used for email links) |

**Email Providers:**

**Generic SMTP (Gmail, Outlook, etc.):**
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=${GMAIL_APP_PASSWORD}
FROM_EMAIL=your-email@gmail.com
FROM_NAME=URL Shortener
BASE_URL=https://yourdomain.com
```

**AWS SES:**
```bash
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=${SES_SMTP_USERNAME}
SMTP_PASSWORD=${SES_SMTP_PASSWORD}
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=URL Shortener
BASE_URL=https://yourdomain.com
```

**SendGrid (future implementation):**
```bash
SENDGRID_API_KEY=${SENDGRID_API_KEY}
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=URL Shortener
BASE_URL=https://yourdomain.com
```

**Development (Console Output):**
```bash
# Leave SMTP variables empty to use console provider
FROM_EMAIL=dev@localhost
FROM_NAME=URL Shortener Dev
BASE_URL=http://localhost:3000
```

---

### CDN Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `CDN_PROVIDER` | string | No | `""` | CDN provider (`cloudfront`, `fastly`, `cloudflare`) |
| `CLOUDFRONT_DISTRIBUTION_ID` | string | Conditional | `""` | AWS CloudFront distribution ID |
| `FASTLY_API_KEY` | string | Conditional | `""` | Fastly API key |
| `FASTLY_SERVICE_ID` | string | Conditional | `""` | Fastly service ID |
| `CLOUDFLARE_ZONE_ID` | string | Conditional | `""` | Cloudflare zone ID |
| `CLOUDFLARE_API_TOKEN` | string | Conditional | `""` | Cloudflare API token |

**Examples:**

**AWS CloudFront:**
```bash
CDN_PROVIDER=cloudfront
CLOUDFRONT_DISTRIBUTION_ID=E123EXAMPLEID
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY}
AWS_SECRET_ACCESS_KEY=${AWS_SECRET_KEY}
```

**Fastly:**
```bash
CDN_PROVIDER=fastly
FASTLY_API_KEY=${FASTLY_API_KEY}
FASTLY_SERVICE_ID=abc123service
```

**Cloudflare:**
```bash
CDN_PROVIDER=cloudflare
CLOUDFLARE_ZONE_ID=abc123zone
CLOUDFLARE_API_TOKEN=${CLOUDFLARE_TOKEN}
```

---

### Logging Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `LOG_LEVEL` | string | No | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `LOG_FORMAT` | string | No | `json` | Log format (`json`, `text`) |
| `LOG_OUTPUT` | string | No | `stdout` | Log output (`stdout`, `file`, `both`) |
| `LOG_FILE_PATH` | string | Conditional | `logs/app.log` | Log file path (if `LOG_OUTPUT=file` or `both`) |

**Log Levels:**
- `debug`: Detailed debugging information (verbose, use only in development)
- `info`: General informational messages (default for production)
- `warn`: Warning messages that don't stop execution
- `error`: Error messages for failures

**Examples:**
```bash
# Development (verbose logging)
LOG_LEVEL=debug
LOG_FORMAT=text
LOG_OUTPUT=stdout

# Production (structured logging)
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=both
LOG_FILE_PATH=/var/log/urlshortener/app.log

# Production (cloud logging - stdout only)
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout  # Captured by CloudWatch, Stackdriver, etc.
```

---

### Polar Integration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `POLAR_ACCESS_TOKEN` | string | Conditional | `""` | Polar API access token |
| `POLAR_WEBHOOK_SECRET` | string | Conditional | `""` | Polar webhook secret for signature validation |
| `POLAR_ORGANIZATION_ID` | string | Conditional | `""` | Polar organization ID |
| `POLAR_API_URL` | string | No | `https://api.polar.sh/v1` | Polar API base URL |

**Examples:**
```bash
# With Polar integration
POLAR_ACCESS_TOKEN=${POLAR_ACCESS_TOKEN}
POLAR_WEBHOOK_SECRET=${POLAR_WEBHOOK_SECRET}
POLAR_ORGANIZATION_ID=org_abc123
POLAR_API_URL=https://api.polar.sh/v1

# Without Polar integration
# Leave variables empty or omit them
```

---

### Additional Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `NODE_ID` | int64 | No | `1` | Snowflake node ID for distributed deployment (1-1023) |
| `CORS_ALLOWED_ORIGINS` | string | No | `*` | Comma-separated list of allowed CORS origins |
| `RATE_LIMIT_REQUESTS` | int | No | `100` | Rate limit: requests per window |
| `RATE_LIMIT_WINDOW` | duration | No | `1m` | Rate limit time window |
| `ENABLE_METRICS` | bool | No | `true` | Enable Prometheus metrics endpoint |
| `METRICS_PORT` | int | No | `9090` | Prometheus metrics port |

**Examples:**
```bash
# Multi-instance deployment
NODE_ID=1  # Different for each instance (1, 2, 3, etc.)

# CORS configuration
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Rate limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1m

# Metrics
ENABLE_METRICS=true
METRICS_PORT=9090
```

---

## Configuration Files

### `.env` File

Create a `.env` file in the project root for local development:

```bash
# Copy from template
cp .env.example .env

# Edit with your values
nano .env
```

**Important:** Never commit `.env` files to version control. They are included in `.gitignore`.

### `.env.example` Template

The repository includes `.env.example` as a template with all available variables and their default values.

---

## Environment-Specific Examples

### Development Environment

**File: `.env.local`**
```bash
# Server
SERVER_PORT=8080
SERVER_HOST=localhost
BASE_URL=http://localhost:8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=urlshortener_dev
DB_MAX_CONNECTIONS=25
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Authentication
JWT_SECRET=dev-secret-key-min-32-characters-long
JWT_ISSUER=urlshortener-dev
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d

# Email (console output)
FROM_EMAIL=dev@localhost
FROM_NAME=URL Shortener Dev
BASE_URL=http://localhost:3000

# Logging
LOG_LEVEL=debug
LOG_FORMAT=text
LOG_OUTPUT=stdout

# Snowflake
NODE_ID=1
```

---

### Staging Environment

**File: `.env.staging`**
```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
BASE_URL=https://staging.yourdomain.com
ENVIRONMENT=staging

# Database (managed PostgreSQL)
DB_HOST=staging-db.yourdomain.com
DB_PORT=5432
DB_USER=app_user
DB_PASSWORD=${DB_PASSWORD_SECRET}
DB_NAME=urlshortener_staging
DB_MAX_CONNECTIONS=50
DB_SSLMODE=require

# Redis (managed Redis)
REDIS_HOST=staging-redis.yourdomain.com
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD_SECRET}
REDIS_DB=0
REDIS_POOL_SIZE=25

# Authentication
JWT_SECRET=${JWT_SECRET_FROM_VAULT}
JWT_ISSUER=urlshortener-staging
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
BCRYPT_COST=11

# Email (SMTP)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=${SENDGRID_API_KEY}
FROM_EMAIL=staging@yourdomain.com
FROM_NAME=URL Shortener Staging
BASE_URL=https://staging.yourdomain.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout

# Metrics
ENABLE_METRICS=true
METRICS_PORT=9090

# Snowflake
NODE_ID=1
```

---

### Production Environment

**File: `.env.production`**
```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
BASE_URL=https://yourdomain.com
ENVIRONMENT=production

# Database (AWS RDS)
DB_HOST=prod-db.abc123.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=app_user
DB_PASSWORD=${DB_PASSWORD_SECRET}  # From AWS Secrets Manager
DB_NAME=urlshortener_prod
DB_MAX_CONNECTIONS=100
DB_SSLMODE=verify-full

# Redis (AWS ElastiCache)
REDIS_HOST=prod-redis.abc123.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD_SECRET}
REDIS_DB=0
REDIS_POOL_SIZE=50
REDIS_MAX_RETRIES=5

# Authentication
JWT_SECRET=${JWT_SECRET_FROM_VAULT}  # From secrets manager
JWT_ISSUER=urlshortener-prod
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
BCRYPT_COST=12

# Google OAuth
GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
GOOGLE_REDIRECT_URL=https://yourdomain.com/auth/google/callback

# Email (AWS SES)
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USERNAME=${SES_SMTP_USERNAME}
SMTP_PASSWORD=${SES_SMTP_PASSWORD}
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=URL Shortener
BASE_URL=https://yourdomain.com

# CDN (CloudFront)
CDN_PROVIDER=cloudfront
CLOUDFRONT_DISTRIBUTION_ID=${CLOUDFRONT_DIST_ID}
AWS_REGION=us-east-1

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout  # CloudWatch captures stdout

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1m

# Metrics
ENABLE_METRICS=true
METRICS_PORT=9090

# Snowflake (multi-instance)
NODE_ID=${INSTANCE_NODE_ID}  # 1, 2, 3, etc. for each instance

# Polar Integration
POLAR_ACCESS_TOKEN=${POLAR_ACCESS_TOKEN}
POLAR_WEBHOOK_SECRET=${POLAR_WEBHOOK_SECRET}
POLAR_ORGANIZATION_ID=${POLAR_ORG_ID}
```

---

### Docker Compose Environment

**File: `docker/.env`**
```bash
# Server
SERVER_PORT=8080
BASE_URL=http://localhost:8080

# Database (Docker service)
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=urlshortener

# Redis (Docker service)
REDIS_HOST=redis
REDIS_PORT=6379

# Authentication
JWT_SECRET=docker-dev-secret-key-min-32-chars

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Configuration Best Practices

### Security

1. **Never commit secrets to version control**
   - Use `.env.example` as a template (no actual values)
   - Add `.env*` to `.gitignore` (except `.env.example`)
   - Use secrets management services in production

2. **Use strong secrets**
   - JWT secrets: minimum 32 characters, cryptographically random
   - Database passwords: 16+ characters with mixed case, numbers, symbols
   - API keys: use provider-generated keys, never custom values

3. **Rotate secrets regularly**
   - Follow [SECRETS_ROTATION.md](SECRETS_ROTATION.md) procedures
   - Set reminders for quarterly rotation
   - Document rotation procedures for your team

4. **Use environment-specific secrets**
   - Different secrets for dev, staging, production
   - Never use production secrets in development
   - Never use development secrets in production

### Configuration Management

1. **Use environment variables, not config files**
   - Easy to change without code deployment
   - Works with container orchestration
   - Integrates with secrets managers

2. **Validate configuration on startup**
   - Check required variables are set
   - Validate formats (URLs, connection strings)
   - Fail fast with clear error messages

3. **Document all configuration**
   - Keep this document up to date
   - Add new variables to `.env.example`
   - Document valid values and ranges

4. **Use defaults wisely**
   - Secure defaults (e.g., `DB_SSLMODE=require`)
   - Development-friendly defaults for optional settings
   - No defaults for secrets (require explicit setting)

### Secrets Management

**Recommended Tools:**

- **AWS Secrets Manager**: For AWS deployments
- **HashiCorp Vault**: Cross-platform secrets management
- **Kubernetes Secrets**: For K8s deployments
- **Azure Key Vault**: For Azure deployments
- **GCP Secret Manager**: For Google Cloud deployments

**Example with AWS Secrets Manager:**
```bash
# Fetch secret from AWS Secrets Manager
export JWT_SECRET=$(aws secretsmanager get-secret-value \
  --secret-id prod/urlshortener/jwt-secret \
  --query SecretString \
  --output text)
```

**Example with HashiCorp Vault:**
```bash
# Fetch secret from Vault
export DB_PASSWORD=$(vault kv get -field=password secret/urlshortener/db)
```

---

## Validation

The application validates configuration on startup. Missing required variables or invalid values will cause startup to fail with clear error messages.

**Required Variables:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `REDIS_HOST`, `REDIS_PORT`
- `JWT_SECRET` (minimum 32 characters)
- `BASE_URL`

**Validation Checks:**
- JWT secret length (minimum 32 characters)
- Database connection (connectivity test)
- Redis connection (ping test)
- Email configuration (if SMTP configured)
- Port availability

**Example Startup Validation:**
```bash
go run cmd/server/main.go

# Output on validation failure:
# ERROR: JWT_SECRET must be at least 32 characters long
# ERROR: Failed to connect to database: connection refused
# ERROR: SMTP_HOST is required when using SMTP provider
```

---

## Troubleshooting

### Configuration Not Loading

**Problem:** Environment variables not being read

**Solutions:**
1. Check `.env` file is in the project root
2. Verify no syntax errors in `.env` (no spaces around `=`)
3. Restart the application after changing `.env`
4. Check file permissions (`.env` should be readable)

### Database Connection Fails

**Problem:** Cannot connect to database

**Solutions:**
1. Verify `DB_HOST`, `DB_PORT` are correct
2. Check `DB_USER`, `DB_PASSWORD` credentials
3. Ensure database exists (`DB_NAME`)
4. Check network connectivity: `telnet $DB_HOST $DB_PORT`
5. Review database logs for connection rejections

### Redis Connection Fails

**Problem:** Cannot connect to Redis

**Solutions:**
1. Verify Redis is running: `redis-cli ping`
2. Check `REDIS_HOST`, `REDIS_PORT` are correct
3. Test password: `redis-cli -a $REDIS_PASSWORD ping`
4. Check network connectivity
5. Review Redis logs

### Email Not Sending

**Problem:** Emails not being delivered

**Solutions:**
1. Check SMTP credentials (`SMTP_USERNAME`, `SMTP_PASSWORD`)
2. Verify SMTP host and port
3. Check firewall rules (allow outbound on SMTP port)
4. Verify sender email is authorized
5. Check spam folder
6. Review application logs for email errors

For more troubleshooting help, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

---

## Related Documentation

- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues and solutions
- [Secret Rotation](SECRETS_ROTATION.md) - Rotating credentials and secrets
- [Deployment Guides](DEPLOYMENT_GUIDES/) - Cloud-specific deployment instructions
- [Security Best Practices](SECURITY.md) - Security guidelines

---

**Last Updated:** 2025-12-19
**Version:** 1.0.0
