# Admin Panel Implementation Plan

## Executive Summary

This document outlines a comprehensive plan to enhance and complete the Admin Panel for the URL Shortener system. The system already has a foundation with a frontend admin UI and RBAC system, but several features are disabled or incomplete. This plan focuses on completing, enhancing, and securing the admin capabilities.

---

## Current State Analysis

### ✅ What Already Exists

1. **Frontend Admin UI** (`/frontend/src/app/admin/page.tsx`)
   - User management (view, activate/deactivate, delete)
   - CMS management interface
   - Visual page builder
   - SEO tools interface
   - API key management view
   - Activity log viewer
   - System information dashboard

2. **Backend RBAC System** (`/internal/models/rbac.go`, `/internal/services/rbac_service.go`)
   - Role and permission models
   - Team management
   - User role assignments
   - RBAC handlers (currently disabled)

3. **Authentication & Authorization**
   - JWT-based auth with refresh tokens
   - API key support
   - Multi-factor authentication (OTP)
   - OAuth integration

4. **Billing & Subscription Management**
   - Polar integration
   - Plan-based feature gating
   - Usage tracking and limits

### ⚠️ What's Incomplete or Disabled

1. **RBAC Routes** - Commented out in routes.go
2. **CMS Backend Routes** - Not fully connected
3. **Global Settings Handlers** - Disabled
4. **Webhook Admin Routes** - Disabled
5. **Advanced Analytics Admin Features** - Disabled
6. **System-wide reporting and monitoring**

---

## Admin Panel Architecture

### Architecture Principles

1. **Security First**: Admin operations require special permissions and audit logging
2. **Role-Based Access Control (RBAC)**: Granular permissions for different admin levels
3. **Audit Everything**: All admin actions logged with user, timestamp, IP, and details
4. **Real-time Monitoring**: Live dashboards for system health and usage
5. **Self-Service**: Reduce need for database access with comprehensive UI
6. **API-First**: All admin operations accessible via REST API for automation

### Admin Hierarchy

```
Super Admin (System Owner)
├── Full system access
├── Can manage other admins
├── Can configure global settings
├── Can access all data
└── Cannot be deleted

System Admin
├── User management
├── Content management
├── API key management
├── View analytics
└── System monitoring

Support Admin
├── View-only access to users
├── Can assist users (password resets, verification)
├── View analytics
└── Cannot delete or modify critical data

Billing Admin
├── Subscription management
├── Billing dispute resolution
├── Plan configuration
└── Usage reports

Content Admin
├── CMS management
├── SEO tools
├── Page builder
└── Static content updates
```

---

## Core Admin Features

### 1. Enhanced User Management

#### Features
- **User Search & Filtering**
  - Search by email, name, ID
  - Filter by account type, status, plan, registration date
  - Bulk operations support

- **User Details View**
  - Profile information
  - Account status and verification
  - Subscription details and billing history
  - URL ownership (with click stats)
  - Activity history
  - Login history with IPs and devices
  - API keys owned

- **User Actions**
  - Activate/Deactivate accounts
  - Force email verification
  - Force password reset
  - Ban users (with reason)
  - Unban users
  - Delete users (soft delete with retention period)
  - Impersonate user (for support, with audit log)
  - Adjust user limits
  - Upgrade/downgrade plans manually
  - Send system notifications

#### API Endpoints
```
GET    /api/v1/admin/users              # List with filters
GET    /api/v1/admin/users/:id          # User details
PUT    /api/v1/admin/users/:id/status   # Update status
PUT    /api/v1/admin/users/:id/verify   # Force verification
POST   /api/v1/admin/users/:id/reset-password  # Force password reset
POST   /api/v1/admin/users/:id/ban      # Ban user
DELETE /api/v1/admin/users/:id/ban      # Unban user
DELETE /api/v1/admin/users/:id          # Delete user
POST   /api/v1/admin/users/:id/impersonate    # Generate impersonation token
GET    /api/v1/admin/users/:id/activity        # User activity log
GET    /api/v1/admin/users/:id/urls            # User's URLs
PUT    /api/v1/admin/users/:id/limits          # Adjust user limits
PUT    /api/v1/admin/users/:id/plan            # Change subscription
POST   /api/v1/admin/users/:id/notify          # Send notification
POST   /api/v1/admin/users/bulk-export         # Export users as CSV
```

---

