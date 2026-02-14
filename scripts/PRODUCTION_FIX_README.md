# Production 502 Error Fix Guide

## Problem Summary

All admin panel endpoints are returning **502 Bad Gateway** errors because database migrations have NOT been applied to production yet.

**Affected Endpoints:**
- `GET /api/v1/admin/seo/analyses` → 502
- `GET /api/v1/admin/seo/sitemap` → 502
- `GET /api/v1/admin/seo/meta-tags` → 502
- `GET /api/v1/admin/cms/pages` → 502
- `GET /api/v1/admin/content-blocks` → 502
- `GET /api/v1/admin/page-templates` → 502

**Root Cause:**
The backend Go server crashes or returns errors when trying to query tables that don't exist (`static_pages`, `content_blocks`, `meta_tags`, `seo_analyses`, etc.).

## Quick Fix (Recommended)

### Option 1: Automated Fix Script

**On Linux/macOS Production Server:**

```bash
# 1. SSH into production
ssh user@trunc.site

# 2. Navigate to application directory
cd /path/to/URLshorter

# 3. Pull latest code (includes migrations)
git pull origin main

# 4. Run diagnostic script first
./scripts/diagnose-production.sh

# 5. Run automated fix script
./scripts/fix-production-502.sh

# 6. Verify fix
curl https://trunc.site/health
```

**On Windows Production Server:**

```powershell
# 1. Navigate to application directory
cd D:\github\URLshorter

# 2. Pull latest code
git pull origin main

# 3. Run automated fix script
.\scripts\fix-production-502.ps1 -DBHost "localhost" -DBUser "urlshortener" -DBName "urlshortener_db" -DBPassword "your_password"

# 4. Restart backend manually
go run cmd\server\main.go
```

## Manual Fix (If Scripts Don't Work)

### Step 1: Check Current State

```bash
# Connect to database
psql -h localhost -U urlshortener -d urlshortener_db

# Check applied migrations
SELECT version, executed_at FROM schema_migrations ORDER BY version;

# Check existing tables
\dt

# Exit psql
\q
```

### Step 2: Apply Missing Migrations

```bash
# Navigate to project directory
cd /path/to/URLshorter

# Apply migration 017 (CMS tables)
psql -h localhost -U urlshortener -d urlshortener_db -f configs/migrations/017_cms_tables_up.sql

# Apply migration 025 (Activity logs)
psql -h localhost -U urlshortener -d urlshortener_db -f configs/migrations/025_create_activity_logs_up.sql

# Apply migration 027 (SEO tables)
psql -h localhost -U urlshortener -d urlshortener_db -f configs/migrations/027_seo_tables_up.sql

# Record migrations in schema_migrations table
psql -h localhost -U urlshortener -d urlshortener_db <<EOF
INSERT INTO schema_migrations (version, executed_at)
VALUES
    ('017', NOW()),
    ('025', NOW()),
    ('027', NOW())
ON CONFLICT (version) DO NOTHING;
EOF
```

### Step 3: Verify Tables Were Created

```bash
psql -h localhost -U urlshortener -d urlshortener_db -c "
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN (
    'static_pages', 'content_blocks', 'page_templates',
    'meta_tags', 'seo_analyses', 'activity_logs',
    'global_settings', 'robots_txt', 'url_redirects', 'page_seo'
)
ORDER BY table_name;
"
```

**Expected Output (should show all 10 tables):**
```
    table_name
-------------------
 activity_logs
 content_blocks
 global_settings
 meta_tags
 page_seo
 page_templates
 robots_txt
 seo_analyses
 static_pages
 url_redirects
```

### Step 4: Restart Backend

**Using systemd:**
```bash
sudo systemctl restart urlshortener
sudo systemctl status urlshortener
```

**Using Docker:**
```bash
docker restart backend-container-name
docker logs --tail=50 backend-container-name
```

**Manual restart:**
```bash
# Kill old process
pkill -f "server"

# Start new process
cd /path/to/URLshorter
nohup ./server > /var/log/urlshortener/app.log 2>&1 &
```

### Step 5: Verify Fix

**Test health endpoint:**
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy"}
```

**Test admin endpoints (replace TOKEN with your admin token):**
```bash
TOKEN="your-admin-token-here"

# Test SEO endpoints
curl -H "Authorization: Bearer $TOKEN" https://trunc.site/api/v1/admin/seo/analyses
# Expected: {"analyses":[],"total":0,...}

