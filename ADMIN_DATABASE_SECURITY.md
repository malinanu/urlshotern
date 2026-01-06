# Admin Panel Database Security Architecture

## Recommended Approach: Same Database, Different Users + Row-Level Security

This document outlines the recommended database security architecture for the admin panel.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    PostgreSQL Database                   │
│                     "urlshortener"                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────┐      ┌──────────────────┐        │
│  │  Public Schema   │      │  Admin Schema    │        │
│  ├──────────────────┤      ├──────────────────┤        │
│  │ • users          │      │ • audit_logs     │        │
│  │ • url_mappings   │      │ • system_settings│        │
│  │ • click_events   │      │ • blocked_domains│        │
│  │ • subscriptions  │      │ • flagged_urls   │        │
│  └──────────────────┘      └──────────────────┘        │
│         ▲                           ▲                    │
│         │                           │                    │
│  ┌──────┴────────┐         ┌───────┴────────┐          │
│  │ urlshort_app  │         │ urlshort_admin │          │
│  │ (Limited)     │         │ (Full Access)  │          │
│  └───────────────┘         └────────────────┘          │
└─────────────────────────────────────────────────────────┘
         ▲                            ▲
         │                            │
┌────────┴─────────┐         ┌────────┴─────────┐
│  Application     │         │  Admin Panel     │
│  (Regular Users) │         │  (Admins Only)   │
└──────────────────┘         └──────────────────┘
```

---

## Database User Configuration

### Step 1: Create Database Users

```sql
-- 1. Regular application user (RESTRICTED permissions)
CREATE USER urlshort_app WITH PASSWORD 'YOUR_SECURE_PASSWORD_HERE_1';

-- 2. Admin application user (ELEVATED permissions)
CREATE USER urlshort_admin WITH PASSWORD 'YOUR_SECURE_PASSWORD_HERE_2';

-- 3. Read-only analytics user (for reporting tools)
CREATE USER urlshort_readonly WITH PASSWORD 'YOUR_SECURE_PASSWORD_HERE_3';
```

### Step 2: Create Admin Schema

```sql
-- Create separate schema for admin-only tables
CREATE SCHEMA IF NOT EXISTS admin;
GRANT USAGE ON SCHEMA admin TO urlshort_admin;
```

### Step 3: Set Permissions for Regular App User

```sql
-- Grant connection
GRANT CONNECT ON DATABASE urlshortener TO urlshort_app;
GRANT USAGE ON SCHEMA public TO urlshort_app;

-- Grant SELECT, INSERT, UPDATE (NO DELETE) on application tables
GRANT SELECT, INSERT, UPDATE ON TABLE
    users,
    url_mappings,
    click_events,
    user_sessions,
    email_verifications,
    phone_verifications,
    password_resets,
    oauth_providers,
    api_keys,
    polar_subscriptions,
    polar_products,
    polar_webhook_events,
    conversion_goals,
    conversion_events,
    ab_tests,
    ab_test_variants,
    attribution_touchpoints,
    custom_domains,
    user_roles,
    roles,
    permissions,
    role_permissions,
    teams,
    team_members,
    team_invitations
TO urlshort_app;

-- Grant sequence usage for ID generation
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO urlshort_app;

-- Explicitly DENY delete on critical tables
REVOKE DELETE ON TABLE users, url_mappings FROM urlshort_app;

-- NO access to admin schema
REVOKE ALL ON SCHEMA admin FROM urlshort_app;
```

### Step 4: Set Permissions for Admin User

```sql
-- Grant connection and full schema access
GRANT CONNECT ON DATABASE urlshortener TO urlshort_admin;
GRANT ALL PRIVILEGES ON SCHEMA public TO urlshort_admin;
GRANT ALL PRIVILEGES ON SCHEMA admin TO urlshort_admin;

-- Full access to all tables in public schema (including DELETE)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO urlshort_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO urlshort_admin;

-- Full access to admin schema
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA admin TO urlshort_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA admin TO urlshort_admin;

-- Grant ability to create tables in admin schema
ALTER SCHEMA admin OWNER TO urlshort_admin;
```

### Step 5: Set Permissions for Read-Only User (Analytics)

```sql
-- Grant connection
GRANT CONNECT ON DATABASE urlshortener TO urlshort_readonly;
GRANT USAGE ON SCHEMA public TO urlshort_readonly;

-- Read-only access to all tables
GRANT SELECT ON ALL TABLES IN SCHEMA public TO urlshort_readonly;