### 2. URL Management & Moderation

#### Features
- **Global URL Search**
  - Search by short code, original URL, owner
  - Filter by status, creation date, expiration
  - View URLs from all users

- **URL Moderation**
  - Review flagged URLs
  - Block malicious URLs
  - Delete inappropriate content
  - Quarantine suspicious URLs
  - Whitelist trusted domains

- **URL Actions**
  - View URL details (owner, stats, history)
  - Deactivate/reactivate URLs
  - Transfer ownership
  - Extend expiration
  - View full analytics
  - View click history with details

#### API Endpoints
```
GET    /api/v1/admin/urls                      # List all URLs
GET    /api/v1/admin/urls/:shortCode           # URL details
PUT    /api/v1/admin/urls/:shortCode/status    # Update status
DELETE /api/v1/admin/urls/:shortCode           # Delete URL
PUT    /api/v1/admin/urls/:shortCode/owner     # Transfer ownership
PUT    /api/v1/admin/urls/:shortCode/expiration # Extend expiration
GET    /api/v1/admin/urls/flagged              # List flagged URLs
POST   /api/v1/admin/urls/:shortCode/flag      # Flag URL
POST   /api/v1/admin/urls/:shortCode/unflag    # Unflag URL
POST   /api/v1/admin/domains/block             # Block domain
GET    /api/v1/admin/domains/blocked           # List blocked domains
```

---

### 3. System Analytics & Reporting

#### Features
- **Real-time Dashboard**
  - Active users (last 5min, 1hr, 24hr)
  - URLs created (today, this week, this month)
  - Total clicks (today, this week, this month)
  - System health metrics
  - API response times
  - Error rates
  - Cache hit rates

- **System Reports**
  - User growth trends
  - URL creation trends
  - Click analytics by time period
  - Geographic distribution
  - Device/browser analytics
  - Top performing URLs
  - Revenue metrics (from Polar)
  - Conversion funnel analysis

- **Exportable Reports**
  - CSV/Excel export
  - PDF reports with charts
  - Scheduled email reports
  - Custom date ranges
  - Filterable dimensions

#### API Endpoints
```
GET    /api/v1/admin/analytics/dashboard       # Real-time dashboard
GET    /api/v1/admin/analytics/users           # User analytics
GET    /api/v1/admin/analytics/urls            # URL analytics
GET    /api/v1/admin/analytics/clicks          # Click analytics
GET    /api/v1/admin/analytics/revenue         # Revenue analytics
GET    /api/v1/admin/analytics/geography       # Geographic data
GET    /api/v1/admin/analytics/devices         # Device/browser data
GET    /api/v1/admin/analytics/trends          # Trend data
POST   /api/v1/admin/analytics/export          # Export report
GET    /api/v1/admin/analytics/top-urls        # Top performing URLs
```

---

### 4. Billing & Subscription Administration

#### Features
- **Subscription Management**
  - View all subscriptions
  - Subscription status overview
  - Expired/expiring subscriptions
  - Manual subscription adjustments
  - Refund processing
  - Billing dispute resolution

- **Plan Management**
  - Create/edit subscription plans
  - Set feature limits per plan
  - Price adjustments
  - Plan migration tools
  - Trial period management

- **Revenue Analytics**
  - MRR (Monthly Recurring Revenue)
  - Churn rate
  - Lifetime value (LTV)
  - Revenue by plan
  - Failed payment tracking

- **Polar Integration Management**
  - Webhook event viewer
  - Retry failed webhooks
  - Sync subscription status
  - View Polar logs

#### API Endpoints
```
GET    /api/v1/admin/billing/subscriptions     # All subscriptions
GET    /api/v1/admin/billing/subscriptions/:id # Subscription details
PUT    /api/v1/admin/billing/subscriptions/:id # Update subscription
POST   /api/v1/admin/billing/subscriptions/:id/cancel   # Cancel
POST   /api/v1/admin/billing/subscriptions/:id/refund   # Process refund
GET    /api/v1/admin/billing/plans             # All plans
POST   /api/v1/admin/billing/plans             # Create plan
PUT    /api/v1/admin/billing/plans/:id         # Update plan
DELETE /api/v1/admin/billing/plans/:id         # Delete plan
GET    /api/v1/admin/billing/revenue           # Revenue analytics
GET    /api/v1/admin/billing/webhooks          # Webhook events
POST   /api/v1/admin/billing/webhooks/:id/retry # Retry webhook
POST   /api/v1/admin/billing/sync              # Sync with Polar
```