curl -H "Authorization: Bearer $TOKEN" https://trunc.site/api/v1/admin/seo/meta-tags
# Expected: {"meta_tags":[],"total":0,...}

# Test CMS endpoints
curl -H "Authorization: Bearer $TOKEN" https://trunc.site/api/v1/admin/cms/pages
# Expected: {"pages":[],"total":0,...}

curl -H "Authorization: Bearer $TOKEN" https://trunc.site/api/v1/admin/content-blocks
# Expected: {"blocks":[],"total":0,...}
```

**Test frontend:**
1. Visit https://trunc.site/admin
2. Click "SEO Tools" tab → Should load without 502 errors
3. Click "Content Management" tab → Should load without 502 errors
4. Open browser DevTools Console → No 502 errors

## Troubleshooting

### Issue: "relation does not exist" errors in backend logs

**Solution:**
```bash
# Check if migrations were applied
psql -h localhost -U urlshortener -d urlshortener_db -c "SELECT version FROM schema_migrations WHERE version IN ('017', '025', '027');"

# If empty, apply migrations manually (see Step 2 above)
```

### Issue: Backend keeps crashing after applying migrations

**Solution:**
```bash
# Check backend logs for the actual error
journalctl -u urlshortener -n 100 --no-pager
# OR
docker logs --tail=100 backend-container-name

# Common causes:
# 1. Database connection string incorrect in .env
# 2. Database permissions issue
# 3. Port already in use

# Fix database permissions
psql -h localhost -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE urlshortener_db TO urlshortener;"
psql -h localhost -U postgres -d urlshortener_db -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO urlshortener;"
```

### Issue: "Port already in use" error

**Solution:**
```bash
# Find process using port 8080
lsof -ti:8080

# Kill the process
lsof -ti:8080 | xargs kill -9

# Restart backend
systemctl restart urlshortener
```

### Issue: Still getting 502 errors after fix

**Check Nginx/reverse proxy:**
```bash
# Check if backend is reachable from proxy
curl http://localhost:8080/health

# Check Nginx error logs
tail -100 /var/log/nginx/error.log

# Restart Nginx if needed
sudo systemctl restart nginx
```

## Files Modified/Created

**Migration Files (Already in Git):**
- `configs/migrations/017_cms_tables_up.sql` - Creates CMS tables
- `configs/migrations/025_create_activity_logs_up.sql` - Creates activity logs
- `configs/migrations/027_seo_tables_up.sql` - Creates SEO tables

**New Scripts:**
- `scripts/diagnose-production.sh` - Diagnostic script for Linux/macOS
- `scripts/fix-production-502.sh` - Automated fix script for Linux/macOS
- `scripts/fix-production-502.ps1` - Automated fix script for Windows

**Backend Code (Already Deployed):**
- Admin endpoints for SEO tools (commit: 825d7e1)
- Admin endpoints for CMS (commit: ecbabdf)
- Frontend fixes (commits: cf20c78, 2059212)

## Expected Tables After Fix

Migration 017 creates:
- `static_pages` - CMS pages
- `content_blocks` - Reusable content blocks
- `page_templates` - Page templates
- `global_settings` - Global site settings

Migration 025 creates:
- `activity_logs` - User activity tracking

Migration 027 creates:
- `meta_tags` - SEO meta tags
- `seo_analyses` - SEO analysis results
- `page_seo` - Page-specific SEO settings
- `url_redirects` - URL redirect rules
- `robots_txt` - Robots.txt configuration

## Success Criteria

- [ ] Backend server is running (verified with `ps aux` or `systemctl status`)
- [ ] Backend health endpoint returns 200 OK
- [ ] All 6 admin endpoints return 200 (not 502)
- [ ] Frontend admin panel loads all tabs without errors
- [ ] Database has all 10 required tables from migrations 017, 025, 027
- [ ] Backend logs show successful API requests (200 status codes)
- [ ] No errors in Nginx/reverse proxy logs

## Support

If you encounter issues not covered here:
1. Check backend logs: `journalctl -u urlshortener -n 100 --no-pager`
2. Check database connection: `psql -h localhost -U urlshortener -d urlshortener_db -c "\conninfo"`
3. Check Nginx logs: `tail -100 /var/log/nginx/error.log`
4. Verify migrations: `SELECT * FROM schema_migrations ORDER BY version;`