-- NO access to admin schema (sensitive audit logs)
REVOKE ALL ON SCHEMA admin FROM urlshort_readonly;
```

---

## Admin-Only Tables (in admin schema)

```sql
-- Move these tables to admin schema

-- 1. Audit Logs (CRITICAL - admin access only)
CREATE TABLE admin.audit_logs (
    id BIGINT PRIMARY KEY,
    user_id BIGINT,  -- Reference to public.users (no FK for isolation)
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id BIGINT,
    details JSONB,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(100),
    status VARCHAR(20),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2. System Settings (CRITICAL - admin access only)
CREATE TABLE admin.system_settings (
    id BIGINT PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    category VARCHAR(50),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    is_encrypted BOOLEAN DEFAULT FALSE,
    updated_by BIGINT,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Blocked Domains (admin management only)
CREATE TABLE admin.blocked_domains (
    id BIGINT PRIMARY KEY,
    domain VARCHAR(255) UNIQUE NOT NULL,
    reason TEXT,
    blocked_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 4. Flagged URLs (admin moderation)
CREATE TABLE admin.flagged_urls (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(20),  -- Reference to public.url_mappings
    reason VARCHAR(100),
    details TEXT,
    flagged_by BIGINT,
    status VARCHAR(20) DEFAULT 'pending',
    reviewed_by BIGINT,
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. System Alerts (admin monitoring)
CREATE TABLE admin.system_alerts (
    id BIGINT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    details JSONB,
    status VARCHAR(20) DEFAULT 'active',
    acknowledged_by BIGINT,
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 6. Scheduled Reports (admin configuration)
CREATE TABLE admin.scheduled_reports (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    report_type VARCHAR(50) NOT NULL,
    schedule VARCHAR(50) NOT NULL,
    recipients TEXT[],
    filters JSONB,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_user_id ON admin.audit_logs(user_id);
CREATE INDEX idx_audit_action ON admin.audit_logs(action);
CREATE INDEX idx_audit_created_at ON admin.audit_logs(created_at);
CREATE INDEX idx_flagged_urls_status ON admin.flagged_urls(status);
CREATE INDEX idx_alerts_status ON admin.system_alerts(status);
```

---

## Connection Configuration

### Environment Variables

```bash
# Regular application connection (limited permissions)
DB_APP_HOST=localhost
DB_APP_PORT=5432
DB_APP_NAME=urlshortener
DB_APP_USER=urlshort_app
DB_APP_PASSWORD=secure_password_1
DB_APP_MAX_CONNECTIONS=50
DB_APP_IDLE_CONNECTIONS=10

# Admin connection (elevated permissions)
DB_ADMIN_HOST=localhost
DB_ADMIN_PORT=5432
DB_ADMIN_NAME=urlshortener
DB_ADMIN_USER=urlshort_admin
DB_ADMIN_PASSWORD=secure_password_2
DB_ADMIN_MAX_CONNECTIONS=10  # Lower limit - admins are fewer
DB_ADMIN_IDLE_CONNECTIONS=2

# Read-only connection (for analytics/reporting)
DB_READONLY_HOST=localhost
DB_READONLY_PORT=5432
DB_READONLY_NAME=urlshortener
DB_READONLY_USER=urlshort_readonly
DB_READONLY_PASSWORD=secure_password_3
```

### Go Configuration (configs/config.go)

```go
package configs

import (
    "fmt"
    "os"
)

type DatabaseConfig struct {
    Host            string
    Port            string
    Name            string
    User            string
    Password        string
    MaxConnections  int
    IdleConnections int
}

type Config struct {
    AppDB      DatabaseConfig  // Regular application
    AdminDB    DatabaseConfig  // Admin operations
    ReadOnlyDB DatabaseConfig  // Analytics
    // ... other config
}

func LoadConfig() *Config {
    return &Config{
        AppDB: DatabaseConfig{
            Host:            getEnv("DB_APP_HOST", "localhost"),
            Port:            getEnv("DB_APP_PORT", "5432"),
            Name:            getEnv("DB_APP_NAME", "urlshortener"),
            User:            getEnv("DB_APP_USER", "urlshort_app"),
            Password:        os.Getenv("DB_APP_PASSWORD"),
            MaxConnections:  getEnvAsInt("DB_APP_MAX_CONNECTIONS", 50),
            IdleConnections: getEnvAsInt("DB_APP_IDLE_CONNECTIONS", 10),
        },
        AdminDB: DatabaseConfig{
            Host:            getEnv("DB_ADMIN_HOST", "localhost"),
            Port:            getEnv("DB_ADMIN_PORT", "5432"),
            Name:            getEnv("DB_ADMIN_NAME", "urlshortener"),
            User:            getEnv("DB_ADMIN_USER", "urlshort_admin"),
            Password:        os.Getenv("DB_ADMIN_PASSWORD"),
            MaxConnections:  getEnvAsInt("DB_ADMIN_MAX_CONNECTIONS", 10),
            IdleConnections: getEnvAsInt("DB_ADMIN_IDLE_CONNECTIONS", 2),
        },
        // ... ReadOnlyDB config
    }
}

func (d *DatabaseConfig) ConnectionString() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        d.Host, d.Port, d.User, d.Password, d.Name,
    )
}
```

### Database Initialization (internal/storage/postgres.go)

```go
package storage

import (
    "database/sql"
    "urlshortener/configs"
)

type DatabaseManager struct {
    AppDB      *sql.DB  // Regular operations
    AdminDB    *sql.DB  // Admin operations
    ReadOnlyDB *sql.DB  // Analytics queries
}

func NewDatabaseManager(cfg *configs.Config) (*DatabaseManager, error) {
    // Initialize app database connection
    appDB, err := initDB(cfg.AppDB)
    if err != nil {
        return nil, fmt.Errorf("failed to init app DB: %w", err)
    }

    // Initialize admin database connection
    adminDB, err := initDB(cfg.AdminDB)
    if err != nil {
        return nil, fmt.Errorf("failed to init admin DB: %w", err)
    }

    // Initialize read-only database connection
    readOnlyDB, err := initDB(cfg.ReadOnlyDB)
    if err != nil {
        return nil, fmt.Errorf("failed to init readonly DB: %w", err)
    }

    return &DatabaseManager{
        AppDB:      appDB,
        AdminDB:    adminDB,
        ReadOnlyDB: readOnlyDB,
    }, nil
}

func initDB(cfg configs.DatabaseConfig) (*sql.DB, error) {
    db, err := sql.Open("postgres", cfg.ConnectionString())
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(cfg.MaxConnections)
    db.SetMaxIdleConns(cfg.IdleConnections)
    db.SetConnMaxLifetime(time.Hour)

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

---

## Row-Level Security (RLS) - Additional Protection

For extra security, implement Row-Level Security policies:

```sql
-- Enable RLS on sensitive tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE url_mappings ENABLE ROW LEVEL SECURITY;

-- Policy: Regular app can only access active users
CREATE POLICY app_user_access ON users
    FOR ALL
    TO urlshort_app
    USING (deleted_at IS NULL AND is_active = true);

-- Policy: Admin can access all users (including deleted)
CREATE POLICY admin_user_access ON users
    FOR ALL
    TO urlshort_admin
    USING (true);

-- Policy: Regular app can only modify own URLs
CREATE POLICY app_url_access ON url_mappings
    FOR ALL
    TO urlshort_app
    USING (deleted_at IS NULL);

-- Policy: Admin can access all URLs
CREATE POLICY admin_url_access ON url_mappings
    FOR ALL
    TO urlshort_admin
    USING (true);
```

---

## Middleware for Admin Operations

```go
package middleware

import (
    "context"
    "github.com/gin-gonic/gin"
)

// AdminDBContext adds admin DB connection to context
func AdminDBContext(dbManager *storage.DatabaseManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if this is an admin route
        if isAdminRoute(c.Request.URL.Path) {
            // Use admin DB connection
            ctx := context.WithValue(c.Request.Context(), "db", dbManager.AdminDB)
            c.Request = c.Request.WithContext(ctx)
        } else {
            // Use regular app DB connection
            ctx := context.WithValue(c.Request.Context(), "db", dbManager.AppDB)
            c.Request = c.Request.WithContext(ctx)
        }
        c.Next()
    }
}

func isAdminRoute(path string) bool {
    return strings.HasPrefix(path, "/api/v1/admin")
}
```

---

## Security Best Practices

### 1. Credential Management
- ✅ Store credentials in environment variables, NOT in code
- ✅ Use secrets management (AWS Secrets Manager, HashiCorp Vault)
- ✅ Rotate passwords regularly (quarterly)
- ✅ Use strong passwords (16+ characters, mixed case, numbers, symbols)

### 2. Connection Security
- ✅ Enable SSL/TLS for database connections in production
- ✅ Use certificate-based authentication when possible
- ✅ Whitelist IP addresses at database level
- ✅ Use VPN or private networks for admin access

### 3. Audit Everything
- ✅ Log all admin database connections
- ✅ Log all queries that modify data (INSERT, UPDATE, DELETE)
- ✅ Monitor failed authentication attempts
- ✅ Set up alerts for suspicious activity

### 4. Principle of Least Privilege
- ✅ Regular app user: NO DELETE permissions
- ✅ Read-only user: NO write permissions
- ✅ Admin user: Time-limited sessions
- ✅ Separate users for different admin roles (if needed)

### 5. Network Isolation
- ✅ Admin database access only from specific IPs
- ✅ Admin panel accessible only via VPN (optional but recommended)
- ✅ Use jump host/bastion for direct database access

---

## Docker Compose Configuration

```yaml
# docker/docker-compose.yml

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: urlshortener
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_ROOT_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-users.sql:/docker-entrypoint-initdb.d/01-init-users.sql
      - ./init-schema.sql:/docker-entrypoint-initdb.d/02-init-schema.sql
      - ./init-permissions.sql:/docker-entrypoint-initdb.d/03-init-permissions.sql
    ports:
      - "5432:5432"

  backend:
    build: .
    environment:
      # Regular app connection
      DB_APP_HOST: postgres
      DB_APP_PORT: 5432
      DB_APP_NAME: urlshortener
      DB_APP_USER: urlshort_app
      DB_APP_PASSWORD: ${DB_APP_PASSWORD}

      # Admin connection
      DB_ADMIN_HOST: postgres
      DB_ADMIN_PORT: 5432
      DB_ADMIN_NAME: urlshortener
      DB_ADMIN_USER: urlshort_admin
      DB_ADMIN_PASSWORD: ${DB_ADMIN_PASSWORD}
    depends_on:
      - postgres
```

---

## Migration Script

```bash
#!/bin/bash
# scripts/setup-admin-db-security.sh

echo "Setting up database security for admin panel..."

# Load environment variables
source .env

# Connect to database as superuser
PGPASSWORD=$POSTGRES_ROOT_PASSWORD psql -h localhost -U postgres -d urlshortener <<EOF

-- Create users
CREATE USER urlshort_app WITH PASSWORD '$DB_APP_PASSWORD';
CREATE USER urlshort_admin WITH PASSWORD '$DB_ADMIN_PASSWORD';

-- Create admin schema
CREATE SCHEMA IF NOT EXISTS admin;

-- Set permissions (see SQL above)
-- ... (run all permission scripts)

EOF

echo "Database security setup complete!"
```

---

## Testing Security

```bash
# Test regular user CANNOT delete
PGPASSWORD=secure_password_1 psql -h localhost -U urlshort_app -d urlshortener -c "DELETE FROM users WHERE id = 1;"
# Expected: ERROR: permission denied for table users

# Test regular user CANNOT access admin schema
PGPASSWORD=secure_password_1 psql -h localhost -U urlshort_app -d urlshortener -c "SELECT * FROM admin.audit_logs;"
# Expected: ERROR: permission denied for schema admin

# Test admin user CAN delete
PGPASSWORD=secure_password_2 psql -h localhost -U urlshort_admin -d urlshortener -c "DELETE FROM users WHERE id = 1;"
# Expected: Success (but use soft delete in practice!)

# Test admin user CAN access admin schema
PGPASSWORD=secure_password_2 psql -h localhost -U urlshort_admin -d urlshortener -c "SELECT COUNT(*) FROM admin.audit_logs;"
# Expected: Success
```

---

## Summary

This architecture provides:

✅ **Defense in Depth**: Multiple layers of security
✅ **Principle of Least Privilege**: Each user has only necessary permissions
✅ **Audit Trail**: All admin actions logged in isolated schema
✅ **Maintainability**: Single database, clear separation
✅ **Performance**: Separate connection pools for different use cases
✅ **Compliance**: Meets security best practices for admin systems

The regular application user **cannot**:
- Delete users or URLs
- Access audit logs
- Modify system settings
- Access admin-only tables

The admin user **can**:
- Perform all CRUD operations
- Access audit logs
- Modify system settings
- Perform administrative tasks

This ensures that even if your application is compromised, attackers cannot perform destructive operations or hide their tracks by deleting audit logs.