---

### 5. Role & Permission Management (RBAC)

#### Features
- **Role Management**
  - Create custom roles
  - Edit role permissions
  - View role assignments
  - System roles (protected from deletion)
  - Role templates

- **Permission Management**
  - Granular permissions (resource:action)
  - Permission groups
  - Permission inheritance via teams

- **User Role Assignment**
  - Assign roles to users
  - Time-based role expiration
  - Role history tracking
  - Bulk role assignments

#### API Endpoints
```
GET    /api/v1/admin/rbac/roles                # List roles
POST   /api/v1/admin/rbac/roles                # Create role
GET    /api/v1/admin/rbac/roles/:id            # Role details
PUT    /api/v1/admin/rbac/roles/:id            # Update role
DELETE /api/v1/admin/rbac/roles/:id            # Delete role
GET    /api/v1/admin/rbac/permissions          # List permissions
POST   /api/v1/admin/rbac/roles/:id/permissions # Assign permissions
DELETE /api/v1/admin/rbac/roles/:id/permissions/:permId # Remove permission
POST   /api/v1/admin/rbac/users/:userId/roles  # Assign role to user
DELETE /api/v1/admin/rbac/users/:userId/roles/:roleId # Remove role
GET    /api/v1/admin/rbac/users/:userId/roles  # User's roles
```

---

### 6. API Key Administration

#### Features
- **Global API Key Management**
  - View all API keys across users
  - Search and filter keys
  - Usage statistics per key
  - Rate limit monitoring
  - Revoke compromised keys

- **System API Keys**
  - Create system-level API keys
  - Service-to-service authentication
  - Integration keys for webhooks
  - Key rotation policies

#### API Endpoints
```
GET    /api/v1/admin/api-keys                  # All API keys
GET    /api/v1/admin/api-keys/:id              # Key details
PUT    /api/v1/admin/api-keys/:id/revoke       # Revoke key
DELETE /api/v1/admin/api-keys/:id              # Delete key
GET    /api/v1/admin/api-keys/:id/usage        # Usage stats
POST   /api/v1/admin/api-keys/system           # Create system key
GET    /api/v1/admin/api-keys/system           # List system keys
```

---

### 7. Content Management System (CMS)

#### Features
- **Page Management**
  - Create/edit static pages
  - Page templates
  - Draft/publish workflow
  - Version history
  - Scheduled publishing

- **Media Library**
  - Upload images, files
  - Media organization
  - CDN integration
  - Usage tracking

- **Navigation Management**
  - Menu builder
  - Footer configuration
  - Breadcrumb management

#### API Endpoints
```
GET    /api/v1/admin/cms/pages                 # List pages
POST   /api/v1/admin/cms/pages                 # Create page
GET    /api/v1/admin/cms/pages/:id             # Page details
PUT    /api/v1/admin/cms/pages/:id             # Update page
DELETE /api/v1/admin/cms/pages/:id             # Delete page
GET    /api/v1/admin/cms/pages/:id/versions    # Version history
POST   /api/v1/admin/cms/pages/:id/publish     # Publish page
POST   /api/v1/admin/cms/media                 # Upload media
GET    /api/v1/admin/cms/media                 # List media
DELETE /api/v1/admin/cms/media/:id             # Delete media
```

---

### 8. System Configuration & Settings

#### Features
- **Global Settings**
  - System name, logo, branding
  - Email configuration (SMTP)
  - SMS provider (for OTP)
  - OAuth provider credentials
  - Rate limiting configuration
  - Feature flags
  - Maintenance mode

- **Security Settings**
  - Password policies (length, complexity)
  - Session timeout configuration
  - IP whitelisting/blacklisting
  - Max login attempts
  - 2FA enforcement

- **Performance Settings**
  - Cache TTL configuration
  - Database connection pool settings
  - Redis configuration
  - CDN settings

#### API Endpoints
```
GET    /api/v1/admin/settings                  # All settings
GET    /api/v1/admin/settings/:key             # Get setting
PUT    /api/v1/admin/settings/:key             # Update setting
GET    /api/v1/admin/settings/categories       # Settings by category
PUT    /api/v1/admin/settings/bulk             # Bulk update
POST   /api/v1/admin/settings/reset            # Reset to defaults
```

