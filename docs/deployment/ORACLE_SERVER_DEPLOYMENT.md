# Oracle Server Deployment Guide
## Centralized PostgreSQL Architecture

This guide shows how to deploy URLShortener on your Oracle server with a centralized PostgreSQL database shared across multiple applications (Odoo, URLShortener, future systems).

## 🎯 Architecture Overview

```
Oracle Cloud Server (Ubuntu)
│
├── Centralized PostgreSQL (Port 5432)
│   ├── Database: odoo_db (for Odoo)
│   ├── Database: urlshortener_db (for URLShortener)
│   └── Database: [future applications]
│
├── Odoo Stack (Port 8069)
├── URLShortener Stack (Port 8090)
├── Mailcow Stack (Ports 25, 110, 143, 465, 587, 993, 995)
└── Adminer - Database UI (Port 8081)
```

## 📋 Step-by-Step Deployment

### Step 1: Create Centralized PostgreSQL Container

```bash
# Create a dedicated Docker network for database access
docker network create --driver bridge shared-db-network

# Create a directory for PostgreSQL data persistence
mkdir -p /opt/postgresql/data
sudo chown -R 999:999 /opt/postgresql/data  # PostgreSQL user ID

# Start centralized PostgreSQL container
docker run -d \
  --name postgresql-central \
  --network shared-db-network \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=CvUpK3NYZy2h5K9#oyVk3754mbQa45ra \
  -e POSTGRES_USER=postgres \
  -v /opt/postgresql/data:/var/lib/postgresql/data \
  --restart unless-stopped \
  postgres:15-alpine

# Verify it's running
docker ps | grep postgresql-central
```

### Step 2: Create Databases for Each Application

```bash
# Access PostgreSQL
docker exec -it postgresql-central psql -U postgres

# Create databases
CREATE DATABASE odoo_db;
CREATE DATABASE urlshortener_db;

# Create dedicated users with strong passwords
CREATE USER odoo_user WITH ENCRYPTED PASSWORD 'OdooSecurePassword123!';
CREATE USER urlshortener_user WITH ENCRYPTED PASSWORD 'URLShortenerSecurePassword123!';

# Grant privileges
GRANT ALL PRIVILEGES ON DATABASE odoo_db TO odoo_user;
GRANT ALL PRIVILEGES ON DATABASE urlshortener_db TO urlshortener_user;

# For PostgreSQL 15+, grant schema privileges
\c odoo_db
GRANT ALL ON SCHEMA public TO odoo_user;

\c urlshortener_db
GRANT ALL ON SCHEMA public TO urlshortener_user;

# Exit
\q
```

### Step 3: Backup and Migrate Existing Odoo Database

```bash
# Backup current Odoo database
docker exec odoo-deploy_db_1 pg_dump -U odoo odoo > /tmp/odoo_backup_$(date +%Y%m%d).sql

# Restore to centralized PostgreSQL
docker exec -i postgresql-central psql -U odoo_user -d odoo_db < /tmp/odoo_backup_$(date +%Y%m%d).sql

# Verify backup
docker exec -it postgresql-central psql -U odoo_user -d odoo_db -c "\dt"
```

### Step 4: Update Odoo to Use Centralized Database

```bash
cd /path/to/your/odoo-deploy

# Edit docker-compose.yml
nano docker-compose.yml
```

**Modify Odoo's docker-compose.yml:**

```yaml
version: '3'
services:
  odoo:
    image: odoo:18.0
    container_name: odoo-central
    depends_on:
      - db
    ports:
      - "8069:8069"
    environment:
      - HOST=postgresql-central  # Changed from 'db'
      - USER=odoo_user           # Changed from default
      - PASSWORD=OdooSecurePassword123!
    volumes:
      - odoo-data:/var/lib/odoo
    networks:
      - shared-db-network        # Use shared network

  # Remove the old 'db' service completely
  # db:
  #   image: postgres:13
  #   ...

networks:
  shared-db-network:
    external: true               # Use external network

volumes:
  odoo-data:
```

**Redeploy Odoo:**

```bash
# Stop old Odoo containers
docker-compose down

# Remove old database container (already backed up)
docker rm odoo-deploy_db_1

# Start Odoo with centralized database
docker-compose up -d

# Check logs
docker logs -f odoo-central
```

### Step 5: Deploy URLShortener with Centralized Database

```bash
# Clone repository
cd /opt
git clone https://github.com/malinanu/urlshotern.git urlshortener
cd urlshortener

# Create custom production docker-compose for Oracle server
nano docker-compose.oracle.yml
```

