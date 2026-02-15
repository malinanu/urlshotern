# Docker Compose Production Deployment Guide

This guide covers deploying the URL Shortener application using Docker Compose in a production environment with proper security, monitoring, and high availability configurations.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Production Architecture](#production-architecture)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [SSL/TLS Setup](#ssltls-setup)
- [Monitoring and Logging](#monitoring-and-logging)
- [Backup and Restore](#backup-and-restore)
- [Scaling](#scaling)
- [Security Hardening](#security-hardening)
- [Maintenance](#maintenance)
- [Troubleshooting](#troubleshooting)

---

## Overview

This production Docker Compose setup includes:
- **Application** containers with health checks and restart policies
- **PostgreSQL** database with persistent volumes and backups
- **Redis** cache with persistence enabled
- **Nginx** reverse proxy with SSL termination and rate limiting
- **Prometheus** and **Grafana** for monitoring
- **Automated backups** using cron jobs
- **Log aggregation** with centralized logging

---

## Prerequisites

### System Requirements
- Linux server (Ubuntu 20.04+ or similar)
- Minimum: 2 CPU cores, 4GB RAM, 50GB disk
- Recommended: 4 CPU cores, 8GB RAM, 100GB SSD
- Docker Engine 20.10+ and Docker Compose 2.0+

### Install Docker and Docker Compose

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose (v2)
sudo apt-get update
sudo apt-get install docker-compose-plugin

# Verify installation
docker --version
docker compose version
```

### Domain and DNS
- Registered domain name pointed to your server IP
- DNS records configured (A record for yourdomain.com)

---

## Production Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ :443 (HTTPS)
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Nginx Reverse Proxy                       │
│         (SSL Termination, Rate Limiting, Caching)           │
└──────────┬──────────────────────────────────┬───────────────┘
           │                                  │
┌──────────▼─────────┐              ┌────────▼──────────┐
│  App Container 1   │              │  App Container 2  │
│  (urlshortener)    │              │  (urlshortener)   │
└──────────┬─────────┘              └────────┬──────────┘
           │                                  │
           │      ┌──────────────────┐        │
           └──────▶  Redis Container ◀────────┘
                  │  (Cache + Session)│
                  └──────────────────┘
                           │
                  ┌────────▼─────────┐
                  │ PostgreSQL       │
                  │ (Persistent Data)│
                  └──────────────────┘
```

---

## Configuration

### Directory Structure

Create the following structure:

```
/opt/urlshortener/
├── docker-compose.prod.yml
├── .env.production
├── nginx/
│   ├── nginx.conf
│   ├── ssl/
│   │   ├── cert.pem
│   │   └── key.pem
│   └── conf.d/
│       └── default.conf
├── prometheus/
│   └── prometheus.yml
├── grafana/
│   └── dashboards/
├── backups/
│   ├── postgres/
│   └── redis/
└── logs/
    ├── nginx/
    ├── app/
    └── postgres/
```

### Environment Variables

Create `.env.production`:

```bash
# Application
ENVIRONMENT=production
SERVER_PORT=8080
BASE_URL=https://yourdomain.com

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=appuser
DB_PASSWORD=<SECURE_DB_PASSWORD>
DB_NAME=urlshortener_prod
DB_MAX_CONNECTIONS=100
DB_SSLMODE=disable  # Within Docker network

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=<SECURE_REDIS_PASSWORD>
REDIS_DB=0

# Authentication
JWT_SECRET=<SECURE_JWT_SECRET_MIN_32_CHARS>
JWT_ISSUER=urlshortener-prod
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
BCRYPT_COST=12

# Email (SMTP)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=<SENDGRID_API_KEY>
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=URL Shortener

# Note: Email links use BASE_URL configured above

# Google OAuth
GOOGLE_CLIENT_ID=<YOUR_GOOGLE_CLIENT_ID>
GOOGLE_CLIENT_SECRET=<YOUR_GOOGLE_CLIENT_SECRET>
GOOGLE_REDIRECT_URL=https://yourdomain.com/auth/google/callback

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout

# Snowflake (set different NODE_ID for each instance)
NODE_ID=1

# Monitoring
ENABLE_METRICS=true
METRICS_PORT=9090
```

**Security Note**: Never commit `.env.production` to version control. Use secure password generation:

```bash
# Generate secure passwords
openssl rand -base64 32  # For DB_PASSWORD
openssl rand -base64 48  # For JWT_SECRET
openssl rand -base64 32  # For REDIS_PASSWORD
```

---

## Docker Compose Configuration

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:14-alpine
    container_name: urlshortener-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./configs/database.sql:/docker-entrypoint-initdb.d/init.sql
      - ./backups/postgres:/backups
      - ./logs/postgres:/var/log/postgresql
    networks:
      - backend
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5
    command:
      - "postgres"
      - "-c"
      - "max_connections=200"
      - "-c"
      - "shared_buffers=256MB"
      - "-c"
      - "effective_cache_size=1GB"
      - "-c"
      - "work_mem=16MB"
      - "-c"
      - "maintenance_work_mem=128MB"
      - "-c"
      - "log_destination=stderr"
      - "-c"
      - "logging_collector=on"
      - "-c"
      - "log_directory=/var/log/postgresql"

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: urlshortener-redis
    restart: unless-stopped
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --maxmemory 512mb
      --maxmemory-policy allkeys-lru
      --appendonly yes
      --appendfsync everysec
      --save 900 1
      --save 300 10
      --save 60 10000
    volumes:
      - redis_data:/data
      - ./backups/redis:/backups
    networks:
      - backend
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Application Instance 1
  app1:
    build:
      context: .
      dockerfile: Dockerfile
    image: urlshortener:latest
    container_name: urlshortener-app1
    restart: unless-stopped
    env_file:
      - .env.production
    environment:
      - NODE_ID=1
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./logs/app:/app/logs
    networks:
      - backend
      - frontend
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Application Instance 2 (for load balancing)
  app2:
    build:
      context: .
      dockerfile: Dockerfile
    image: urlshortener:latest
    container_name: urlshortener-app2
    restart: unless-stopped
    env_file:
      - .env.production
    environment:
      - NODE_ID=2
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./logs/app:/app/logs
    networks:
      - backend
      - frontend
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Nginx Reverse Proxy
  nginx:
    image: nginx:alpine
    container_name: urlshortener-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./logs/nginx:/var/log/nginx
      - nginx_cache:/var/cache/nginx
    depends_on:
      - app1
      - app2
    networks:
      - frontend
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Prometheus Monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: urlshortener-prometheus
    restart: unless-stopped
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    networks:
      - monitoring
      - backend
    ports:
      - "9090:9090"

  # Grafana Dashboards
  grafana:
    image: grafana/grafana:latest
    container_name: urlshortener-grafana
    restart: unless-stopped
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=<SECURE_GRAFANA_PASSWORD>
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
    networks:
      - monitoring
    ports:
      - "3001:3000"
    depends_on:
      - prometheus

  # Backup Service (runs daily at 3 AM)
  backup:
    image: postgres:14-alpine
    container_name: urlshortener-backup
    restart: "no"
    environment:
      PGHOST: postgres
      PGUSER: ${DB_USER}
      PGPASSWORD: ${DB_PASSWORD}
      PGDATABASE: ${DB_NAME}
    volumes:
      - ./backups/postgres:/backups
      - ./scripts/backup.sh:/backup.sh:ro
    networks:
      - backend
    entrypoint: /bin/sh
    command: -c "crond -f -l 2"

networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
  monitoring:
    driver: bridge

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local
  prometheus_data:
    driver: local
  grafana_data:
    driver: local
  nginx_cache:
    driver: local
```

---

## Nginx Configuration

Create `nginx/conf.d/default.conf`:

```nginx
# Upstream backend servers
upstream app_backend {
    least_conn;
    server app1:8080 max_fails=3 fail_timeout=30s;
    server app2:8080 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

# Rate limiting zones
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;
limit_req_zone $binary_remote_addr zone=shorten_limit:10m rate=20r/m;

# Cache configuration
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=url_cache:10m max_size=1g inactive=60m use_temp_path=off;

# HTTP to HTTPS redirect
server {
    listen 80;
    listen [::]:80;
    server_name yourdomain.com www.yourdomain.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL configuration
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Logging
    access_log /var/log/nginx/access.log combined;
    error_log /var/log/nginx/error.log warn;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # API endpoints with rate limiting
    location /api/ {
        limit_req zone=api_limit burst=20 nodelay;

        proxy_pass http://app_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Shorten endpoint with stricter rate limiting
    location /api/v1/shorten {
        limit_req zone=shorten_limit burst=5 nodelay;

        proxy_pass http://app_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Short URL redirects with caching
    location ~ ^/[a-zA-Z0-9]{6,8}$ {
        proxy_pass http://app_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # Cache redirects
        proxy_cache url_cache;
        proxy_cache_valid 200 302 10m;
        proxy_cache_key "$scheme$request_method$host$request_uri";
        add_header X-Cache-Status $upstream_cache_status;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://app_backend;
        access_log off;
    }

    # Static files (if any)
    location /static/ {
        alias /var/www/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

---

## SSL/TLS Setup

### Option 1: Let's Encrypt (Recommended)

```bash
# Install Certbot
sudo apt-get install certbot

# Obtain certificate
sudo certbot certonly --standalone \
  -d yourdomain.com \
  -d www.yourdomain.com \
  --email your-email@example.com \
  --agree-tos \
  --non-interactive

# Copy certificates to nginx directory
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /opt/urlshortener/nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /opt/urlshortener/nginx/ssl/key.pem

# Set up auto-renewal
sudo crontab -e
# Add: 0 3 * * * certbot renew --quiet && cp /etc/letsencrypt/live/yourdomain.com/*.pem /opt/urlshortener/nginx/ssl/ && docker compose -f /opt/urlshortener/docker-compose.prod.yml restart nginx
```

### Option 2: Self-Signed (Development Only)

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=yourdomain.com"
```

---

## Deployment

### Initial Deployment

```bash
# Navigate to project directory
cd /opt/urlshortener

# Create necessary directories
mkdir -p nginx/ssl logs/{nginx,app,postgres} backups/{postgres,redis}

# Set proper permissions
chmod 600 .env.production
chmod 600 nginx/ssl/*

# Pull latest images
docker compose -f docker-compose.prod.yml pull

# Build application image
docker compose -f docker-compose.prod.yml build

# Start services
docker compose -f docker-compose.prod.yml up -d

# Check status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

### Update Deployment

```bash
# Pull latest code
git pull origin main

# Rebuild images
docker compose -f docker-compose.prod.yml build --no-cache

# Rolling update (zero-downtime)
docker compose -f docker-compose.prod.yml up -d --no-deps --build app1
sleep 30
docker compose -f docker-compose.prod.yml up -d --no-deps --build app2

# Or restart all services
docker compose -f docker-compose.prod.yml restart
```

---

## Monitoring and Logging

### Prometheus Configuration

Create `prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'urlshortener'
    static_configs:
      - targets: ['app1:9090', 'app2:9090']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:5432']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']

  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx:9113']
```

### View Logs

```bash
# All services
docker compose -f docker-compose.prod.yml logs -f

# Specific service
docker compose -f docker-compose.prod.yml logs -f app1

# Application logs
tail -f logs/app/app.log

# Nginx access logs
tail -f logs/nginx/access.log

# PostgreSQL logs
tail -f logs/postgres/postgresql.log
```

---

## Backup and Restore

### Automated Backups

Create `scripts/backup.sh`:

```bash
#!/bin/bash
set -e

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# PostgreSQL backup
pg_dump -h postgres -U $PGUSER -d $PGDATABASE | gzip > "$BACKUP_DIR/db_backup_$DATE.sql.gz"

# Redis backup
redis-cli -h redis -a $REDIS_PASSWORD BGSAVE

# Cleanup old backups (keep last 30 days)
find $BACKUP_DIR -name "*.gz" -type f -mtime +30 -delete

echo "Backup completed: $DATE"
```

### Manual Backup

```bash
# Database
docker exec urlshortener-postgres pg_dump -U appuser urlshortener_prod | gzip > backups/postgres/manual_$(date +%Y%m%d).sql.gz

# Redis
docker exec urlshortener-redis redis-cli -a <PASSWORD> BGSAVE
docker cp urlshortener-redis:/data/dump.rdb backups/redis/
```

### Restore

```bash
# Restore database
gunzip < backups/postgres/db_backup_20251219.sql.gz | \
  docker exec -i urlshortener-postgres psql -U appuser -d urlshortener_prod

# Restore Redis
docker cp backups/redis/dump.rdb urlshortener-redis:/data/
docker compose -f docker-compose.prod.yml restart redis
```

---

## Scaling

### Horizontal Scaling

Add more application instances in `docker-compose.prod.yml`:

```yaml
  app3:
    # ... (copy app1 configuration)
    environment:
      - NODE_ID=3
```

Update nginx upstream:

```nginx
upstream app_backend {
    server app1:8080;
    server app2:8080;
    server app3:8080;
}
```

### Vertical Scaling

Set resource limits:

```yaml
services:
  app1:
    # ...
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

---

## Security Hardening

1. **Use non-root users in containers**
2. **Enable Docker secrets** for sensitive data
3. **Regular security updates**: `docker compose pull && docker compose up -d`
4. **Firewall configuration**: Only allow ports 80, 443
5. **Rate limiting**: Already configured in Nginx
6. **Regular backups**: Automated daily backups
7. **Log monitoring**: Set up alerts for suspicious activity
8. **SSL/TLS**: Use Let's Encrypt with auto-renewal

---

## Troubleshooting

### Container won't start

```bash
# Check logs
docker compose -f docker-compose.prod.yml logs <service>

# Check resource usage
docker stats

# Inspect container
docker inspect urlshortener-app1
```

### Database connection issues

```bash
# Test connection
docker exec -it urlshortener-postgres psql -U appuser -d urlshortener_prod

# Check database logs
docker compose -f docker-compose.prod.yml logs postgres
```

For more troubleshooting, see [TROUBLESHOOTING.md](../TROUBLESHOOTING.md).

---

**Last Updated:** 2025-12-19
**Maintained By:** DevOps Team