---

### 9. Activity & Audit Logs

#### Features
- **Comprehensive Logging**
  - All admin actions logged
  - User actions logged
  - System events logged
  - Failed login attempts
  - API call logs

- **Log Viewer**
  - Search and filter logs
  - Real-time log streaming
  - Export logs
  - Log retention policies

- **Audit Trail**
  - Who did what, when, from where
  - Before/after values for changes
  - IP address and device tracking
  - User agent logging

#### API Endpoints
```
GET    /api/v1/admin/audit-logs                # List audit logs
GET    /api/v1/admin/audit-logs/:id            # Log details
GET    /api/v1/admin/audit-logs/user/:userId   # User's audit trail
GET    /api/v1/admin/audit-logs/resource/:type/:id # Resource audit trail
POST   /api/v1/admin/audit-logs/export         # Export logs
DELETE /api/v1/admin/audit-logs/cleanup        # Cleanup old logs
```

---

### 10. System Health & Monitoring

#### Features
- **Health Checks**
  - Database connectivity
  - Redis connectivity
  - External service status (Polar, OAuth)
  - Disk space monitoring
  - Memory usage
  - CPU usage
  - Goroutine count

- **Performance Metrics**
  - Request rate
  - Response times (p50, p95, p99)
  - Error rates
  - Cache hit rates
  - Database query performance

- **Alerting**
  - Email/SMS alerts for critical issues
  - Threshold-based alerting
  - Alert history
  - Alert acknowledgment

#### API Endpoints
```
GET    /api/v1/admin/health                    # Overall health
GET    /api/v1/admin/health/database           # Database health
GET    /api/v1/admin/health/redis              # Redis health
GET    /api/v1/admin/health/services           # External services
GET    /api/v1/admin/metrics                   # System metrics
GET    /api/v1/admin/metrics/performance       # Performance metrics
GET    /api/v1/admin/alerts                    # Active alerts
POST   /api/v1/admin/alerts/:id/acknowledge    # Acknowledge alert
GET    /api/v1/admin/logs/errors               # Recent errors
```

---

## Implementation Phases

### Phase 1: Foundation & Security (Weeks 1-2)

**Goal**: Establish secure admin authentication and authorization

1. **Enable and Complete RBAC System**
   - Uncomment RBAC routes
   - Complete RBAC handlers
   - Test role and permission assignment
   - Create default admin roles (super_admin, system_admin, support_admin)

2. **Admin Authentication Middleware**
   - Create `RequireAdmin` middleware
   - Create `RequirePermission` middleware
   - Add admin role checking to existing auth middleware

3. **Audit Logging System**
   - Create audit log model and table
   - Create audit logging middleware
   - Log all admin actions automatically
   - Create audit log viewer

4. **Database Schema Updates**
   ```sql
   -- Add admin-specific fields
   ALTER TABLE users ADD COLUMN account_type VARCHAR(50) DEFAULT 'user'; -- user, admin, super_admin

   -- Audit logs table
   CREATE TABLE admin_audit_logs (
       id BIGINT PRIMARY KEY,
       user_id BIGINT REFERENCES users(id),
       action VARCHAR(100) NOT NULL,
       resource_type VARCHAR(50),
       resource_id BIGINT,
       details JSONB,
       ip_address INET,
       user_agent TEXT,
       created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE INDEX idx_audit_user_id ON admin_audit_logs(user_id);
   CREATE INDEX idx_audit_action ON admin_audit_logs(action);
   CREATE INDEX idx_audit_created_at ON admin_audit_logs(created_at);
   ```

5. **Testing**
   - Unit tests for RBAC service
   - Integration tests for admin middleware
   - Test audit logging

**Deliverables**:
- ✅ RBAC fully functional
- ✅ Admin middleware protecting routes
- ✅ Audit logging operational
- ✅ Tests passing

---

### Phase 2: Core Admin Features (Weeks 3-4)

**Goal**: Implement essential admin operations

1. **Enhanced User Management**
   - Complete all user management endpoints
   - Add user search and filtering
   - Implement user actions (ban, verify, etc.)
   - Create user details view
   - Add user activity history

2. **URL Management & Moderation**
   - Global URL search endpoint
   - URL moderation features
   - Domain blocking system
   - Flagging system

