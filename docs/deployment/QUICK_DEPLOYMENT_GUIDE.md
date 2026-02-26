# Quick Deployment Guide - Database Schema Migration

## TL;DR

```bash
# On Oracle server:
cd /path/to/URLshorter
git pull origin main
docker exec urlshortener-app /app/scripts/run-migrations.sh
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';"
# Expected: 69-70 tables (was 24 before)
docker-compose -f docker-compose.oracle.yml restart app
docker logs urlshortener-app -f --tail=100
```

## What This Does

Creates **46 missing database tables** in 8 categories:
- 5 RBAC tables (roles, permissions)
- 3 Team management tables
- 9 URL collaboration tables
- 2 Custom domain tables
- 8 CMS tables
- 7 Billing tables
- 14 Advanced analytics tables
- Extended url_mappings with 13 new columns

## Before & After

**Before:**
- ❌ 24 tables
- ❌ Missing: RBAC, teams, billing, CMS, advanced analytics
- ❌ Features don't work

**After:**
- ✅ 70 tables
- ✅ All features supported
- ✅ Migration-based schema management

## Verification Commands

### Check Migration Status
```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT version, applied_at FROM schema_migrations ORDER BY version;"
```

### Check Table Count
```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT COUNT(*) as table_count FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';"
```

### List All Tables
```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name;"
```

### Verify RBAC Tables
```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('roles', 'permissions', 'role_permissions', 'user_roles', 'oauth_providers') ORDER BY table_name;"
```

### Verify Team Tables
```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%team%' ORDER BY table_name;"
```

### Verify Billing Tables
```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('subscriptions', 'billing_events', 'usage_tracking', 'payment_methods', 'coupons', 'coupon_redemptions') ORDER BY table_name;"
```

### Verify CMS Tables
```bash
docker exec -it postgresql-central psql -U postgres -d urlshortener_db -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('static_pages', 'page_revisions', 'media_files', 'global_settings', 'content_blocks', 'page_templates', 'navigation_menus') ORDER BY table_name;"
```

### Check Application Logs
```bash
docker logs urlshortener-app -f --tail=100
# Should see no "table does not exist" errors
```

## Rollback (if needed)

```bash
# Rollback to version 011 (before these migrations)
docker exec urlshortener-app /app/scripts/migrate.sh rollback-to 011

# Or rollback one at a time in reverse order
docker exec urlshortener-app /app/scripts/migrate.sh down 020
docker exec urlshortener-app /app/scripts/migrate.sh down 019
# ... etc
```

## Expected Migration Output

```
Running migrations...
✓ Migration 012_auth_extended_up.sql (skipped - empty)
✓ Migration 013_rbac_tables_up.sql (5 tables)
✓ Migration 014_team_management_up.sql (3 tables)
✓ Migration 015_url_collaboration_up.sql (9 tables)
✓ Migration 016_custom_domains_up.sql (2 tables)
✓ Migration 017_cms_tables_up.sql (8 tables)
✓ Migration 018_billing_tables_up.sql (7 tables)
✓ Migration 019_advanced_analytics_up.sql (14 tables)
✓ Migration 020_url_mappings_extended_up.sql (extended)
Migrations completed successfully!
```

## Common Issues & Solutions

### Issue: "relation already exists"
**Cause:** Tables already created manually
**Solution:** Safe - migrations use `IF NOT EXISTS`

### Issue: "foreign key constraint violation"
**Cause:** Missing parent tables
**Solution:** Run migrations in order (012-020)

### Issue: "function generate_snowflake_id does not exist"
**Cause:** Billing migration expects Snowflake ID function
**Solution:** Safe - migration falls back to BIGSERIAL

### Issue: Application won't start after migration
**Cause:** Check logs for specific error
**Solution:**
```bash
docker logs urlshortener-app --tail=50
docker-compose -f docker-compose.oracle.yml restart app
```

## Safety Features

✅ **Idempotent:** Can run multiple times safely
✅ **Non-destructive:** No existing tables dropped
✅ **Reversible:** Every migration has rollback
✅ **Foreign key safe:** Proper CASCADE/SET NULL
✅ **Indexed:** All frequently queried columns indexed

## Files Created

18 new migration files:
- `012_auth_extended_*.sql` (empty placeholder)
- `013_rbac_tables_*.sql` (RBAC system)
- `014_team_management_*.sql` (Teams)
- `015_url_collaboration_*.sql` (Collaboration)
- `016_custom_domains_*.sql` (Custom domains)
- `017_cms_tables_*.sql` (CMS)
- `018_billing_tables_*.sql` (Billing)
- `019_advanced_analytics_*.sql` (Analytics)
- `020_url_mappings_extended_*.sql` (URL extensions)

## Success Checklist

- [ ] Git pull completed
- [ ] Migrations executed
- [ ] Table count = 69-70 (was 24)
- [ ] No migration errors
- [ ] Application restarted
- [ ] No "table does not exist" errors
- [ ] RBAC tables verified (5 tables)
- [ ] Team tables verified (3 tables)
- [ ] Billing tables verified (7 tables)
- [ ] CMS tables verified (8 tables)
- [ ] Application logs clean

## Next Steps After Migration

1. **Test Core Functionality:**
   - Create a short URL
   - View analytics
   - Test user authentication

2. **Test New Features:**
   - Create a team
   - Assign roles/permissions
   - Create CMS page
   - View billing dashboard

3. **Monitor Performance:**
   - Check query performance
   - Monitor database connections
   - Watch application logs

## Support

If you encounter issues:
1. Check `MIGRATION_SUMMARY.md` for detailed info
2. Review migration logs
3. Check database logs: `docker logs postgresql-central`
4. Verify foreign key constraints
5. Consider rollback if needed

---

**Quick Reference:**
- **Before:** 24 tables
- **After:** 70 tables
- **Added:** 46 tables in 9 migrations
- **Time:** ~30 seconds to run all migrations
- **Rollback:** Safe and reversible
