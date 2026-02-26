# Production Deployment Checklist
## Fix Password Reset Pages 404 - February 15, 2026

---

## Pre-Deployment

- [ ] Latest code confirmed on GitHub main branch
- [ ] Portainer or SSH access ready
- [ ] Backup/rollback plan reviewed
- [ ] Team notified of deployment window

---

## Deployment (Choose One Method)

### Method A: Portainer UI
- [ ] Navigate to Portainer
- [ ] Go to Stacks → trunc.site
- [ ] Click "Editor"
- [ ] Enable "☑ Re-pull image and redeploy"
- [ ] Click "Update the stack"
- [ ] Wait for completion (green status)

### Method B: SSH Command Line
- [ ] SSH into Oracle server
- [ ] Navigate to project directory: `cd /path/to/URLshorter`
- [ ] Pull latest: `git pull origin main`
- [ ] Stop containers: `docker-compose -f docker/docker-compose.oracle.yml down`
- [ ] Rebuild: `docker-compose -f docker/docker-compose.oracle.yml build --no-cache`
- [ ] Start: `docker-compose -f docker/docker-compose.oracle.yml up -d`
- [ ] Verify: `docker-compose -f docker/docker-compose.oracle.yml ps`

---

## Verification

### 1. Container Status
```bash
docker-compose -f docker/docker-compose.oracle.yml ps
```
- [ ] All containers show "Up" status

### 2. Static Files Check
```bash
# Run verification script (recommended)
bash scripts/verify-deployment.sh

# Or check manually
CONTAINER_ID=$(docker-compose -f docker/docker-compose.oracle.yml ps -q app)
docker exec -it $CONTAINER_ID ls -la /app/static/forgot-password.html
docker exec -it $CONTAINER_ID ls -la /app/static/reset-password.html
```
- [ ] forgot-password.html exists
- [ ] reset-password.html exists

### 3. Browser Testing
- [ ] Open https://trunc.site/forgot-password
- [ ] Page loads without 404
- [ ] Form displays correctly
- [ ] Open https://trunc.site/reset-password?token=test
- [ ] Page loads without 404
- [ ] Password form displays correctly

### 4. API Testing
```bash
curl -X POST https://trunc.site/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```
- [ ] API returns 200 status
- [ ] No errors in response

### 5. Log Check
```bash
docker-compose -f docker/docker-compose.oracle.yml logs --tail=50
```
- [ ] No 404 errors for /forgot-password
- [ ] No 404 errors for /reset-password
- [ ] No critical errors

---

## End-to-End Testing

### 6. Request Password Reset
- [ ] Go to https://trunc.site/forgot-password
- [ ] Enter valid email address
- [ ] Submit form
- [ ] See success message

### 7. Check Email
- [ ] Password reset email received
- [ ] Email contains reset link
- [ ] Link format correct: https://trunc.site/reset-password?token=...

### 8. Reset Password
- [ ] Click reset link from email
- [ ] Reset password page loads
- [ ] Enter new password (min 8 chars)
- [ ] Confirm password
- [ ] Submit form
- [ ] See success message
- [ ] Redirected to login

### 9. Test Login
- [ ] Login with new password
- [ ] Successfully authenticated
- [ ] Access granted to dashboard

---

## Post-Deployment

### 10. Monitor
- [ ] Check logs for 15 minutes: `docker-compose -f docker/docker-compose.oracle.yml logs -f`
- [ ] No errors appearing
- [ ] Application stable

### 11. Documentation
- [ ] Update MEMORY.md with deployment results
- [ ] Note any issues encountered
- [ ] Document any config changes

---

## Rollback (If Needed)

If critical issues occur:

```bash
# Stop current deployment
docker-compose -f docker/docker-compose.oracle.yml down

# View previous images
docker images | grep trunc

# Rollback to previous image
docker tag <PREVIOUS_IMAGE_ID> <CURRENT_IMAGE_TAG>
docker-compose -f docker/docker-compose.oracle.yml up -d
```

- [ ] Rollback executed successfully
- [ ] Service restored to previous state
- [ ] Incident documented

---

## Success Criteria (All Must Pass)

- [ ] Docker rebuild completed without errors
- [ ] Static HTML files present in container
- [ ] Forgot password page returns 200
- [ ] Reset password page returns 200
- [ ] Forms display correctly
- [ ] API endpoints working
- [ ] Email delivery successful
- [ ] End-to-end password reset works
- [ ] User can login with new password
- [ ] No errors in logs
- [ ] All containers healthy

---

## Quick Commands Reference

```bash
# Container status
docker-compose -f docker/docker-compose.oracle.yml ps

# View logs
docker-compose -f docker/docker-compose.oracle.yml logs -f app

# Restart app
docker-compose -f docker/docker-compose.oracle.yml restart app

# Check static files
CONTAINER_ID=$(docker-compose -f docker/docker-compose.oracle.yml ps -q app)
docker exec -it $CONTAINER_ID ls -la /app/static/

# Test endpoints
curl -i https://trunc.site/forgot-password
curl -i https://trunc.site/reset-password
curl -i https://trunc.site/health

# Run verification script
bash scripts/verify-deployment.sh
```

---

**Deployment Date:** _______________
**Deployed By:** _______________
**Start Time:** _______________
**End Time:** _______________
**Status:** [ ] Success  [ ] Failed  [ ] Rolled Back
**Notes:**

_______________________________________________

_______________________________________________

_______________________________________________