3. **System Dashboard**
   - Real-time metrics endpoint
   - System health checks
   - Basic analytics dashboard

4. **Frontend Admin UI Updates**
   - Connect existing UI to new backend endpoints
   - Add user management enhancements
   - Add URL management tab
   - Add system dashboard

**Deliverables**:
- ✅ User management fully functional
- ✅ URL moderation operational
- ✅ System dashboard live
- ✅ Frontend connected to backend

---

### Phase 3: Analytics & Reporting (Weeks 5-6)

**Goal**: Comprehensive system analytics and reporting

1. **Admin Analytics Service**
   - System-wide analytics aggregation
   - User growth metrics
   - URL creation trends
   - Click analytics
   - Revenue metrics integration

2. **Reporting System**
   - Report generation engine
   - CSV/PDF export functionality
   - Scheduled reports
   - Email delivery

3. **Frontend Analytics Dashboard**
   - Real-time charts
   - Filterable reports
   - Export functionality
   - Date range selection

**Deliverables**:
- ✅ Analytics service operational
- ✅ Report generation working
- ✅ Frontend analytics dashboard complete

---

### Phase 4: Billing & Subscription Admin (Weeks 7-8)

**Goal**: Complete billing administration capabilities

1. **Subscription Administration**
   - View all subscriptions
   - Manual adjustments
   - Refund processing
   - Dispute resolution

2. **Plan Management**
   - Plan CRUD operations
   - Feature limit configuration
   - Price adjustments

3. **Polar Integration Admin**
   - Webhook event viewer
   - Retry failed webhooks
   - Sync tools

4. **Revenue Analytics**
   - MRR/ARR tracking
   - Churn analysis
   - LTV calculations

**Deliverables**:
- ✅ Subscription admin complete
- ✅ Plan management operational
- ✅ Revenue analytics live

---

### Phase 5: Content & Configuration (Weeks 9-10)

**Goal**: CMS and system configuration tools

1. **Complete CMS System**
   - Enable CMS routes
   - Complete page management
   - Add media library
   - Version control

2. **Global Settings Management**
   - Settings CRUD endpoints
   - Settings categories
   - Settings validation
   - Frontend settings UI

3. **API Key Administration**
   - Global API key management
   - System API keys
   - Usage monitoring

**Deliverables**:
- ✅ CMS fully operational
- ✅ Settings management complete
- ✅ API key admin functional

---

### Phase 6: Monitoring & Alerting (Weeks 11-12)

**Goal**: Production-ready monitoring and alerting

1. **Enhanced Health Checks**
   - Comprehensive health endpoints
   - External service monitoring
   - Resource usage tracking

2. **Performance Monitoring**
   - Request metrics
   - Response time tracking
   - Error rate monitoring
   - Cache performance

3. **Alerting System**
   - Alert configuration
   - Notification delivery (email/SMS)
   - Alert acknowledgment
   - Alert history

4. **Frontend Monitoring Dashboard**
   - Real-time metrics display
   - Alert management UI
   - Performance charts

**Deliverables**:
- ✅ Health monitoring complete
- ✅ Alerting system operational
- ✅ Monitoring dashboard live

---

## Security Considerations

### Authentication & Authorization

1. **Multi-Factor Authentication**
   - Enforce MFA for all admin accounts
   - Support TOTP (Google Authenticator, Authy)
   - Backup codes

2. **Session Management**
   - Shorter session timeouts for admins (30 minutes idle)
   - Require re-authentication for sensitive operations
   - Device tracking and session termination

3. **IP Whitelisting**
   - Optional IP whitelisting for admin access
   - Geographic restrictions
   - VPN enforcement

### Audit & Compliance

1. **Comprehensive Logging**
   - Log all admin actions with full context
   - Log all data access (who viewed what)
   - Immutable audit logs

2. **Data Retention**
   - Configurable retention periods
   - Automatic log archival
   - Compliance with GDPR/CCPA

3. **User Impersonation Safeguards**
   - Require special permission
   - Time-limited impersonation tokens
   - Clear audit trail
   - User notification (optional)

### Data Protection

1. **Sensitive Data Handling**
   - Mask sensitive data in logs
   - Encrypt sensitive settings
   - Secure credential storage

2. **Rate Limiting**
   - Stricter rate limits for admin endpoints
   - Progressive delays for failed auth
   - Account lockout after repeated failures

---

## Technical Architecture

### Backend Structure

