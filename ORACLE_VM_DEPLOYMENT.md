# 🚀 Oracle VM Deployment Guide - URL Shortener System

Complete deployment guide for deploying the URL shortener system to Oracle VM, with options for shared or separate PostgreSQL with existing Odoo installation.

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Docker Installation](#docker-installation)
4. [PostgreSQL Strategy](#postgresql-strategy)
5. [Application Deployment](#application-deployment)
6. [SSL and Domain Configuration](#ssl-and-domain-configuration)
7. [Security Configuration](#security-configuration)
8. [Monitoring and Maintenance](#monitoring-and-maintenance)
9. [Backup Strategy](#backup-strategy)
10. [Troubleshooting](#troubleshooting)

---

## 📋 Prerequisites

- Oracle VM with Ubuntu/CentOS/RHEL (recommended: Ubuntu 20.04+)
- Root or sudo access
- Domain name pointed to your VM's IP address
- At least 4GB RAM, 2 CPU cores, 20GB storage
- Existing Odoo installation with PostgreSQL (optional)

---

## 🔧 Environment Setup

### 1. Update System

```bash
# Ubuntu/Debian
sudo apt update && sudo apt upgrade -y

# CentOS/RHEL
sudo yum update -y
# or for newer versions:
sudo dnf update -y
```

### 2. Install Essential Dependencies

```bash
# Ubuntu/Debian
sudo apt install -y curl wget git vim unzip htop net-tools ufw

# CentOS/RHEL
sudo yum install -y curl wget git vim unzip htop net-tools firewalld
# or:
sudo dnf install -y curl wget git vim unzip htop net-tools firewalld
```

### 3. Create Application User

```bash
sudo useradd -m -s /bin/bash urlshortener
sudo usermod -aG sudo urlshortener
sudo su - urlshortener
```

---

## 🐳 Docker Installation

### 1. Install Docker

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
sudo systemctl enable docker
sudo systemctl start docker

# CentOS/RHEL
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker $USER
```

### 2. Install Docker Compose

```bash
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### 3. Re-login to Apply Docker Group Changes

```bash
exit  # Exit from urlshortener user
sudo su - urlshortener  # Log back in
```

---

## 🗄️ PostgreSQL Strategy

### Option A: Shared PostgreSQL with Odoo ✅ **RECOMMENDED**

**Benefits:**
- Lower resource usage (saves 300-500MB RAM)
- Single PostgreSQL instance to maintain
- Unified backup strategy
- Cost-effective for Oracle VM

### Option B: Separate PostgreSQL

**Benefits:**
- Complete isolation between applications
- Independent scaling and configuration
- No risk of resource conflicts

## Shared PostgreSQL Setup (Recommended)

### 1. Assess Current Odoo Setup

```bash
# Find your Odoo containers
docker ps | grep -E "(odoo|postgres)"

# Check Odoo's docker-compose configuration
ls -la /opt/odoo/ || ls -la ~/odoo/

# Examine the docker-compose.yml
cat /path/to/your/odoo/docker-compose.yml
```

### 2. Create Shared Network

```bash
# Create a shared network for both applications
docker network create shared-backend
```

### 3. Connect Odoo to Shared Network

Add to your Odoo's `docker-compose.yml`:
```yaml
networks:
  default:
    name: shared-backend
    external: true
```

### 4. Database Setup in Shared PostgreSQL

```bash
# Connect to your Odoo PostgreSQL container
docker exec -it <odoo_postgres_container_name> psql -U odoo
```

In PostgreSQL shell:
```sql
-- Create dedicated user for URL shortener
CREATE USER urlshortener WITH PASSWORD 'your_secure_password_here';

-- Create database
CREATE DATABASE urlshortener_db OWNER urlshortener;

-- Grant necessary privileges
GRANT ALL PRIVILEGES ON DATABASE urlshortener_db TO urlshortener;

-- Exit PostgreSQL
\q
```

---

## 📁 Application Deployment

### 1. Clone Repository

```bash
cd /home/urlshortener
git clone https://github.com/yourusername/URLshorter.git
cd URLshorter
```

### 2. Create Production Environment File

Create `/home/urlshortener/URLshorter/.env.prod`:

```env
# Database Configuration (using shared PostgreSQL)
URLSHORTENER_DB_PASSWORD=your_secure_password_for_urlshortener_user

# Application Configuration
BASE_URL=https://yourdomain.com
JWT_SECRET=your_very_secure_jwt_secret_key_here_minimum_32_characters
JWT_ISSUER=yourdomain.com

# Redis Configuration
REDIS_PORT=6380

# OAuth Configuration (optional)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret

# Email Configuration (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_app_password

# SMS Configuration (optional)
TWILIO_ACCOUNT_SID=your_twilio_sid
TWILIO_AUTH_TOKEN=your_twilio_token
TWILIO_PHONE_NUMBER=your_twilio_number
```

### 3. Generate Strong Passwords and Secrets

```bash
# Generate strong database password
openssl rand -base64 32

# Generate JWT secret (at least 32 characters)
openssl rand -base64 48

# Make env file readable only by owner
chmod 600 /home/urlshortener/URLshorter/.env.prod
```

### 4. Create Shared PostgreSQL Docker Compose

Create `/home/urlshortener/URLshorter/docker/docker-compose.shared.yml`:

```yaml
version: '3.8'

services:
  app:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    environment:
      # Connect to existing Odoo PostgreSQL
      - DB_HOST=odoo_postgres  # Use your Odoo's PostgreSQL container name
      - DB_PORT=5432
      - DB_USER=urlshortener
      - DB_PASSWORD=${URLSHORTENER_DB_PASSWORD}
      - DB_NAME=urlshortener_db
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - SERVER_PORT=8080
      - BASE_URL=${BASE_URL}
      - ENVIRONMENT=production
      - JWT_SECRET=${JWT_SECRET}
      - JWT_ISSUER=${JWT_ISSUER}
    restart: unless-stopped
    ports:
      - "8080:8080"
    depends_on:
      - redis
    networks:
      - shared-backend
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    volumes:
      - redis_data:/data
    ports:
      - "127.0.0.1:6380:6379"  # Different port to avoid conflict
    networks:
      - shared-backend
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
        reservations:
          memory: 128M
          cpus: '0.1'
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.prod.conf:/etc/nginx/nginx.conf:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - app
    restart: unless-stopped
    networks:
      - shared-backend

volumes:
  redis_data:
    driver: local

networks:
  shared-backend:
    external: true
```

### 5. Initialize URL Shortener Schema

```bash
# Run your database initialization script
docker exec -i <odoo_postgres_container_name> psql -U urlshortener -d urlshortener_db < /home/urlshortener/URLshorter/configs/database.sql
```

---

## 🌐 SSL and Domain Configuration

### 1. Configure DNS

Point your domain's A record to your Oracle VM's public IP:
```
Type: A
Name: @ (or your subdomain)
Value: YOUR_VM_PUBLIC_IP
TTL: 300
```

### 2. Install Certbot for SSL

```bash
# Ubuntu/Debian
sudo apt install -y certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install -y epel-release
sudo yum install -y certbot python3-certbot-nginx
# or:
sudo dnf install -y epel-release
sudo dnf install -y certbot python3-certbot-nginx
```

### 3. Create Nginx Production Configuration

Create `/home/urlshortener/URLshorter/docker/nginx.prod.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream app {
        server app:8080;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
    limit_req_zone $binary_remote_addr zone=api:10m rate=100r/m;
    limit_req_zone $binary_remote_addr zone=general:10m rate=200r/m;

    server {
        listen 80;
        server_name yourdomain.com www.yourdomain.com;

        # Redirect HTTP to HTTPS
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name yourdomain.com www.yourdomain.com;

        # SSL Configuration
        ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
        ssl_session_timeout 1d;
        ssl_session_cache shared:SSL:50m;
        ssl_session_tickets off;

        # Modern SSL configuration
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # HSTS
        add_header Strict-Transport-Security "max-age=63072000" always;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Referrer-Policy "strict-origin-when-cross-origin";

        # API routes with rate limiting
        location /api/v1/auth/ {
            limit_req zone=login burst=10 nodelay;
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        location /api/ {
            limit_req zone=api burst=50 nodelay;
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Short URL redirects
        location ~ ^/[a-zA-Z0-9]+$ {
            limit_req zone=general burst=100 nodelay;
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # All other requests
        location / {
            limit_req zone=general burst=100 nodelay;
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Health check endpoint
        location /health {
            proxy_pass http://app;
            access_log off;
        }
    }
}
```

---

## 🔒 Security Configuration

### 1. Configure UFW Firewall (Ubuntu/Debian)

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw --force enable
sudo ufw status
```

### 2. Configure Firewalld (CentOS/RHEL)

```bash
sudo systemctl start firewalld
sudo systemctl enable firewalld
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
sudo firewall-cmd --list-all
```

### 3. Optimize PostgreSQL for Multiple Applications

```sql
-- Connect to PostgreSQL and optimize settings
-- Increase max connections if needed
ALTER SYSTEM SET max_connections = 200;

-- Configure memory settings
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET work_mem = '4MB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';

-- Restart PostgreSQL container to apply changes
```

---

## 🚀 Deployment Process

### 1. Deploy URL Shortener with Shared PostgreSQL

```bash
cd /home/urlshortener/URLshorter/docker

# Load environment variables
export $(cat ../.env.prod | xargs)

# Connect Odoo to shared network (if not already done)
cd /path/to/your/odoo
docker-compose down
# Update odoo docker-compose.yml to use shared-backend network
docker-compose up -d

# Deploy URL Shortener
cd /home/urlshortener/URLshorter/docker
docker-compose -f docker-compose.shared.yml up -d --build
```

### 2. Obtain SSL Certificate

```bash
# Stop nginx temporarily
docker-compose stop nginx

# Get SSL certificate
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# Start nginx again
docker-compose start nginx
```

### 3. Set up SSL Auto-renewal

```bash
# Test renewal
sudo certbot renew --dry-run

# Add cron job for auto-renewal
echo "0 12 * * * /usr/bin/certbot renew --quiet && cd /home/urlshortener/URLshorter/docker && /usr/local/bin/docker-compose restart nginx" | sudo crontab -
```

### 4. Verify Integration

```bash
# Check all containers are in the same network
docker network inspect shared-backend

# Test database connectivity
docker exec -it urlshortener_app_1 psql -h odoo_postgres -U urlshortener -d urlshortener_db

# Test application endpoints
curl https://yourdomain.com/health
curl -X POST https://yourdomain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

---

## 📊 Monitoring and Maintenance

### 1. Create System Service (Optional)

Create `/etc/systemd/system/urlshortener.service`:

```ini
[Unit]
Description=URL Shortener Service
Requires=docker.service
After=docker.service

[Service]
Type=forking
RemainAfterExit=yes
User=urlshortener
Group=urlshortener
WorkingDirectory=/home/urlshortener/URLshorter/docker
ExecStart=/usr/local/bin/docker-compose -f docker-compose.shared.yml up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

Enable the service:
```bash
sudo systemctl enable urlshortener
sudo systemctl start urlshortener
```

### 2. Create Log Rotation Configuration

Create `/etc/logrotate.d/urlshortener`:
```
/home/urlshortener/URLshorter/logs/*.log {
    daily
    missingok
    rotate 14
    compress
    notifempty
    create 644 urlshortener urlshortener
    postrotate
        cd /home/urlshortener/URLshorter/docker && docker-compose restart app
    endscript
}
```

---

## 🔄 Backup Strategy

### 1. Create Unified Backup Script

Create `/home/urlshortener/unified-backup.sh`:

```bash
#!/bin/bash

BACKUP_DIR="/home/backups"
DATE=$(date +%Y%m%d_%H%M%S)
POSTGRES_CONTAINER="<your_odoo_postgres_container_name>"

# Create backup directory
mkdir -p $BACKUP_DIR/{odoo,urlshortener}

# Backup Odoo database
docker exec $POSTGRES_CONTAINER pg_dump -U odoo -d odoo > $BACKUP_DIR/odoo/odoo_$DATE.sql

# Backup URL Shortener database
docker exec $POSTGRES_CONTAINER pg_dump -U urlshortener -d urlshortener_db > $BACKUP_DIR/urlshortener/urlshortener_$DATE.sql

# Backup Redis data
docker exec urlshortener_redis_1 redis-cli SAVE
docker cp urlshortener_redis_1:/data/dump.rdb $BACKUP_DIR/urlshortener/redis_$DATE.rdb

# Backup application files
tar -czf $BACKUP_DIR/urlshortener/app_$DATE.tar.gz -C /home/urlshortener URLshorter

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "Unified backup completed: $DATE"
```

### 2. Set Up Automated Backups

```bash
chmod +x /home/urlshortener/unified-backup.sh

# Add to cron (daily at 2 AM)
echo "0 2 * * * /home/urlshortener/unified-backup.sh >> /home/urlshortener/backup.log 2>&1" | crontab -
```

---

## 🔧 Maintenance Commands

### Daily Operations

```bash
# View service status
cd /home/urlshortener/URLshorter/docker && docker-compose ps

# View logs
docker-compose logs -f --tail=100 app
docker-compose logs -f --tail=100 nginx

# Update application
git pull origin main
docker-compose down
docker-compose -f docker-compose.shared.yml up -d --build

# Manual database backup
docker exec <odoo_postgres_container> pg_dump -U urlshortener -d urlshortener_db > backup_$(date +%Y%m%d).sql
```

### System Monitoring

```bash
# Check disk space
df -h

# Check memory usage
free -h

# Check running processes
htop

# Check Docker resources
docker stats

# Check network connectivity
docker network inspect shared-backend
```

---

## 🚨 Troubleshooting

### Common Issues

#### 1. Database Connection Issues
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test database connection
docker exec -it <postgres_container> psql -U urlshortener -d urlshortener_db

# Check database exists
docker exec -it <postgres_container> psql -U urlshortener -c "\l"
```

#### 2. Network Issues
```bash
# Check if containers are in the same network
docker network inspect shared-backend

# Restart networking
docker-compose down
docker-compose -f docker-compose.shared.yml up -d
```

#### 3. SSL Certificate Issues
```bash
# Check certificate validity
openssl x509 -in /etc/letsencrypt/live/yourdomain.com/fullchain.pem -text -noout

# Renew certificate manually
sudo certbot renew --force-renewal
```

#### 4. Application Errors
```bash
# Check application logs
docker-compose logs -f app

# Check database schema
docker exec -it <postgres_container> psql -U urlshortener -d urlshortener_db -c "\dt"

# Restart application
docker-compose restart app
```

### Performance Monitoring

```bash
# Monitor resource usage
docker stats --no-stream

# Check PostgreSQL performance
docker exec -it <postgres_container> psql -U odoo -c "SELECT * FROM pg_stat_activity;"

# Monitor Redis
docker exec -it urlshortener_redis_1 redis-cli info memory
```

---

## ✅ Production Checklist

- [ ] Domain DNS configured and propagated
- [ ] SSL certificate installed and auto-renewal configured
- [ ] Firewall configured (ports 80, 443, SSH only)
- [ ] Environment variables set securely (`.env.prod` with 600 permissions)
- [ ] Database initialized with proper schema
- [ ] Backups automated and tested
- [ ] Monitoring and logging configured
- [ ] Log rotation set up
- [ ] Rate limiting configured in Nginx
- [ ] Security headers implemented
- [ ] Resource limits configured for Docker containers
- [ ] System service created for auto-start
- [ ] Network integration with Odoo tested
- [ ] Health checks verified
- [ ] Performance baseline established

---

## 📈 Migration Path: Separate PostgreSQL (Future)

If you need to separate PostgreSQL later due to performance or compliance requirements:

### 1. Export Current Database
```bash
docker exec <odoo_postgres_container> pg_dump -U urlshortener -d urlshortener_db > urlshortener_migration.sql
```

### 2. Create Separate PostgreSQL Configuration
```yaml
# Add to docker-compose.separate.yml
postgres:
  image: postgres:15-alpine
  environment:
    POSTGRES_DB: urlshortener_db
    POSTGRES_USER: urlshortener
    POSTGRES_PASSWORD: ${DB_PASSWORD}
  volumes:
    - postgres_data:/var/lib/postgresql/data
  ports:
    - "127.0.0.1:5433:5432"  # Different port to avoid conflict
  networks:
    - shared-backend
```

### 3. Import Data to New Instance
```bash
docker exec -i new_postgres_container psql -U urlshortener -d urlshortener_db < urlshortener_migration.sql
```

### 4. Update Application Configuration
```yaml
# Update app service environment
- DB_HOST=postgres  # Change from odoo_postgres
- DB_PORT=5432
```

---

## 🎯 Final Notes

This deployment guide provides a production-ready setup for your URL shortener system on Oracle VM with the following benefits:

- **Resource Efficient**: Shared PostgreSQL saves 300-500MB RAM
- **Secure**: SSL encryption, rate limiting, security headers
- **Scalable**: Easy migration path to separate databases if needed
- **Maintainable**: Unified backup strategy and monitoring
- **Reliable**: Auto-restart, health checks, and proper logging

Remember to:
1. Regularly update your system packages
2. Monitor resource usage and performance
3. Test backups and disaster recovery procedures
4. Keep SSL certificates renewed
5. Monitor application logs for security issues

Your URL shortener is now production-ready! 🚀

---

## 📞 Support

For issues and questions:
- Check the troubleshooting section above
- Review Docker and application logs
- Monitor system resources
- Verify network connectivity between services

Happy deploying! 🎉