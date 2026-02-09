# Quick Fix: PostgreSQL Permissions Error

## The Error

```
pq: permission denied for schema public
```

## Quick Fix (5 minutes)

On your Oracle server, run these commands:

### Step 1: Grant Permissions

```bash
# Connect to PostgreSQL as superuser
docker exec -it postgresql-central psql -U postgres -d urlshortener_db

# Grant permissions (run these SQL commands)
GRANT USAGE ON SCHEMA public TO urlshortener_user;
GRANT CREATE ON SCHEMA public TO urlshortener_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO urlshortener_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO urlshortener_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO urlshortener_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO urlshortener_user;

# Verify permissions
SELECT has_schema_privilege('urlshortener_user', 'public', 'CREATE');
-- Should return: t (true)

# Exit
\q
```

### Step 2: Restart Application

```bash
# Pull latest code (includes updated run-migrations.sh)
cd /path/to/URLshorter
git pull origin main

# Rebuild and restart
docker-compose -f docker-compose.oracle.yml build app
docker-compose -f docker-compose.oracle.yml restart app

# Watch logs
docker logs urlshortener-app -f
```

### Expected Output

You should see:
```
🚀 Starting URL Shortener Application...
🔄 Running database migrations...
⏳ Waiting for database to be ready...
✅ Database is ready!
🔑 Attempting to grant schema permissions...
   ⚠️  Could not grant permissions as postgres superuser (this is OK if permissions were already set)
📋 Ensuring schema_migrations table exists...
🔧 Applying migration 001: 001_add_api_keys_up.sql
   ✅ Migration 001 applied successfully
...
🎉 Migration complete!
✅ Application initialization complete!
[INFO] Connected to database successfully
[INFO] Server starting on port 8080
```

## What Changed

The `run-migrations.sh` script now:
- ✅ **Tries** to grant permissions automatically
- ✅ **Continues** if permission grant fails (instead of crashing)
- ✅ **Shows helpful message** if manual setup is needed
- ✅ **Runs migrations** even without superuser access

## Future Deployments

After you manually grant permissions once:
- ✅ New deployments will work automatically
- ✅ No manual intervention needed
- ✅ Migrations run on container startup

## Troubleshooting

### "Database is unavailable - sleeping"
- Check `postgresql-central` container is running: `docker ps | grep postgresql`
- Check network: `docker network inspect shared-db-network`
- Verify DB_HOST=postgresql-central in docker-compose.oracle.yml

### "Migration X failed"
- Check migration logs: `docker logs urlshortener-app --tail 100`
- Verify database connection: `docker exec postgresql-central psql -U urlshortener_user -d urlshortener_db -c '\dt'`
- Check migration file: `cat configs/migrations/XXX_*.sql`

### "Still getting permission denied"
- Verify permissions were actually granted (Step 1)
- Check correct username: Should be `urlshortener_user`
- Try granting as database owner instead of postgres user

## One-Line Fix (If Above Doesn't Work)

```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "GRANT ALL ON SCHEMA public TO urlshortener_user; GRANT CREATE ON SCHEMA public TO urlshortener_user;"
```

Then restart the app container.