```
internal/
├── handlers/
│   └── admin/
│       ├── user_admin_handlers.go
│       ├── url_admin_handlers.go
│       ├── billing_admin_handlers.go
│       ├── analytics_admin_handlers.go
│       ├── rbac_admin_handlers.go
│       ├── cms_admin_handlers.go
│       ├── settings_admin_handlers.go
│       └── monitoring_admin_handlers.go
├── services/
│   └── admin/
│       ├── user_admin_service.go
│       ├── url_admin_service.go
│       ├── billing_admin_service.go
│       ├── analytics_admin_service.go
│       ├── audit_log_service.go
│       └── monitoring_service.go
├── middleware/
│   ├── admin_auth.go
│   ├── admin_audit.go
│   └── admin_rate_limit.go
└── models/
    └── admin/
        ├── audit_log.go
        ├── alert.go
        └── system_setting.go
```

### Frontend Structure

```
frontend/src/app/admin/
├── layout.tsx                   # Admin layout with navigation
├── page.tsx                     # Overview/dashboard
├── users/
│   ├── page.tsx                 # User list
│   └── [id]/page.tsx           # User details
├── urls/
│   ├── page.tsx                 # URL list
│   └── [shortCode]/page.tsx    # URL details
├── billing/
│   ├── subscriptions/page.tsx
│   ├── plans/page.tsx
│   └── revenue/page.tsx
├── analytics/page.tsx
├── rbac/
│   ├── roles/page.tsx
│   └── permissions/page.tsx
├── api-keys/page.tsx
├── cms/
│   ├── pages/page.tsx
│   └── media/page.tsx
├── settings/page.tsx
├── audit-logs/page.tsx
└── monitoring/page.tsx
```

---

## Database Schema Updates

```sql
-- Admin account types
ALTER TABLE users
ADD COLUMN IF NOT EXISTS account_type VARCHAR(50) DEFAULT 'user',
ADD COLUMN IF NOT EXISTS admin_notes TEXT;

CREATE INDEX idx_users_account_type ON users(account_type);

-- Audit logs (comprehensive)
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id BIGINT PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id BIGINT,
    details JSONB,
    changes JSONB,              -- Before/after values
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(100),
    status VARCHAR(20),          -- success, failed, unauthorized
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_user_id ON admin_audit_logs(user_id);
CREATE INDEX idx_audit_action ON admin_audit_logs(action);
CREATE INDEX idx_audit_resource ON admin_audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_created_at ON admin_audit_logs(created_at);

-- System settings
CREATE TABLE IF NOT EXISTS system_settings (
    id BIGINT PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    category VARCHAR(50),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,  -- Can non-admins read this?
    updated_by BIGINT REFERENCES users(id),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_settings_category ON system_settings(category);

-- Blocked domains
CREATE TABLE IF NOT EXISTS blocked_domains (
    id BIGINT PRIMARY KEY,
    domain VARCHAR(255) UNIQUE NOT NULL,
    reason TEXT,
    blocked_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_blocked_domains_domain ON blocked_domains(domain);

-- Flagged URLs
CREATE TABLE IF NOT EXISTS flagged_urls (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(20) REFERENCES url_mappings(short_code),
    reason VARCHAR(100),
    details TEXT,
    flagged_by BIGINT REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',  -- pending, reviewed, dismissed, removed
    reviewed_by BIGINT REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_flagged_urls_status ON flagged_urls(status);
CREATE INDEX idx_flagged_urls_created_at ON flagged_urls(created_at);

-- System alerts
CREATE TABLE IF NOT EXISTS system_alerts (
    id BIGINT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,           -- error_rate, response_time, disk_space, etc.
    severity VARCHAR(20) NOT NULL,       -- info, warning, critical
    message TEXT NOT NULL,
    details JSONB,
    status VARCHAR(20) DEFAULT 'active', -- active, acknowledged, resolved
    acknowledged_by BIGINT REFERENCES users(id),
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_alerts_status ON system_alerts(status);
CREATE INDEX idx_alerts_severity ON system_alerts(severity);
CREATE INDEX idx_alerts_created_at ON system_alerts(created_at);

-- Scheduled reports
CREATE TABLE IF NOT EXISTS scheduled_reports (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    report_type VARCHAR(50) NOT NULL,
    schedule VARCHAR(50) NOT NULL,       -- daily, weekly, monthly
    recipients TEXT[],                   -- Email addresses
    filters JSONB,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_scheduled_reports_next_run ON scheduled_reports(next_run_at) WHERE is_active = TRUE;
```

