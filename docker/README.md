# Docker Setup for URL Shortener

This directory contains Docker configuration for local development and production deployment of the URL Shortener application.

## Quick Start

### Local Development (Hot Reload Enabled)

```bash
# Start all services with hot reload
docker-compose -f docker/docker-compose.local.yml up --build

# Or use the makefile command
make docker-dev
```

### Access Points

- **Frontend (Main)**: http://localhost:3000 (Next.js with HMR)
- **Backend API**: http://localhost:8080 (Go with air hot reload)
- **Admin Panel**: http://localhost:3000/admin
- **Adminer (DB UI)**: http://localhost:8081 (Database management)
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## Default Development Credentials

### ⚠️ SECURITY WARNING

**These credentials are for LOCAL DEVELOPMENT ONLY!**
**DO NOT use in production environments!**

### Application Users

| Role | Email | Password | Access Level |
|------|-------|----------|--------------|
| Admin User | `admin@example.com` | `admin123` | Full admin panel access (super_admin) |
| Test User | `test@example.com` | `admin123` | Regular user access (free tier) |

### Database Access

| Service | User | Password | Database |
|---------|------|----------|----------|
| PostgreSQL | `urlshortener` | `devpassword` | `urlshortener_db` |
| App Connection | `urlshort_app` | `app_dev_password` | `urlshortener_db` |
| Admin Connection | `urlshort_admin` | `admin_dev_password` | `urlshortener_db` |
| Readonly Connection | `urlshort_readonly` | `readonly_dev_password` | `urlshortener_db` |

### Adminer Access

- **URL**: http://localhost:8081
- **System**: PostgreSQL
- **Server**: postgres (Docker network name)
- **Username**: `urlshortener`
- **Password**: `devpassword`
- **Database**: `urlshortener_db`

## Database Initialization

When Docker containers start for the first time, PostgreSQL runs initialization scripts in alphabetical order:

1. `01-init.sql` - Creates basic schema (url_mappings, click_events)
2. `02-auth-schema.sql` - Creates authentication tables (users, sessions, etc.)
3. `03-setup-users.sql` - Creates PostgreSQL database users and permissions
4. `04-create-admin-user.sql` - **Creates default admin and test users** ⭐

### Resetting the Database

To reset the database and recreate all data:

```bash
# Stop and remove all containers and volumes
docker-compose -f docker/docker-compose.local.yml down -v

# Start fresh
docker-compose -f docker/docker-compose.local.yml up --build
```

**Note**: This will delete ALL data including the default users. They will be recreated automatically.

## Creating Additional Users

### Using the Admin Panel

1. Login as admin: http://localhost:3000/admin/login
2. Navigate to Users section
3. Create new users with appropriate permissions

### Using the CLI Script

```bash
# Windows
.\scripts\create-admin.ps1 -Name "New Admin" -Email "admin2@example.com" -Password "secure123" -Type "super_admin"

# Linux/macOS
./scripts/create-admin.sh -n "New Admin" -e "admin2@example.com" -p "secure123" -t "super_admin"

# Or using Go directly
go run scripts/create-admin.go -name="New Admin" -email="admin2@example.com" -password="secure123" -type="super_admin"
```

### Account Types

- `super_admin` - Full system access
- `system_admin` - System management
- `support_admin` - User support
- `billing_admin` - Billing management
- `content_admin` - Content management
- `free` - Regular user (free tier)
- `premium` - Premium user
- `enterprise` - Enterprise user

## Testing Login

### Normal User Login

1. Navigate to: http://localhost:3000/login
2. Email: `test@example.com`
3. Password: `admin123`
4. Should redirect to homepage

### Admin Login

1. Navigate to: http://localhost:3000/admin/login
2. Email: `admin@example.com`
3. Password: `admin123`
4. Should redirect to admin dashboard

## Environment Variables

Default environment variables are set in `docker-compose.local.yml`. You can override them by creating a `.env` file in the `docker` directory.

Example `.env`:
```env
DB_PASSWORD=your_custom_password
JWT_SECRET=your_jwt_secret_key
REDIS_PASSWORD=your_redis_password
```

## Docker Compose Files

- `docker-compose.local.yml` - Local development with hot reload
- `docker-compose.yml` - Production deployment (if exists)

## Troubleshooting

### Login Fails with "Invalid Credentials"

1. **Check if users exist in database:**
```bash
docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db -c "SELECT email, name, account_type, is_admin FROM users;"
```

Expected output should show admin@example.com and test@example.com

2. **If no users found, check initialization logs:**
```bash
docker-compose -f docker/docker-compose.local.yml logs postgres | grep "DOCKER DEVELOPMENT USERS CREATED"
```

3. **Reset the database:**
```bash
docker-compose -f docker/docker-compose.local.yml down -v
docker-compose -f docker/docker-compose.local.yml up --build
```

### Database Connection Issues

Check if PostgreSQL is healthy:
```bash
docker-compose -f docker/docker-compose.local.yml ps postgres
```

View database logs:
```bash
docker-compose -f docker/docker-compose.local.yml logs postgres
```

### Backend Not Starting

View backend logs:
```bash
docker-compose -f docker/docker-compose.local.yml logs backend
```

Check if migrations ran successfully:
```bash
docker-compose -f docker/docker-compose.local.yml logs backend | grep migration
```

## Production Deployment

### ⚠️ Before Deploying to Production

1. **Change ALL default passwords**
2. **Remove or disable the default user creation script**
3. **Use environment-specific secrets**
4. **Enable SSL/TLS for PostgreSQL**
5. **Configure proper authentication methods**
6. **Set up proper backup procedures**
7. **Review all environment variables**

### Disable Default User Creation

In production, comment out or remove this line from `docker-compose.yml`:
```yaml
# - ./init-scripts/04-create-admin-user.sql:/docker-entrypoint-initdb.d/04-create-admin-user.sql
```

Or set an environment variable to skip user creation:
```yaml
environment:
  CREATE_DEFAULT_ADMIN: "false"
```

## Additional Resources

- [PGBouncer Configuration](README_PGBOUNCER.md)
- [Build Troubleshooting](TROUBLESHOOTING_BUILD.md)
- [Main Project README](../README.md)
- [API Documentation](../docs/API.md) (if exists)

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review logs: `docker-compose logs [service_name]`
3. Check project issues on GitHub
4. Consult the main project documentation

---

**Last Updated**: January 2026
**Docker Compose Version**: 3.8
**PostgreSQL Version**: 15-alpine
**Redis Version**: 7-alpine
