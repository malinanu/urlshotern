# Troubleshooting Guide

This comprehensive guide helps you diagnose and resolve common issues with the URL Shortener application.

## Table of Contents

1. [Local Development Issues](#local-development-issues)
2. [Database Connection Problems](#database-connection-problems)
3. [Redis Connection Issues](#redis-connection-issues)
4. [Authentication & Authorization Errors](#authentication--authorization-errors)
5. [Frontend Build & Runtime Errors](#frontend-build--runtime-errors)
6. [API Client Issues](#api-client-issues)
7. [Performance Problems](#performance-problems)
8. [Deployment Issues](#deployment-issues)
9. [Quick Reference Commands](#quick-reference-commands)
10. [Getting Help](#getting-help)

---

## Local Development Issues

### Error: "Port 8080 already in use"

**Symptoms**: Cannot start backend server, error message about port already in use

**Cause**: Another process is using port 8080

**Solution**:

**Windows:**
```powershell
# Find process using port 8080
netstat -ano | findstr :8080

# Kill the process (replace <PID> with actual process ID)
taskkill /PID <PID> /F
```

**Linux/macOS:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process (replace <PID> with actual process ID)
kill -9 <PID>

# Or use one command
kill -9 $(lsof -t -i:8080)
```

**Alternative**: Change the server port in `.env`:
```env
SERVER_PORT=8081
```

---

### Error: "go: module not found"

**Symptoms**: Cannot run `go run cmd/server/main.go`, missing module errors

**Solution**:
```bash
# Download missing dependencies
go mod download

# Clean up and sync dependencies
go mod tidy

# If still failing, clear module cache
go clean -modcache
go mod download
```

**For specific module errors**:
```bash
# Update specific module
go get -u github.com/gin-gonic/gin

# Verify go.mod and go.sum are not corrupted
git status go.mod go.sum
```

---

### Error: "npm ERR! code ENOENT"

**Symptoms**: Frontend build fails with ENOENT errors

**Solution**:
```bash
cd frontend

# Clean install
rm -rf node_modules package-lock.json
npm install

# If using Windows and getting permission errors
npm cache clean --force
npm install
```

**For persistent issues**:
```bash
# Check Node.js version (should be 18.x or higher)
node --version

# Update npm
npm install -g npm@latest
```

---

### Error: "Cannot find module '@/components/...'"

**Symptoms**: Import errors in TypeScript/React code

**Solution**:

1. Verify `tsconfig.json` has correct paths:
```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

2. Restart TypeScript server:
- VS Code: `Cmd+Shift+P` → "TypeScript: Restart TS Server"
- Or restart your IDE

3. Clear Next.js cache:
```bash
rm -rf .next
npm run dev
```

---

## Database Connection Problems

### Error: "pq: password authentication failed"

**Symptoms**: Cannot connect to PostgreSQL

**Cause**: Incorrect database credentials

**Solution**:

1. Verify credentials in `.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_actual_password
DB_NAME=urlshortener
```

2. Test connection manually:
```bash
psql -h localhost -U postgres -d urlshortener
```

3. Reset PostgreSQL password if needed:
```sql
-- Connect as superuser
sudo -u postgres psql

-- Reset password
ALTER USER postgres PASSWORD 'newpassword';
```

4. Check `pg_hba.conf` authentication method:
```bash
# Linux
sudo nano /etc/postgresql/14/main/pg_hba.conf

# macOS (Homebrew)
nano /usr/local/var/postgres/pg_hba.conf

# Change method to 'md5' or 'scram-sha-256'
```

---

### Error: "database does not exist"

**Symptoms**: Application can't find database

**Solution**:

1. Create database:
```bash
createdb -U postgres urlshortener
```

2. Or via SQL:
```sql
psql -U postgres
CREATE DATABASE urlshortener;
\q
```

3. Run migrations:
```bash
psql -U postgres -d urlshortener -f configs/database.sql
```

4. Verify database exists:
```bash
psql -U postgres -l | grep urlshortener
```

---

### Error: "too many connections"

**Symptoms**: Application can't get DB connections, "sorry, too many clients already" error

**Cause**: Connection pool exhausted or PostgreSQL max_connections reached

**Solution**:

1. Check current connections:
```sql
SELECT count(*) FROM pg_stat_activity;
SELECT max_conn,used FROM (
  SELECT count(*) used FROM pg_stat_activity) t1,
  (SELECT setting::int max_conn FROM pg_settings WHERE name='max_connections') t2;
```

2. Increase PostgreSQL max_connections:
```bash
# Edit postgresql.conf
sudo nano /etc/postgresql/14/main/postgresql.conf

# Change:
max_connections = 200  # default is 100
```

3. Adjust application connection pool:
```env
# In .env
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

4. Find and kill idle connections:
```sql
-- View idle connections
SELECT pid, usename, state, query_start
FROM pg_stat_activity
WHERE state = 'idle'
AND query_start < now() - interval '5 minutes';

-- Kill idle connections
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
AND query_start < now() - interval '5 minutes';
```

---

### Error: "relation does not exist"

**Symptoms**: SQL errors about missing tables

**Cause**: Database migrations not run

**Solution**:
```bash
# Run database migrations
psql -U postgres -d urlshortener -f configs/database.sql

# Verify tables exist
psql -U postgres -d urlshortener -c "\dt"

# Check specific table
psql -U postgres -d urlshortener -c "\d users"
```

---

## Redis Connection Issues

### Error: "dial tcp 127.0.0.1:6379: connect: connection refused"

**Symptoms**: Cannot connect to Redis

**Cause**: Redis is not running

**Solution**:

**Check Redis status:**
```bash
# Test connection
redis-cli ping
# Should return: PONG
```

**Start Redis:**

**Windows:**
```powershell
# If installed via Chocolatey
redis-server

# Or start as service
sc start Redis
```

**Linux:**
```bash
# SystemD
sudo systemctl start redis
sudo systemctl enable redis  # Auto-start on boot

# Or direct
redis-server /etc/redis/redis.conf
```

**macOS:**
```bash
# Homebrew
brew services start redis

# Or direct
redis-server /usr/local/etc/redis.conf
```

**Docker:**
```bash
docker start redis
# Or if not exists
docker run -d --name redis -p 6379:6379 redis:alpine
```

---

### Error: "NOAUTH Authentication required"

**Symptoms**: Redis requires password but none provided

**Solution**:

1. Add password to `.env`:
```env
REDIS_PASSWORD=your_redis_password
```

2. If you don't want password authentication:
```bash
# Edit redis.conf
# Comment out or remove:
# requirepass yourpassword

# Restart Redis
sudo systemctl restart redis
```

---

### Error: "WRONGTYPE Operation against a key holding the wrong kind of value"

**Symptoms**: Redis type mismatch errors

**Cause**: Key exists with different data type than expected

**Solution**:
```bash
# Check key type
redis-cli TYPE url:abc123

# Delete problematic key
redis-cli DEL url:abc123

# Or flush entire Redis DB (WARNING: deletes all data)
redis-cli FLUSHDB
```

---

## Authentication & Authorization Errors

### Error: 401 Unauthorized

**Common Causes & Solutions**:

#### 1. Expired Token
**Symptoms**: Was working, now getting 401
**Solution**:
```javascript
// Frontend: Refresh token
const refreshToken = async () => {
  const response = await fetch('/api/v1/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: localStorage.getItem('refresh_token') })
  });
  const data = await response.json();
  localStorage.setItem('access_token', data.access_token);
};
```

#### 2. Invalid Token Format
**Symptoms**: Immediate 401, token looks malformed
**Solution**:
```javascript
// Verify Authorization header format
headers: {
  'Authorization': `Bearer ${token}`  // Note: "Bearer " prefix required
}
```

#### 3. Token Blacklisted (After Logout)
**Symptoms**: 401 with "token_revoked" error
**Solution**: User needs to log in again to get new token

---

### Error: 403 Forbidden

**Cause**: User lacks required permissions

**Solution**:

1. Check user role:
```sql
SELECT id, email, role, account_type, is_admin
FROM users
WHERE email = 'user@example.com';
```

2. Grant admin access:
```sql
UPDATE users SET is_admin = true WHERE email = 'user@example.com';
```

3. Check RBAC permissions:
```sql
SELECT * FROM user_roles WHERE user_id = 123;
SELECT * FROM role_permissions WHERE role_id = 1;
```

---

### Error: "Token refresh failed"

**Symptoms**: User logged out unexpectedly, can't refresh token

**Solution**:

1. Check refresh token expiry:
```sql
SELECT id, user_id, expires_at, created_at
FROM user_sessions
WHERE user_id = 123
ORDER BY created_at DESC;
```

2. Verify refresh endpoint:
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"your_refresh_token"}'
```

3. Check for blacklisted refresh tokens:
```bash
redis-cli GET blacklist:your_refresh_token
```

---

## Frontend Build & Runtime Errors

### Error: "Module not found: Can't resolve..."

**Solution**:
```bash
# Verify package is installed
npm list package-name

# Install if missing
npm install package-name

# Clear cache and rebuild
rm -rf node_modules .next
npm install
npm run build
```

---

### Error: "Hydration failed"

**Symptoms**: React hydration errors, content mismatch

**Cause**: Server-rendered HTML doesn't match client

**Solution**:

1. Check for browser-only code in component:
```typescript
// BAD
const UserProfile = () => {
  const data = localStorage.getItem('user'); // ❌ Breaks SSR
  return <div>{data}</div>;
};

// GOOD
const UserProfile = () => {
  const [data, setData] = useState(null);

  useEffect(() => {
    setData(localStorage.getItem('user')); // ✅ Only runs on client
  }, []);

  return <div>{data}</div>;
};
```

2. Use dynamic imports for client-only components:
```typescript
import dynamic from 'next/dynamic';

const ClientOnlyComponent = dynamic(
  () => import('@/components/ClientOnly'),
  { ssr: false }
);
```

---

### Error: "localStorage is not defined"

**Solution**:
```typescript
// Always check if running in browser
if (typeof window !== 'undefined') {
  localStorage.setItem('key', 'value');
}

// Or use try-catch
try {
  localStorage.setItem('key', 'value');
} catch (e) {
  console.error('localStorage not available:', e);
}
```

---

### Build Error: "Module parse failed"

**Solution**:
```bash
# Clear Next.js cache
rm -rf .next

# Check for syntax errors
npm run lint

# Rebuild
npm run build
```

---

## API Client Issues

### Error: "Network request failed"

**Symptoms**: All API calls fail with fetch errors

**Cause**: Backend not running or CORS issues

**Solution**:

1. Verify backend is running:
```bash
curl http://localhost:8080/health
```

2. Check CORS configuration in `cmd/server/main.go`:
```go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

3. Verify API URL in `.env.local`:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

### Error: "Failed to fetch" in production

**Cause**: Using localhost URL in production

**Solution**:

1. Set production API URL:
```env
# Production .env
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

2. Verify environment variable is loaded:
```javascript
console.log('API URL:', process.env.NEXT_PUBLIC_API_URL);
```

---

### Error: "CORS policy: No 'Access-Control-Allow-Origin' header"

**Solution**:

1. Add frontend domain to backend CORS whitelist
2. For development, allow localhost:
```go
AllowOrigins: []string{
    "http://localhost:3000",
    "https://yourdomain.com",
},
```

---

## Performance Problems

### Slow Page Load Times

**Diagnostics**:
```bash
# Run Lighthouse audit
npm run build
npm start
# Open Chrome DevTools > Lighthouse > Analyze page load

# Check bundle size
npm run build
# Look for large chunks in output
```

**Solutions**:

1. **Enable Redis caching**:
```env
REDIS_HOST=localhost
REDIS_PORT=6379
```

2. **Optimize images**:
```typescript
// Use Next.js Image component
import Image from 'next/image';

<Image
  src="/logo.png"
  width={500}
  height={300}
  alt="Logo"
/>
```

3. **Implement code splitting**:
```typescript
const HeavyComponent = dynamic(() => import('./HeavyComponent'));
```

4. **Add database indexes**:
```sql
CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE INDEX idx_urls_short_code ON urls(short_code);
CREATE INDEX idx_clicks_url_id ON clicks(url_id);
```

---

### High Memory Usage

**Diagnostics**:
```bash
# Check Go memory profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Check current memory usage
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof -http=:8081 heap.prof
```

**Solutions**:

1. **Implement connection pooling**:
```env
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

2. **Fix memory leaks**:
- Close database connections properly
- Use `defer rows.Close()` after queries
- Don't store large datasets in memory

3. **Optimize query result sets**:
```sql
-- Instead of SELECT *
SELECT id, short_code, original_url FROM urls LIMIT 100;
```

---

## Deployment Issues

### Docker Build Fails

**Common Errors**:

**Error: "Cannot find module"**
```dockerfile
# Ensure node_modules is not in .dockerignore
# Copy package files first
COPY package*.json ./
RUN npm ci --only=production
COPY . .
```

**Error: "Permission denied"**
```bash
# Check file permissions
chmod +x scripts/*.sh

# Or in Dockerfile
RUN chmod +x /app/scripts/*.sh
```

---

### Kubernetes Pod CrashLoopBackOff

**Diagnostics**:
```bash
# Check pod status
kubectl describe pod <pod-name>

# View logs
kubectl logs <pod-name>

# View previous logs if pod restarted
kubectl logs <pod-name> --previous
```

**Common Causes**:

1. **Missing environment variables**:
```yaml
# Verify ConfigMap/Secret exists
kubectl get configmap
kubectl get secret

# Check if mounted correctly
kubectl exec <pod-name> -- env | grep DB_
```

2. **Database connection failures**:
```bash
# Test from pod
kubectl exec <pod-name> -- nc -zv postgres-service 5432
```

3. **Invalid secrets**:
```bash
# Decode and verify secret
kubectl get secret db-credentials -o jsonpath='{.data.password}' | base64 -d
```

---

### SSL Certificate Issues

**Error**: "certificate has expired"

**Solution**:
```bash
# Check certificate expiration
openssl x509 -in /etc/nginx/ssl/cert.pem -noout -dates

# Renew Let's Encrypt certificate
sudo certbot renew

# Reload nginx
sudo systemctl reload nginx
```

**Error**: "certificate verify failed"

**Solution**:
```bash
# Verify certificate chain
openssl verify -CAfile ca-bundle.crt server.crt

# Check certificate matches key
openssl x509 -noout -modulus -in server.crt | openssl md5
openssl rsa -noout -modulus -in server.key | openssl md5
# MD5 hashes should match
```

---

## Quick Reference Commands

### Health Checks
```bash
# Backend health
curl http://localhost:8080/health

# Database connectivity
psql -U postgres -d urlshortener -c "SELECT 1"

# Redis connectivity
redis-cli ping

# Frontend dev server
curl http://localhost:3000
```

### Logs
```bash
# Backend logs (if using systemd)
journalctl -u urlshortener -f

# Database logs
tail -f /var/log/postgresql/postgresql-14-main.log

# Nginx logs
tail -f /var/log/nginx/error.log
tail -f /var/log/nginx/access.log

# Docker logs
docker logs -f urlshortener-backend
docker logs -f urlshortener-frontend
```

### Database Maintenance
```bash
# Vacuum database
psql -U postgres -d urlshortener -c "VACUUM ANALYZE;"

# Check database size
psql -U postgres -d urlshortener -c "SELECT pg_size_pretty(pg_database_size('urlshortener'));"

# List large tables
psql -U postgres -d urlshortener -c "
SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::regclass))
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(tablename::regclass) DESC;
"
```

### Redis Maintenance
```bash
# Check memory usage
redis-cli INFO memory

# View all keys (WARNING: slow on large datasets)
redis-cli KEYS "*"

# Count keys
redis-cli DBSIZE

# Clear specific pattern
redis-cli --scan --pattern 'url:*' | xargs redis-cli DEL
```

### Reset Everything (Nuclear Option)
```bash
# Stop all services
docker-compose down -v

# Clean databases
dropdb urlshortener
createdb urlshortener
psql -U postgres -d urlshortener -f configs/database.sql
redis-cli FLUSHALL

# Clean frontend
cd frontend
rm -rf node_modules .next package-lock.json
npm install

# Clean backend
cd ..
go clean -modcache
go mod download

# Restart
docker-compose up -d
# Or
go run cmd/server/main.go
cd frontend && npm run dev
```

---

## Getting Help

If you've tried these solutions and still have issues:

### Before Asking for Help

1. **Enable debug logging**:
```env
LOG_LEVEL=debug
```

2. **Collect diagnostic info**:
- Error messages and full stack traces
- Environment details (OS, Go version, Node version)
- Configuration (sanitize sensitive data like passwords)
- Steps to reproduce the issue
- What you've already tried

3. **Check existing issues**:
- GitHub Issues: https://github.com/yourusername/URLshorter/issues
- Search for similar problems

### Create a GitHub Issue

Include the following template:

```markdown
## Issue Description
[Clear description of the problem]

## Environment
- OS: [e.g., Ubuntu 22.04, macOS 13, Windows 11]
- Go Version: [run `go version`]
- Node Version: [run `node --version`]
- Database: [PostgreSQL version]
- Deployment: [Local, Docker, Kubernetes]

## Steps to Reproduce
1.
2.
3.

## Expected Behavior
[What should happen]

## Actual Behavior
[What actually happens]

## Error Messages
```
[paste error messages here]
```

## Configuration
[sanitized .env or relevant config]

## Logs
[relevant log entries]

## What I've Tried
- [ ] Checked this troubleshooting guide
- [ ] Verified database connection
- [ ] Verified Redis connection
- [ ] Cleared caches
- [ ] Restarted services
```

### Community Support

- GitHub Discussions
- Stack Overflow (tag: `url-shortener`)
- Discord/Slack community (if available)

---

## Additional Resources

- [Installation Guide](../README.md#installation)
- [Configuration Reference](./CONFIGURATION.md)
- [Secret Rotation Guide](./SECRETS_ROTATION.md)
- [API Documentation](./API_DOCUMENTATION.md)
- [Deployment Guides](./DEPLOYMENT_GUIDES/)