**Create `docker-compose.oracle.yml`:**

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    container_name: urlshortener-app
    ports:
      - "8090:8080"
    environment:
      - SERVER_PORT=8080
      - DB_HOST=postgresql-central     # Centralized PostgreSQL
      - DB_PORT=5432
      - DB_USER=urlshortener_user
      - DB_PASSWORD=URLShortenerSecurePassword123!
      - DB_NAME=urlshortener_db
      - DB_MAX_CONNECTIONS=25
      - DB_MAX_IDLE_CONNECTIONS=10
      - DB_CONN_MAX_LIFETIME=3600
      - DB_CONN_MAX_IDLE_TIME=600
      - DB_SSLMODE=disable
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - NODE_ID=1
      - BASE_URL=https://your-domain.com  # Update with your domain
      - ENVIRONMENT=production
      - POLAR_ACCESS_TOKEN=${POLAR_ACCESS_TOKEN:-}
      - POLAR_WEBHOOK_SECRET=${POLAR_WEBHOOK_SECRET:-}
    depends_on:
      - redis
    networks:
      - shared-db-network          # Connect to shared database network
      - urlshortener-network       # Internal network for Redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  redis:
    image: redis:7-alpine
    container_name: urlshortener-redis
    ports:
      - "127.0.0.1:6380:6379"      # Changed port to avoid conflicts
    volumes:
      - redis-data:/data
    networks:
      - urlshortener-network
    restart: unless-stopped
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

  adminer:
    image: adminer:latest
    container_name: urlshortener-adminer
    ports:
      - "8081:8080"
    environment:
      - ADMINER_DEFAULT_SERVER=postgresql-central
    networks:
      - shared-db-network
    restart: unless-stopped

networks:
  shared-db-network:
    external: true                 # Use external centralized network
  urlshortener-network:
    driver: bridge

volumes:
  redis-data:
```

**Deploy URLShortener:**

#### Option A: Deploy via Docker CLI (Direct)

```bash
# Build and start
docker-compose -f docker-compose.oracle.yml up -d --build

# Check logs
docker logs -f urlshortener-app

# Verify health
curl http://localhost:8080/health
```

#### Option B: Deploy via Portainer (Recommended for Easier Management)

**Prerequisites:**
- Portainer installed and accessible (e.g., `https://your-oracle-server:9443`)
- Access to your Git repository (public or with credentials)

**Step-by-Step Portainer Deployment:**

1. **Access Portainer:**
   - Open your browser: `https://your-oracle-server:9443`
   - Log in with your Portainer credentials

2. **Create or Update Stack:**
   - Navigate to **Stacks** in the left sidebar
   - If `urlshortener` stack exists: Click on it → **Editor** button
   - If new: Click **+ Add stack** button

3. **Configure Stack:**
   - **Name:** `urlshortener`
   - **Build method:** Choose "Repository" (recommended)
   - **Repository URL:** `https://github.com/malinanu/urlshotern`
   - **Branch:** `main`
   - **Compose path:** `docker-compose.oracle.yml`
   - **Authentication:** Add if private repository

4. **Set Environment Variables:**
   Scroll to **Environment variables** section and add:
   ```
   BASE_URL=https://url.yourdomain.com
   DB_PASSWORD=URLShortenerSecurePassword123!
   POLAR_ACCESS_TOKEN=your_token_here (optional)
   POLAR_WEBHOOK_SECRET=your_secret_here (optional)
   ```

5. **Deploy:**
   - Check **Pull and redeploy** checkbox (rebuilds with latest code)
   - Click **Deploy the stack** (new) or **Update the stack** (existing)
   - Wait 8-12 minutes for first build (installs Node.js + Go dependencies, builds frontend)

6. **Monitor Deployment:**
   - Go to **Containers** → `urlshortener-app`
   - Click **Logs** tab
   - Toggle **Auto-refresh logs**
   - Watch for successful build messages:
     - `📦 Building frontend...`
     - `✅ Frontend build complete!`
     - `Server starting on port 8080`

7. **Verify Static Files:**
   - Click **Console** button on container
   - Select `/bin/sh`, click **Connect**
   - Run: `ls -la /app/static/`
   - Should see: `index.html`, `_next/` directory, etc.

**Portainer Management Tips:**

- **View Logs in Real-Time:**
  - Containers → `urlshortener-app` → **Logs** → Toggle **Auto-refresh**

