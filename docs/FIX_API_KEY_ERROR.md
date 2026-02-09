# Fix API Key Creation 500 Error

## Problem Summary

**Issue:** Users can login successfully but get `500 Internal Server Error` when trying to create API keys.

**Root Cause:** The `api_keys` and `api_key_usage` tables don't exist in the production database.

**Why:** The database was initialized before the API key feature was added, and migrations haven't been applied.

## Quick Fix (Recommended)

### Option 1: Run Diagnostic Script (Easiest)

We've created diagnostic scripts that will check and fix the database automatically.

**On Linux/macOS:**
```bash
cd /path/to/URLshorter
chmod +x scripts/diagnose-and-fix-db.sh
./scripts/diagnose-and-fix-db.sh
```

**On Windows:**
```powershell
cd D:\github\URLshorter
.\scripts\diagnose-and-fix-db.ps1
```

### Option 2: Restart App Container (Simplest)

The app will run migrations automatically on startup:

```bash
docker-compose -f docker-compose.oracle.yml restart app

# Watch logs to see migration output
docker logs urlshortener-app -f
```

Look for output like:
```
🔄 Running database migrations...
🔧 Applying migration 001: 001_add_api_keys_up.sql
✅ Migration 001 applied successfully
🔧 Applying migration 003: 003_fix_api_keys_id_type_up.sql
✅ Migration 003 applied successfully
🎉 Migration complete!
```

### Option 3: Manually Run Migrations

If automatic migrations don't work:

```bash
# Method 1: Run migration script inside container
docker exec urlshortener-app /app/scripts/run-migrations.sh

# Method 2: Apply migrations manually with psql
docker exec -i postgresql-central psql \
  -U urlshortener_user \
  -d urlshortener_db \
  < configs/migrations/001_add_api_keys_up.sql

docker exec -i postgresql-central psql \
  -U urlshortener_user \
  -d urlshortener_db \
  < configs/migrations/003_fix_api_keys_id_type_up.sql
```

## Verification Steps

After applying the fix, verify the tables exist:

### Check Tables

```bash
docker exec -it postgresql-central psql -U urlshortener_user -d urlshortener_db
```

Then run:
```sql
-- List all tables
\dt

-- Check api_keys table structure
\d api_keys

-- Check api_key_usage table structure
\d api_key_usage

-- Check url_mappings has user_id column
\d url_mappings

-- Exit
\q
```

### Expected Output

You should see:
- ✅ `api_keys` table exists
- ✅ `api_key_usage` table exists
- ✅ `url_mappings` has `user_id` column
- ✅ `users` table exists

### Test API Key Creation

1. **Login** to https://trunc.site/login
2. **Navigate** to API Keys page
3. **Click** "Create New API Key"
4. **Fill in** name and scopes
5. **Click** "Create"

**Expected:** API key is created successfully ✅
**NOT:** 500 Internal Server Error ❌

## What Was Fixed

### Code Changes

**File:** `internal/storage/postgres.go`

Added API keys table creation to the `initAuthenticationTables()` method:

```go
// API Keys table - for programmatic access
`CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(100) NOT NULL,
    permissions JSONB DEFAULT '[]'::jsonb,
    rate_limit INTEGER DEFAULT 1000,
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    last_used_ip VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)`,

// API Key Usage table - for tracking and analytics
`CREATE TABLE IF NOT EXISTS api_key_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
)`,
```

**Why this matters:**
- Future deployments will automatically have the correct schema
- New database instances won't need manual migrations
- Reduces deployment complexity

### Scripts Created

1. **`scripts/diagnose-and-fix-db.sh`** - Bash script for Linux/macOS
2. **`scripts/diagnose-and-fix-db.ps1`** - PowerShell script for Windows

These scripts:
- Check database connectivity
- Verify all required tables exist
- Check all required columns exist
- Automatically attempt to fix issues
- Provide clear diagnostic output

## Deployment Instructions

### For Production (Oracle Server)

1. **Pull latest code:**
   ```bash
   cd /path/to/URLshorter
   git pull origin main
   ```

2. **Rebuild the app container:**
   ```bash
   docker-compose -f docker-compose.oracle.yml build app
   ```

3. **Restart the app (migrations run automatically):**
   ```bash
   docker-compose -f docker-compose.oracle.yml up -d app
   ```

4. **Verify migrations ran:**
   ```bash
   docker logs urlshortener-app --tail 100 | grep -E "(Migration|🔧|✅)"
   ```

5. **Check database:**
   ```bash
   ./scripts/diagnose-and-fix-db.sh
   ```

### For Development (Local)

1. **Pull latest code:**
   ```bash
   git pull origin main
   ```

2. **Restart services:**
   ```bash
   docker-compose -f docker/docker-compose.local.yml restart app
   ```

3. **Verify:**
   ```bash
   ./scripts/diagnose-and-fix-db.sh
   ```

## Troubleshooting

### Issue: Migrations Don't Run on Startup

**Symptom:** No migration messages in logs

**Solution:**
1. Check if migration script exists:
   ```bash
   docker exec urlshortener-app ls -la /app/scripts/run-migrations.sh
   ```

2. Manually run migrations:
   ```bash
   docker exec urlshortener-app /app/scripts/run-migrations.sh
   ```

### Issue: Migrations Fail with Errors

**Symptom:** Error messages in migration logs

**Check database logs:**
```bash
docker logs postgresql-central --tail 100
```

**Check migration status:**
```bash
docker exec -it postgresql-central psql -U urlshortener_user -d urlshortener_db -c "SELECT * FROM schema_migrations ORDER BY version;"
```

**Reset specific migration:**
```bash
# Remove migration record (allows re-running)
docker exec -it postgresql-central psql -U urlshortener_user -d urlshortener_db -c "DELETE FROM schema_migrations WHERE version = '001';"

# Re-run migration
docker exec urlshortener-app /app/scripts/run-migrations.sh
```

### Issue: Tables Exist But Still Get 500 Error

**Check application logs:**
```bash
docker logs urlshortener-app --tail 200 | grep -i error
```

**Common causes:**
1. **JWT_SECRET not set** - Check environment variables
2. **Database connection issues** - Verify DB_HOST, DB_USER, DB_PASSWORD
3. **Permission issues** - Verify user has INSERT privileges
4. **Schema mismatch** - Verify table structure matches expected schema

**Verify environment variables:**
```bash
docker exec urlshortener-app env | grep -E "(DB_|JWT_|SERVER_)"
```

## Prevention

To prevent this issue in the future:

1. **Always run migrations after code updates:**
   ```bash
   git pull
   docker-compose build
   docker-compose up -d
   ```

2. **Monitor migration logs:**
   ```bash
   docker logs urlshortener-app --tail 50
   ```

3. **Use the diagnostic script after deployments:**
   ```bash
   ./scripts/diagnose-and-fix-db.sh
   ```

4. **Test in staging first:**
   - Deploy to staging environment
   - Run full test suite
   - Verify API key creation works
   - Then deploy to production

## Summary

**Before:**
- ❌ API key creation fails with 500 error
- ❌ `api_keys` table doesn't exist
- ❌ `api_key_usage` table doesn't exist

**After:**
- ✅ API key creation works
- ✅ All required tables exist
- ✅ Future deployments will be automatic
- ✅ Diagnostic tools available for verification

**Next Steps:**
1. Run the diagnostic script
2. Verify all tables exist
3. Test API key creation
4. Monitor for any other issues
