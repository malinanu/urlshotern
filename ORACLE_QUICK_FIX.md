# Oracle Server Deployment - Quick Fix Guide

## 🚨 "ContainerConfig" Error Fix

If you get this error when deploying:
```
KeyError: 'ContainerConfig'
```

This is a known bug in docker-compose 1.29.2 with corrupted container metadata.

---

## ⚡ Quick Fix (5 Minutes)

On your Oracle server, run these commands:

```bash
cd /opt/urlshotern  # or wherever your project is

# 1. Stop and remove corrupted container
docker stop urlshortener-app
docker rm -f urlshortener-app

# 2. Remove corrupted image
docker rmi -f $(docker images | grep urlshortener-app | awk '{print $3}')

# 3. Clean up
docker system prune -f

# 4. Rebuild app image
docker build -f Dockerfile.backend.dev -t urlshortener-app:latest .

# 5. Start app manually
docker run -d \
  --name urlshortener-app \
  --network shared-db-network \
  -p 8090:8080 \
  -e SERVER_PORT=8080 \
  -e DB_HOST=postgresql-central \
  -e DB_PORT=5432 \
  -e DB_USER=urlshortener_user \
  -e DB_PASSWORD=URLShortenerSecurePassword123! \
  -e DB_NAME=urlshortener_db \
  -e REDIS_HOST=urlshortener-redis \
  -e REDIS_PORT=6379 \
  -e BASE_URL=https://url.yourdomain.com \
  -e ENVIRONMENT=production \
  -e NODE_ID=1 \
  --restart unless-stopped \
  urlshortener-app:latest

# 6. Connect to Redis network
docker network connect urlshortener-network urlshortener-app 2>/dev/null || true

# 7. Verify
docker ps | grep urlshortener
curl http://localhost:8090/health
```

---

## 🛠️ Using Automation Scripts (Recommended)

After pulling the latest code, you'll have automation scripts:

### 1. Pull Latest Code
```bash
cd /opt/urlshotern
git pull origin main
```

### 2. Make Scripts Executable
```bash
chmod +x scripts/*.sh
```

### 3. Deploy
```bash
# Option A: Full deployment script (checks prerequisites)
./scripts/deploy-oracle.sh

# Option B: Cleanup + manual docker-compose
./scripts/cleanup-containers.sh
docker-compose -f docker-compose.oracle.yml up -d --build
```

### 4. Health Check
```bash
./scripts/health-check.sh
```

---

## 📋 Files Created

New files for Oracle deployment:
- **docker-compose.oracle.yml** - Oracle server compose configuration
- **scripts/cleanup-containers.sh** - Cleanup automation (Linux)
- **scripts/cleanup-containers.ps1** - Cleanup automation (Windows)
- **scripts/deploy-oracle.sh** - Deployment automation
- **scripts/health-check.sh** - Health verification
- **.env.oracle.example** - Environment template

---

## ✅ Verification

After deployment, check:

```bash
# All containers running
docker ps | grep urlshortener

# App health
curl http://localhost:8090/health

# Database pool
curl http://localhost:8090/health/db-pool

# Redis
docker exec urlshortener-redis redis-cli ping

# Logs (should show no errors)
docker logs urlshortener-app --tail 50
```

---

## 🔄 Future Deployments

Always use the deployment script to avoid this issue:

```bash
cd /opt/urlshotern
git pull origin main
./scripts/deploy-oracle.sh
```

---

## 🆘 Troubleshooting

### PostgreSQL Not Found
```bash
# Check if running
docker ps | grep postgresql-central

# If not, start it (see ORACLE_SERVER_DEPLOYMENT.md Step 1)
docker run -d \
  --name postgresql-central \
  --network shared-db-network \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=YourPassword \
  -v /opt/postgresql/data:/var/lib/postgresql/data \
  --restart unless-stopped \
  postgres:15-alpine
```

### Network Not Found
```bash
docker network create --driver bridge shared-db-network
docker network create --driver bridge urlshortener-network
```

### Port Already in Use
```bash
# Find what's using port 8090
sudo netstat -tulpn | grep 8090

# Kill the process or change port in docker-compose.oracle.yml
```

### Still Getting ContainerConfig Error

**Last resort - nuclear option:**
```bash
# Stop ALL URLShortener containers
docker stop $(docker ps -q --filter name=urlshortener)

# Remove ALL URLShortener containers
docker rm -f $(docker ps -aq --filter name=urlshortener)

# Remove ALL URLShortener images
docker rmi -f $(docker images -q | grep urlshortener)

# Prune everything
docker system prune -af

# Then rebuild from scratch
./scripts/deploy-oracle.sh
```

---

## 🔐 Security Notes

- PostgreSQL data is safe (external container)
- Redis cache can be rebuilt (no data loss)
- Always use strong passwords in .env
- Never commit .env to Git

---

**Created:** 2026-02-07
**Last Updated:** 2026-02-07
**Related:** ORACLE_SERVER_DEPLOYMENT.md (full deployment guide)