- **Restart Container:**
  - Containers → `urlshortener-app` → **Restart** button
  - Useful for config changes without rebuilding

- **Rebuild After Code Changes:**
  - Stacks → your stack → **Editor**
  - Check **Pull and redeploy**
  - Click **Update the stack**

- **Check Resource Usage:**
  - Containers → `urlshortener-app` → **Stats** tab
  - View CPU, memory, network usage

- **Delete Old Images (if rebuild fails):**
  - Images → Find `docker-backend` → Delete
  - Return to stack and update again

### Step 6: Configure Domain Mapping in Oracle Cloud Panel

**In your Oracle Cloud Panel:**

1. **Add A Records:**
   - `url.yourdomain.com` → `Your Oracle Server Public IP`
   - `odoo.yourdomain.com` → `Your Oracle Server Public IP`
   - `db.yourdomain.com` → `Your Oracle Server Public IP` (for Adminer)

2. **Configure Firewall Rules:**
   ```bash
   # On Oracle server
   sudo ufw allow 8090/tcp  # URLShortener
   sudo ufw allow 8069/tcp  # Odoo
   sudo ufw allow 8081/tcp  # Adminer (optional, can restrict to specific IPs)
   sudo ufw reload
   ```

3. **Setup Reverse Proxy with Nginx (Recommended):**

```bash
# Install Nginx
sudo apt update
sudo apt install nginx certbot python3-certbot-nginx -y

# Create Nginx config for URLShortener
sudo nano /etc/nginx/sites-available/urlshortener
```

**Nginx configuration (`/etc/nginx/sites-available/urlshortener`):**

> **Note:** The URLShortener uses a unified container architecture where the Go backend serves both the API endpoints and the frontend static files. Nginx proxies all requests to the backend container on port 8090.

```nginx
server {
    listen 80;
    server_name url.yourdomain.com;

    # Proxy all requests to backend (serves both API and frontend static files)
    location / {
        proxy_pass http://localhost:8090;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Increase timeouts for long-running requests
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

server {
    listen 80;
    server_name odoo.yourdomain.com;

    location / {
        proxy_pass http://localhost:8069;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}

server {
    listen 80;
    server_name db.yourdomain.com;

    # Restrict Adminer access to specific IPs (recommended)
    # allow YOUR_IP;
    # deny all;

    location / {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
    }
}
```

**Enable sites and get SSL certificates:**

```bash
# Enable sites
sudo ln -s /etc/nginx/sites-available/urlshortener /etc/nginx/sites-enabled/

# Test Nginx config
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx

# Get SSL certificates (Let's Encrypt)
sudo certbot --nginx -d url.yourdomain.com -d odoo.yourdomain.com -d db.yourdomain.com

# Auto-renewal is enabled by default
sudo systemctl status certbot.timer
```

### Step 7: Verify Deployment

#### Container Health Checks

```bash
# Check all containers
docker ps

# Expected containers:
# - postgresql-central (Port 5432)
# - odoo-central (Port 8069)
# - urlshortener-app (Port 8090)
# - urlshortener-redis (Port 6380)
# - urlshortener-adminer (Port 8081)
# - mailcow containers (existing)

# Check URLShortener container status
docker ps --filter "name=urlshortener-app" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# View container logs
docker logs --tail=50 urlshortener-app

# Test URLShortener health
curl http://localhost:8090/health

# Test Odoo
curl http://localhost:8069/web/database/selector

# Access Adminer
# http://your-server-ip:8081
# Server: postgresql-central
# Username: postgres (or urlshortener_user/odoo_user)
# Password: [as configured]
```

#### Frontend Verification Checklist

After deployment, verify the unified container is serving both frontend and backend:

**✓ Static Files Exist:**
```bash
docker exec urlshortener-app ls -la /app/static/
# Expected output:
# - index.html
# - _next/ directory
# - Other static assets (CSS, JS files)
```

**✓ Frontend Homepage Loads:**
```bash
# Test from server
curl http://localhost:8090/ | head -20
# Should return HTML content starting with <!DOCTYPE html>

# Test via browser
# Visit: https://url.yourdomain.com
# Should show: URLShortener homepage with login/signup interface
```

**✓ Frontend Routes Work:**
- `https://url.yourdomain.com/` - Homepage ✓
- `https://url.yourdomain.com/dashboard` - Dashboard page ✓
- `https://url.yourdomain.com/login` - Login page ✓
- `https://url.yourdomain.com/signup` - Signup page ✓