---

## API Rate Limiting

Admin endpoints should have different rate limiting:

```go
// Regular admin operations
adminRL := middleware.RateLimit(100, time.Minute)  // 100 req/min

// Read-only operations (viewing data)
adminReadRL := middleware.RateLimit(200, time.Minute)  // 200 req/min

// Sensitive operations (user deletion, data export)
adminSensitiveRL := middleware.RateLimit(10, time.Minute)  // 10 req/min
```

---

## Testing Strategy

### Unit Tests
- All admin services with mocked dependencies
- RBAC permission checking logic
- Audit log creation
- Alert threshold detection

### Integration Tests
- Admin API endpoints with real database
- Auth middleware with admin permissions
- Audit log persistence
- Report generation

### E2E Tests
- Admin login flow
- User management workflow
- URL moderation workflow
- Settings update workflow

### Security Tests
- Permission bypass attempts
- SQL injection on admin endpoints
- XSS in admin forms
- CSRF protection
- Rate limit enforcement

---

## Monitoring & Observability

### Metrics to Track
- Admin endpoint usage
- Permission check failures
- Audit log write rate
- Report generation time
- Alert trigger rate
- Admin session duration

### Dashboards (Grafana/Prometheus)
- Admin activity overview
- Permission check latency
- Audit log volume
- System health status
- Alert history

---

## Documentation Requirements

1. **Admin User Guide**
   - How to use each admin feature
   - Screenshots and tutorials
   - Best practices

2. **API Documentation**
   - All admin endpoint documentation
   - Request/response examples
   - Permission requirements

3. **Security Policy**
   - Admin account requirements
   - MFA enforcement
   - Session management
   - Audit log retention

4. **Runbook**
   - Common admin tasks
   - Troubleshooting guide
   - Emergency procedures

---

## Success Criteria

### Phase 1
- ✅ Admin can log in with appropriate permissions
- ✅ All admin actions are audit logged
- ✅ RBAC system fully operational

### Phase 2
- ✅ Admin can manage users (view, edit, ban, delete)
- ✅ Admin can moderate URLs and block domains
- ✅ System dashboard shows real-time metrics

### Phase 3
- ✅ Admin can view comprehensive analytics
- ✅ Reports can be exported as CSV/PDF
- ✅ Scheduled reports delivered via email

### Phase 4
- ✅ Admin can manage all subscriptions
- ✅ Admin can create/edit plans
- ✅ Revenue analytics available

### Phase 5
- ✅ CMS fully functional for content management
- ✅ System settings can be configured via UI
- ✅ API keys can be managed globally

### Phase 6
- ✅ System health monitored continuously
- ✅ Alerts triggered and delivered
- ✅ Performance metrics tracked

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Unauthorized admin access | Critical | Strong auth, MFA, IP whitelisting, audit logs |
| Performance impact of audit logging | Medium | Async logging, log aggregation, indexed tables |
| Data breach via admin panel | Critical | Permission-based access, encryption, audit trail |
| Admin account compromise | Critical | MFA enforcement, session timeout, anomaly detection |
| System overload from reports | Medium | Report queuing, rate limiting, caching |

---

## Future Enhancements (Post-MVP)

1. **AI-Powered Anomaly Detection**
   - Detect unusual user behavior
   - Identify potential abuse
   - Fraud detection

2. **Advanced Analytics**
   - Predictive analytics (churn prediction)
   - Cohort analysis
   - Funnel analysis

3. **Multi-Tenancy**
   - Support for multiple organizations
   - Organization-level admins
   - Data isolation

4. **Workflow Automation**
   - Automated user verification
   - Auto-ban based on rules
   - Automated report generation

5. **Mobile Admin App**
   - Native iOS/Android app
   - Push notifications for critical alerts
   - On-the-go admin access

---

## Conclusion

This admin panel will transform the URL shortener from a user-facing service into a fully manageable, enterprise-ready platform. By implementing comprehensive admin capabilities with strong security and audit trails, system administrators will have complete visibility and control over the platform while maintaining security and compliance standards.

The phased approach ensures steady progress with testable milestones, allowing for iterative refinement based on real-world usage and feedback.
