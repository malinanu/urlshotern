# Deploy RBAC Frontend Updates

## Step 1: Fix CMS Permissions (Do This First!)

### Option A: Using psql from your local machine
```bash
# SSH tunnel to your production database (if needed)
ssh -L 5432:localhost:5432 user@your-server

# Then run the SQL file
psql -h localhost -U urlshortener -d urlshortener_db -f fix-cms-permissions.sql
```

### Option B: Direct SSH and Docker
```bash
# SSH into your production server
ssh user@your-server

# Find your PostgreSQL container
docker ps | grep postgres

# Execute the SQL file
docker exec -i <postgres-container-name> psql -U urlshortener -d urlshortener_db < fix-cms-permissions.sql

# Or run interactively
docker exec -it <postgres-container-name> psql -U urlshortener -d urlshortener_db

# Then manually run the queries from fix-cms-permissions.sql
```

### Option C: Using Portainer
1. Open Portainer at your server
2. Go to Containers
3. Find your PostgreSQL container
4. Click "Console" → "Connect"
5. Run: `psql -U urlshortener -d urlshortener_db`
6. Copy and paste the queries from `fix-cms-permissions.sql`

## Step 2: Clear Redis Cache

After fixing permissions, clear the Redis cache:

```bash
# SSH into your server
ssh user@your-server

# Find Redis container
docker ps | grep redis

# Clear the cache
docker exec -it <redis-container-name> redis-cli FLUSHDB
```

## Step 3: Deploy Frontend Changes

### Option A: Using Git on Production Server (Recommended)

```bash
# SSH into your production server
ssh user@your-server

# Navigate to your project directory
cd /path/to/URLshorter

# Pull the latest changes
git pull origin main

# Rebuild the frontend container via Portainer or docker-compose
docker-compose -f docker/docker-compose.yml up -d --build frontend
```

### Option B: Push Changes and Rebuild

1. **Commit your local changes:**
```bash
git add .
git commit -m "feat: Add RBAC user management UI to admin panel"
git push origin main
```

2. **On your production server:**
```bash
# Pull changes
cd /path/to/URLshorter
git pull origin main

# Rebuild and restart containers
docker-compose down
docker-compose up -d --build
```

### Option C: Using Portainer (Easiest)

1. **First, push your changes to GitHub:**
```bash
git add .
git commit -m "feat: Add RBAC user management UI to admin panel"
git push origin main
```

2. **In Portainer:**
   - Go to Stacks
   - Find your URLshorter stack
   - Click "Git Pull & Redeploy" (if you have Portainer Business)
   - OR manually:
     - SSH into server
     - `cd /path/to/URLshorter && git pull`
     - In Portainer: Stacks → URLshorter → "Update the stack" → Re-deploy

3. **Rebuild the frontend container:**
   - Go to Containers
   - Find your frontend container
   - Click "Recreate" → Check "Pull latest image" → Recreate

## Step 4: Verify Deployment

1. **Check CMS Access:**
   - Go to https://trunc.site/admin
   - Click on "Content" tab
   - Should load without 403 error

2. **Check User Management:**
   - Go to "Users" tab
   - Should see "Add User" button in top right
   - Click it - modal should open

3. **Test User Creation:**
   - Fill out the form
   - Select roles
   - Create user
   - Verify user appears in table

4. **Test Role Management:**
   - Click "View Roles" on any user
   - Modal should open showing roles
   - Try adding/removing roles

## Troubleshooting

### Frontend not updating?
```bash
# Clear Next.js cache and rebuild
cd frontend
rm -rf .next
npm run build

# Or force rebuild in Docker
docker-compose build --no-cache frontend
docker-compose up -d frontend
```

### Still getting 403 on CMS?
```bash
# Verify permissions were added
docker exec -it <postgres-container> psql -U urlshortener -d urlshortener_db

# Run this query:
SELECT r.name, p.name
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin' AND p.name LIKE 'cms.%';

# Should show cms.pages.read and other CMS permissions

# Clear Redis cache again
docker exec -it <redis-container> redis-cli FLUSHDB
```

### Components not found?
```bash
# Check if files exist on server
ls -la frontend/src/components/admin/
ls -la frontend/src/types/

# If missing, ensure git pull completed successfully
git status
git pull origin main
```

## Quick Command Reference

```bash
# Fix permissions
docker exec -i postgres_container psql -U urlshortener -d urlshortener_db < fix-cms-permissions.sql

# Clear Redis
docker exec -it redis_container redis-cli FLUSHDB

# Update code
git pull origin main

# Rebuild frontend
docker-compose up -d --build frontend

# View logs
docker-compose logs -f frontend
```