**✓ API Endpoints Work:**
```bash
# Health check
curl https://url.yourdomain.com/api/v1/health
# Expected: {"status":"ok","database":"connected","redis":"connected"}

# Database pool stats
curl https://url.yourdomain.com/health/db-pool
```

**✓ Static Assets Load:**
- Browser console: No 404 errors for `/_next/static/*` files
- CSS styles applied correctly
- JavaScript interactive elements work

**✓ Short URL Redirects Work:**
```bash
# Create a short URL
curl -X POST https://url.yourdomain.com/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'

# Test redirect (use short code from response)
curl -I https://url.yourdomain.com/{shortCode}
# Expected: HTTP 302 redirect to original URL
```

**✓ Nginx Proxying Correctly:**
```bash
# Check Nginx access logs
sudo tail -f /var/log/nginx/access.log

# Visit frontend in browser, should see requests proxied to backend
# Look for: GET / HTTP/1.1 200 (HTML page)
#           GET /_next/static/... HTTP/1.1 200 (static assets)
```

**Architecture Confirmation:**
- ✓ **Single Container:** Only `urlshortener-app` running (no separate frontend container)
- ✓ **Port 8090:** Container exposes port 8080, mapped to host 8090
- ✓ **Nginx Proxy:** All routes proxied to `localhost:8090`
- ✓ **Backend Serves Both:** Go backend serves API at `/api/v1/*` and static files at `/`

### Step 8: Monitor and Maintain

```bash
# View logs
docker logs -f urlshortener-app
docker logs -f postgresql-central
docker logs -f odoo-central

# Database backups (automated script)
sudo nano /opt/scripts/backup-databases.sh
```

**Backup script (`/opt/scripts/backup-databases.sh`):**

```bash
#!/bin/bash

BACKUP_DIR="/opt/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Backup all databases
docker exec postgresql-central pg_dumpall -U postgres > "$BACKUP_DIR/all_databases_$DATE.sql"

# Backup individual databases
docker exec postgresql-central pg_dump -U urlshortener_user urlshortener_db > "$BACKUP_DIR/urlshortener_$DATE.sql"
docker exec postgresql-central pg_dump -U odoo_user odoo_db > "$BACKUP_DIR/odoo_$DATE.sql"

# Keep only last 7 days of backups
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete

echo "Backup completed: $DATE"
```

**Setup cron job for daily backups:**

```bash
sudo chmod +x /opt/scripts/backup-databases.sh

# Add to crontab
sudo crontab -e

# Add this line (runs daily at 2 AM)
0 2 * * * /opt/scripts/backup-databases.sh >> /var/log/db-backup.log 2>&1
```

## 🔒 Security Considerations

1. **PostgreSQL Security:**
   - Use strong passwords for all database users
   - Restrict network access with `pg_hba.conf` if needed
   - Keep PostgreSQL image updated

2. **Firewall:**
   - Only expose necessary ports (8069, 8090, 443, 80)
   - Restrict Adminer access to trusted IPs
   - Use UFW or Oracle Cloud security lists

