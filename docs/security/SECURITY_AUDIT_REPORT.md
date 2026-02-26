# Security Audit Report

**Project:** URL Shortener
**Audit Date:** 2026-02-27
**Auditor:** Claude Code (automated static analysis)
**Status:** Findings documented — remediation pending

---

## Table of Contents

- [Summary](#summary)
- [Critical Findings](#critical-findings)
- [High Findings](#high-findings)
- [Medium Findings](#medium-findings)
- [Low Findings](#low-findings)
- [Positive Findings](#positive-findings)
- [Remediation Checklist](#remediation-checklist)

---

## Summary

| Severity | Count | Status |
|----------|-------|--------|
| Critical | 7 | Pending |
| High | 5 | Pending |
| Medium | 4 | Pending |
| Low | 4 | Pending |
| **Total** | **20** | **Pending** |

> **Context:** A previous system was compromised due to exposed database ports and weak passwords. This audit identifies the same class of vulnerabilities and additional ones in this codebase before public launch.

---

## Critical Findings

### C-01 — Hardcoded Real API Tokens in Source Code

**Severity:** Critical
**Status:** [ ] Not Fixed

**Description:**
Real Polar API access tokens and webhook secrets are hardcoded as fallback default values in docker-compose files and committed documentation. These are real credentials that can be used to access the Polar payment platform.

**Affected Files:**
- `docker/docker-compose.local.yml` — lines 18–20
- `docs/fixes/POLAR_TEST_RESULTS.md` — lines 251–253
- `docs/integrations/POLAR_TEST_REPORT.md` — lines 22–24

**Evidence:**
```yaml
# docker/docker-compose.local.yml
- POLAR_ACCESS_TOKEN=${POLAR_ACCESS_TOKEN:-polar_oat_bmzD64gals01kuZuN6dCdEloAIlj43F00klYr1x7Pdr}
- POLAR_WEBHOOK_SECRET=${POLAR_WEBHOOK_SECRET:-whsec_test_mock_secret_for_development}
- POLAR_ORGANIZATION_ID=${POLAR_ORGANIZATION_ID:-65dca2de-3a3e-4a9a-9464-ee6053485626}
```

**Remediation:**
1. Immediately revoke all exposed Polar tokens at the Polar dashboard
2. Generate new tokens and store only in environment variables (never in code)
3. Remove all hardcoded tokens from docker-compose files and docs
4. Scan git history for any other committed secrets: `git log -p | grep -i "polar_oat_"`

---

### C-02 — Weak Default Database Passwords

**Severity:** Critical
**Status:** [ ] Not Fixed

**Description:**
Multiple docker-compose files define weak passwords as default fallback values. These are used when environment variables are not set, meaning any misconfigured deployment will have trivially guessable database credentials.

**Affected Files:**
- `docker/docker-compose.yml` — lines 17, 56, 82, 108
- `docker/docker-compose.local.yml` — lines 12, 71
- `docker/docker-compose.observability.yml` — lines 21, 65, 167
- `docker-compose.oracle.yml` — lines 35, 42
- `docker-compose.prod-local.yml` — lines 31, 75
- `configs/config.go` — line 98
- `.env.oracle.example` — line 5

**Evidence:**
```yaml
DB_PASSWORD: ${DB_PASSWORD:-securepassword}   # not actually secure
DB_PASSWORD: ${DB_PASSWORD:-devpassword}
DB_PASSWORD=URLShortenerSecurePassword123!    # used in production docs
```
```go
// configs/config.go
DBPassword: getEnv("DB_PASSWORD", "password"),   // default is "password"
```

**Remediation:**
1. Remove all default password fallbacks from docker-compose files — fail loudly if not set
2. Change `configs/config.go` default to empty string and add startup check
3. Rotate `URLShortenerSecurePassword123!` if it was ever used in production
4. Use secrets manager (AWS Secrets Manager, Docker secrets) in production

---

### C-03 — SSL/TLS Disabled for All Database Connections

**Severity:** Critical
**Status:** [ ] Not Fixed

**Description:**
Database SSL is explicitly disabled in both the application config default and all docker-compose environment configurations. Database credentials and all query data are transmitted unencrypted.

**Affected Files:**
- `configs/config.go` — line 90
- `docker/docker-compose.yml` — line 23
- `docker-compose.oracle.yml` — line 41
- `docker-compose.prod-local.yml` — line 37

**Evidence:**
```go
// configs/config.go
DBSSLMode: getEnv("DB_SSLMODE", "disable"),
```
```yaml
# docker-compose files
DB_SSLMODE=disable
```

**Remediation:**
1. Change default to `require` in `configs/config.go`
2. Update all production docker-compose files to `DB_SSLMODE=verify-full`
3. For internal Docker networks, use `require` at minimum
4. Generate and distribute proper TLS certificates for database server

---

### C-04 — Database Port Exposed to Host Network

**Severity:** Critical
**Status:** [ ] Not Fixed

**Description:**
PostgreSQL port 5432 is bound to all host interfaces (`0.0.0.0`) rather than localhost only. This makes the database directly accessible from any network the host is connected to — the same vulnerability that caused the previous system compromise.

**Affected Files:**
- `docker/docker-compose.yml` — lines 57–58
- `docker/docker-compose.local.yml` — line 74
- `docker-compose.prod-local.yml` — line 71 (uses 5433)

**Evidence:**
```yaml
ports:
  - "5432:5432"    # exposes to 0.0.0.0 — accessible from internet
  # should be:
  - "127.0.0.1:5432:5432"   # localhost only, for dev access
  # or remove entirely in production
```

**Remediation:**
1. In production: remove PostgreSQL port mapping entirely (containers communicate via internal network)
2. In development: bind to `127.0.0.1:5432:5432` for local access only
3. Configure firewall rules to block port 5432 at the host level as a second layer

---

### C-05 — Redis Exposed Without Authentication

**Severity:** Critical
**Status:** [ ] Not Fixed

**Description:**
Redis is exposed on port 6379 with no password configured, bound to all interfaces. Redis stores session data and rate limiting counters. Unauthenticated access allows reading sessions, manipulating rate limits, and executing Redis commands.

**Affected Files:**
- `docker/docker-compose.yml` — lines 125–127
- `docker/docker-compose.local.yml` — line 89
- `configs/config.go` — line 88 (empty default password)

**Evidence:**
```yaml
redis:
  ports:
    - "6379:6379"    # no password, no localhost binding
```
```go
// configs/config.go
RedisPassword: getEnv("REDIS_PASSWORD", ""),   // empty = no auth
```

**Note:** `docker-compose.oracle.yml` and `docker-compose.prod-local.yml` correctly bind to `127.0.0.1` — this pattern should be applied everywhere.

**Remediation:**
1. Set a strong `REDIS_PASSWORD` and pass via `--requirepass` to Redis
2. In production: remove Redis port mapping entirely
3. In development: bind to `127.0.0.1:6379:6379`
4. Change `configs/config.go` to require a non-empty `REDIS_PASSWORD`

---

### C-06 — SQL Injection via Unvalidated Dynamic Queries

**Severity:** Critical
**Status:** [ ] Not Fixed

**Description:**
Multiple functions construct SQL queries using `fmt.Sprintf` with table names, column names, and WHERE clauses that are not validated. If any caller passes user-controlled input, these are directly exploitable SQL injection vulnerabilities.

**Affected Files:**
- `internal/storage/postgres_optimized.go` — lines 270–276 (`BatchInsert`)
- `internal/storage/postgres_optimized.go` — line 315 (`VacuumAnalyze`)
- `internal/services/cms.go` — line 269
- `internal/services/global_settings.go` — line 226
- `internal/storage/postgres.go` — line 1215

**Evidence:**
```go
// postgres_optimized.go — tableName and columns not validated
func (p *PostgresOptimizedStorage) BatchInsert(ctx context.Context, tableName string, columns []string, values [][]interface{}) error {
    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, joinStrings(columns, ", "))

// postgres_optimized.go — tableName not validated
func (p *PostgresOptimizedStorage) VacuumAnalyze(tableName string) error {
    query := fmt.Sprintf("VACUUM ANALYZE %s", tableName)

// cms.go — whereClause constructed dynamically
query := fmt.Sprintf("UPDATE static_pages SET %s %s", strings.Join(setParts, ", "), whereClause)
```

**Remediation:**
1. Add an allowlist of valid table names and validate before use
2. Add an allowlist of valid column names and validate before use
3. Audit all callers of `BatchInsert` and `VacuumAnalyze` to confirm no user input reaches them
4. Replace dynamic WHERE clause construction with parameterized query patterns

---

### C-07 — Weak JWT Secret with No Strength Validation

**Severity:** Critical
**Status:** [ ] Not Fixed

**Description:**
The default JWT secret is a human-readable placeholder string. There is no startup validation to enforce that the secret has been changed or meets minimum entropy requirements. A weak JWT secret allows token forgery.

**Affected Files:**
- `configs/config.go` — line 110
- `.env.example` — line 45
- `.env.oracle.example` — line 25

**Evidence:**
```go
// configs/config.go
JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
```
```env
# .env.example
JWT_SECRET=your-secret-key-change-this-in-production
```

**Remediation:**
1. Change default to empty string and add startup validation:
   ```go
   if len(cfg.JWTSecret) < 32 {
       log.Fatal("JWT_SECRET must be at least 32 characters")
   }
   ```
2. Generate strong secret: `openssl rand -base64 64`
3. Rotate any tokens signed with the weak default secret

---

## High Findings

### H-01 — Adminer Database UI Exposed Without Protection

**Severity:** High
**Status:** [ ] Not Fixed

**Description:**
Adminer (a web-based database administration tool) is exposed on port 8081 with no authentication layer. Anyone who can reach this port can log into the database.

**Affected Files:**
- `docker/docker-compose.yml` — line 104
- `docker-compose.oracle.yml` — line 104

**Remediation:**
1. Remove Adminer from all production docker-compose configurations
2. If needed for development, bind to `127.0.0.1:8081:8080` only
3. Add HTTP Basic Auth via a reverse proxy if Adminer must be remotely accessible

---

### H-02 — SSRF Risk in URL Validation

**Severity:** High
**Status:** [ ] Not Fixed

**Description:**
The URL shortener accepts destination URLs without checking if they point to internal/private network resources. An attacker could create short URLs pointing to internal services (`http://192.168.1.1`, `http://10.0.0.1`, `http://169.254.169.254` AWS metadata) and use the server as a proxy to probe internal infrastructure.

**Affected Files:**
- `internal/utils/validation.go` — lines 184–200

**Evidence:**
```go
func ValidateURL(url string) error {
    urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
    // No check for private IP ranges or internal hostnames
    return nil
}
```

**Remediation:**
1. After parsing URL, resolve the hostname and check against blocked ranges:
   - `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` (RFC 1918)
   - `127.0.0.0/8` (loopback)
   - `169.254.0.0/16` (link-local / AWS metadata)
   - `::1` (IPv6 loopback)
2. Block `localhost` and other internal hostnames
3. Consider using a library like `go-ssrf` for comprehensive protection

---

### H-03 — In-Memory Rate Limiting Ineffective in Multi-Instance Deployments

**Severity:** High
**Status:** [ ] Not Fixed

**Description:**
Rate limiting state is stored in a Go in-memory map. When running multiple application instances (load balanced), each instance has independent state, effectively multiplying the rate limit by the number of instances.

**Affected Files:**
- `internal/middleware/rate_limit.go` — lines 45–46

**Evidence:**
```go
var visitors = make(map[string]*visitor)
```

**Remediation:**
1. Move rate limiting state to Redis (already available in the stack)
2. Use a library like `go-redis-rate` for distributed rate limiting
3. Alternatively, use a reverse proxy (Nginx, Cloudflare) for rate limiting at the edge

---

### H-04 — EKS Kubernetes API Endpoint Publicly Accessible

**Severity:** High
**Status:** [ ] Not Fixed

**Description:**
The Kubernetes control plane API endpoint is configured with public access enabled. This exposes the cluster management API to the internet.

**Affected Files:**
- `terraform/environments/production/main.tf` — line 108

**Evidence:**
```terraform
cluster_endpoint_public_access = true
```

**Remediation:**
1. Set `cluster_endpoint_public_access = false` if VPN/bastion access is available
2. If public access is required, add `cluster_endpoint_public_access_cidrs` to restrict to known IPs
3. Enable audit logging for all API server access

---

### H-05 — Credentials Visible in Process Environment and Logs

**Severity:** High
**Status:** [ ] Not Fixed

**Description:**
Database passwords and API keys are passed as environment variables, making them visible in process listings (`/proc/PID/environ`), container inspection (`docker inspect`), and potentially application logs.

**Affected Files:**
- All docker-compose files — `DATABASE_URL` with embedded password
- `internal/storage/postgres.go` — line 28 (DSN with password)

**Remediation:**
1. Use Docker secrets (`docker secret create`) for sensitive values in Swarm/production
2. Use AWS Secrets Manager and inject at runtime instead of environment variables
3. Ensure connection string passwords are not logged

---

## Medium Findings

### M-01 — CORS Configuration Hardcoded in Source

**Severity:** Medium
**Status:** [ ] Not Fixed

**Description:**
Production domain names are hardcoded in the CORS middleware source code instead of being driven by environment configuration. Any domain change requires a code deployment.

**Affected Files:**
- `internal/middleware/cors.go` — lines 11–18

**Evidence:**
```go
config.AllowOrigins = []string{
    "http://localhost:3000",
    "https://trunc.site",
    "https://www.trunc.site",
    "https://api.trunc.site",
}
```

**Remediation:**
1. Read allowed origins from `CORS_ALLOWED_ORIGINS` environment variable
2. Parse as comma-separated list to support multiple origins

---

### M-02 — No Startup Validation of Critical Secrets

**Severity:** Medium
**Status:** [ ] Not Fixed

**Description:**
The application starts successfully even with default insecure values for critical secrets (JWT secret, database password). There is no fail-fast mechanism to prevent insecure deployments.

**Affected Files:**
- `configs/config.go` — startup/validation logic missing

**Remediation:**
Add a `ValidateProductionConfig()` function called at startup:
```go
func (c *Config) ValidateProductionConfig() error {
    if c.JWTSecret == "your-secret-key-change-this-in-production" || len(c.JWTSecret) < 32 {
        return errors.New("JWT_SECRET is insecure or not set")
    }
    if c.DBPassword == "" || c.DBPassword == "password" || c.DBPassword == "securepassword" {
        return errors.New("DB_PASSWORD is insecure or not set")
    }
    if c.RedisPassword == "" {
        return errors.New("REDIS_PASSWORD is not set")
    }
    return nil
}
```

---

### M-03 — Dynamic SQL WHERE Clause Construction in CMS

**Severity:** Medium
**Status:** [ ] Not Fixed

**Description:**
CMS and global settings services construct SQL `SET` clauses and `WHERE` clauses dynamically. While column names likely originate from the application, the pattern is fragile and could introduce SQL injection if any user-controlled value is incorporated.

**Affected Files:**
- `internal/services/cms.go` — line 269
- `internal/services/global_settings.go` — line 226

**Remediation:**
1. Audit all paths that construct `whereClause` — confirm no user input is included
2. Add column name allowlists for all dynamic SET clause construction
3. Consider switching to a query builder library (e.g., `squirrel`) to enforce parameterization

---

### M-04 — Email Validation Regex Not RFC-Compliant

**Severity:** Medium
**Status:** [ ] Not Fixed

**Description:**
Email validation uses a simplified regex that may reject valid addresses or accept structurally invalid ones, potentially affecting user registration flow.

**Affected Files:**
- `internal/utils/validation.go` — line 164

**Evidence:**
```go
emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
```

**Remediation:**
1. Use `net/mail.ParseAddress()` for basic RFC 5322 validation
2. For production, consider verification-based validation (send confirmation email) rather than relying on regex

---

## Low Findings

### L-01 — Development Docker Containers Run as Root

**Severity:** Low
**Status:** [ ] Not Fixed

**Affected Files:**
- `Dockerfile.dev` — no `USER` directive
- `Dockerfile.frontend.dev` — no `USER` directive

**Remediation:** Add `RUN adduser -D appuser && USER appuser` to development Dockerfiles. Note: production Dockerfile already handles this correctly.

---

### L-02 — API Keys Allow Wildcard Permissions

**Severity:** Low
**Status:** [ ] Not Fixed

**Affected Files:**
- `internal/middleware/api_auth.go` — lines 79–121

**Evidence:**
```go
if perm == "*" || perm == permission {
```

**Remediation:** Remove wildcard permission support. Require explicit permission grants for each API key.

---

### L-03 — Inconsistent Port Binding Across Compose Files

**Severity:** Low
**Status:** [ ] Not Fixed

**Description:**
Production-oriented compose files (`docker-compose.oracle.yml`, `docker-compose.prod-local.yml`) correctly bind Redis to `127.0.0.1`, but local development files bind to `0.0.0.0`. This inconsistency increases risk of developers accidentally running the insecure configuration in production.

**Remediation:** Standardize all compose files to use `127.0.0.1` binding for non-application ports.

---

### L-04 — Phone Validation Using Fragile Regex

**Severity:** Low
**Status:** [ ] Not Fixed

**Affected Files:**
- `internal/utils/validation.go` — lines 11–20

**Remediation:** Use `google/libphonenumber` Go port (`nyaruka/phonenumbers`) for robust international phone number validation.

---

## Positive Findings

These security controls are already in place and should be preserved:

| # | Finding | Location |
|---|---------|----------|
| P-01 | Production Docker runs as non-root user (`appuser`) | `docker/Dockerfile` |
| P-02 | RDS encrypted at rest (`storage_encrypted = true`) | `terraform/.../main.tf:195` |
| P-03 | RDS not publicly accessible (`publicly_accessible = false`) | `terraform/.../main.tf:204` |
| P-04 | RDS deletion protection enabled | `terraform/.../main.tf:225` |
| P-05 | VPC Flow Logs enabled | `terraform/.../main.tf:82-84` |
| P-06 | Private subnets for all databases | Terraform VPC config |
| P-07 | Security groups restrict DB access to EKS nodes only | `terraform/.../main.tf:290-324` |
| P-08 | Docker resource limits configured | Multiple compose files |
| P-09 | Admin routes disabled by default (`ENABLE_ADMIN_ROUTES=false`) | Routes config |
| P-10 | S3 versioning and encryption enabled | Terraform S3 config |

---

## Remediation Checklist

Use this checklist to track progress when implementing fixes.

### Phase 1 — Critical (Before Any Public Exposure)

- [ ] **C-01** Revoke and rotate all exposed Polar API tokens
- [ ] **C-01** Remove hardcoded tokens from docker-compose files and docs
- [ ] **C-02** Remove weak password defaults from all docker-compose files
- [ ] **C-02** Rotate `URLShortenerSecurePassword123!` if used in production
- [ ] **C-03** Change `DB_SSLMODE` default to `require` in `configs/config.go`
- [ ] **C-03** Update all production compose files to use SSL
- [ ] **C-04** Remove PostgreSQL port mapping from production compose files
- [ ] **C-04** Bind to `127.0.0.1` in development compose files
- [ ] **C-05** Add Redis password and configure `requirepass`
- [ ] **C-05** Remove Redis port mapping from production compose files
- [ ] **C-06** Add table/column name allowlist validation to `BatchInsert` and `VacuumAnalyze`
- [ ] **C-07** Add JWT secret strength validation at startup

### Phase 2 — High (Before Public Launch)

- [ ] **H-01** Remove Adminer from production docker-compose files
- [ ] **H-02** Add SSRF protection to URL validation (block private IP ranges)
- [ ] **H-03** Move rate limiting state to Redis
- [ ] **H-04** Restrict EKS endpoint access by CIDR or disable public access
- [ ] **H-05** Evaluate Docker secrets or secrets manager for credentials

### Phase 3 — Medium (Within First Week)

- [ ] **M-01** Move CORS origins to `CORS_ALLOWED_ORIGINS` environment variable
- [ ] **M-02** Add `ValidateProductionConfig()` startup check
- [ ] **M-03** Audit and document all dynamic SQL construction paths
- [ ] **M-04** Replace email regex with `net/mail.ParseAddress()`

### Phase 4 — Low (Ongoing Hardening)

- [ ] **L-01** Add non-root user to development Dockerfiles
- [ ] **L-02** Remove wildcard `*` API key permission support
- [ ] **L-03** Standardize port bindings across all compose files
- [ ] **L-04** Replace phone regex with `nyaruka/phonenumbers` library

---

*Report generated by automated static analysis on 2026-02-27. Findings should be verified by a human reviewer before remediation. Mark each item complete and update the Status field in the findings above when resolved.*