3. **SSL/TLS:**
   - Always use HTTPS in production (Let's Encrypt)
   - Update BASE_URL in URLShortener to use `https://`

4. **Environment Variables:**
   - Never commit passwords to Git
   - Use `.env` files (added to `.gitignore`)

5. **Regular Updates:**
   ```bash
   # Update Docker images
   docker pull postgres:15-alpine
   docker pull redis:7-alpine
   docker pull odoo:18.0

   # Rebuild URLShortener
   cd /opt/urlshortener
   git pull origin main
   docker-compose -f docker-compose.oracle.yml up -d --build
   ```

## 📊 Monitoring

**Access Points:**
- URLShortener: `https://url.yourdomain.com`
- Odoo: `https://odoo.yourdomain.com`
- Adminer: `https://db.yourdomain.com`
- Mailcow: (existing setup)

**Health Checks:**
```bash
# URLShortener
curl https://url.yourdomain.com/health

# Database connection pool
curl https://url.yourdomain.com/health/db-pool

# Goroutines monitoring
curl https://url.yourdomain.com/health/goroutines
```

## 🔄 Future Application Integration

When adding new applications, follow this pattern:

```yaml
services:
  new-app:
    # ... app configuration ...
    environment:
      - DB_HOST=postgresql-central
      - DB_PORT=5432
      - DB_USER=newapp_user
      - DB_PASSWORD=SecurePassword
      - DB_NAME=newapp_db
    networks:
      - shared-db-network
```

Then create the database:

```bash
docker exec -it postgresql-central psql -U postgres
CREATE DATABASE newapp_db;
CREATE USER newapp_user WITH ENCRYPTED PASSWORD 'SecurePassword';
GRANT ALL PRIVILEGES ON DATABASE newapp_db TO newapp_user;
```

## 🆘 Troubleshooting

**Issue: Cannot connect to PostgreSQL**
```bash
# Check if PostgreSQL is running
docker ps | grep postgresql-central

# Check network connectivity
docker network inspect shared-db-network

# Test connection from app container
docker exec urlshortener-app ping postgresql-central
```

**Issue: Odoo won't start after migration**
```bash
# Check Odoo logs
docker logs odoo-central

# Verify database connection
docker exec odoo-central odoo --version

# Reset Odoo config if needed
docker exec -it odoo-central /bin/bash
cat /etc/odoo/odoo.conf
```

**Issue: URLShortener shows database errors**
```bash
# Check migration status
docker exec -it postgresql-central psql -U urlshortener_user -d urlshortener_db
\dt  # List tables

# Run migrations manually if needed
docker exec urlshortener-app /app/scripts/migrate.sh status
```

**Issue: Frontend not loading (blank page or 404 errors)**

This issue occurs if the unified container was built with `Dockerfile.backend.dev` instead of `Dockerfile.dev`.

**Symptoms:**
- Visiting domain shows blank page or Nginx error
- API endpoints work (`/api/v1/health` returns JSON)
- Browser console shows 404 errors for static assets

**Solution:**
1. Update `docker-compose.oracle.yml` to use correct Dockerfile:
   ```yaml
   dockerfile: Dockerfile.dev  # Not Dockerfile.backend.dev
   ```

2. Add frontend environment variables:
   ```yaml
   environment:
     - NODE_ENV=production
     - NEXT_TELEMETRY_DISABLED=1
     - NEXT_PUBLIC_API_URL=${BASE_URL}
   ```

3. Rebuild container:
   ```bash
   # Via Docker CLI
   docker-compose -f docker-compose.oracle.yml up -d --build

   # Via Portainer
   # Stacks → your stack → Editor → Check "Pull and redeploy" → Update
   ```

4. Verify static files exist:
   ```bash
   docker exec urlshortener-app ls -la /app/static/
   # Should show: index.html, _next/ directory, etc.
   ```

5. Check container logs for build errors:
   ```bash
   docker logs urlshortener-app | grep -i "frontend"
   # Look for: "✅ Frontend build complete!"
   ```

**Issue: Portainer build fails with "frontend build not found"**
```bash
# Check disk space (needs ~2GB for build)
df -h

# Check container logs for npm errors
docker logs urlshortener-app | grep -i "error"

# Delete old image and rebuild
docker rmi $(docker images -q docker-backend)
# Then rebuild via Portainer
```

**Issue: Long build time (10+ minutes)**
- This is **expected behavior** for first build
- Installing Node.js + Go dependencies takes 8-12 minutes
- Subsequent builds use cached layers and are faster
- Monitor logs: Stacks → Container → Logs → Auto-refresh

**Issue: Nginx shows 502 Bad Gateway**
```bash
# Check if container is running
docker ps | grep urlshortener-app

# Check if app is listening on port 8080
docker exec urlshortener-app netstat -tlnp | grep 8080

# Test local connection
curl http://localhost:8090/

# Check Nginx configuration
sudo nginx -t
sudo systemctl status nginx

# Verify Nginx is proxying to correct port (8090)
sudo nano /etc/nginx/sites-available/urlshortener
```

## 📈 Performance Tuning

**PostgreSQL optimization for multiple applications:**

```bash
docker exec -it postgresql-central psql -U postgres
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
ALTER SYSTEM SET work_mem = '4MB';

# Restart PostgreSQL
docker restart postgresql-central
```

## ✅ Post-Deployment Checklist

- [ ] PostgreSQL centralized and running
- [ ] Odoo migrated to centralized database
- [ ] URLShortener deployed and accessible
- [ ] Adminer accessible for database management
- [ ] SSL certificates installed (Let's Encrypt)
- [ ] Firewall rules configured
- [ ] Backup script created and tested
- [ ] Cron job for daily backups enabled
- [ ] Health endpoints responding
- [ ] Nginx reverse proxy configured
- [ ] Domain names mapped correctly
- [ ] All services restarting properly (`restart: unless-stopped`)

---

**Created:** 2026-02-07
**Last Updated:** 2026-02-07
**Server:** Oracle Cloud Ubuntu Instance
