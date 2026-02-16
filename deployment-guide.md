# Production Deployment Guide - Fix Password Reset Pages 404

**Date:** February 15, 2026
**Issue:** /forgot-password and /reset-password return 404 in production
**Solution:** Rebuild Docker container to include new static pages
**Production URL:** https://trunc.site

---

## Pre-Deployment Checklist

- [ ] Latest code pushed to GitHub main branch (✓ Already done)
- [ ] Local testing completed successfully
- [ ] Backup plan ready (rollback commands below)
- [ ] Access to Portainer or SSH credentials available
- [ ] Estimated downtime window: 5-10 minutes

---

## Option A: Deploy via Portainer UI (Recommended)

### Step 1: Access Portainer
1. Navigate to your Portainer instance
2. Login with your credentials
3. Select the appropriate environment (Oracle Ubuntu server)

### Step 2: Update Stack
1. Go to **Stacks** in the left sidebar
2. Locate the **trunc.site** stack (or URLshorter stack)
3. Click on the stack name to open it
4. Click **Editor** button at the top
5. Verify the stack is using `docker-compose.oracle.yml`

### Step 3: Rebuild and Deploy
1. Scroll to the bottom of the editor
2. Enable checkbox: **☑ Re-pull image and redeploy**
3. Click **Update the stack** button
4. Wait for deployment to complete (watch the logs)
5. Verify status shows "running" with green indicator

### Step 4: Verify Deployment (See Section Below)

---

## Option B: Deploy via SSH Command Line

### Step 1: Connect to Server
```bash
# Replace with your actual server address and credentials
ssh user@your-oracle-server-ip
```

### Step 2: Navigate to Project Directory
```bash
# Adjust path to match your actual installation
cd /path/to/URLshorter

# Example common paths:
# cd ~/URLshorter
# cd /opt/URLshorter
# cd /home/ubuntu/URLshorter
```

### Step 3: Pull Latest Code
```bash
# Ensure we're on main branch
git branch

# Pull latest changes
git pull origin main

# Verify the new pages exist in repo
ls -la frontend/src/app/forgot-password/
ls -la frontend/src/app/reset-password/
```

### Step 4: Rebuild and Restart Containers
```bash
# Stop current containers
docker-compose -f docker/docker-compose.oracle.yml down

# Rebuild with latest code (includes frontend static build)
docker-compose -f docker/docker-compose.oracle.yml build --no-cache

# Start containers in detached mode
docker-compose -f docker/docker-compose.oracle.yml up -d

# Watch logs for any errors
docker-compose -f docker/docker-compose.oracle.yml logs -f --tail=100
# Press Ctrl+C to stop watching logs
```

### Step 5: Verify Containers are Running
```bash
# Check container status
docker-compose -f docker/docker-compose.oracle.yml ps

# Expected output: All containers should show "Up" status
```

---

## Verification Steps (Required After Either Method)

### 1. Check Static Files in Container

```bash
# Get the container ID for the main app
docker ps | grep trunc

# Or get it via docker-compose
CONTAINER_ID=$(docker-compose -f docker/docker-compose.oracle.yml ps -q app)

# Verify forgot-password static file exists
docker exec -it $CONTAINER_ID ls -la /app/static/forgot-password.html

# Verify reset-password static file exists
docker exec -it $CONTAINER_ID ls -la /app/static/reset-password.html

# Check Next.js static assets
docker exec -it $CONTAINER_ID ls -la /app/static/_next/static/
```

**Expected:** Both HTML files should exist with recent timestamps

### 2. Test in Browser

#### Test Forgot Password Page
1. Open browser (incognito mode recommended)
2. Navigate to: https://trunc.site/forgot-password
3. **Verify:** Page loads without 404 error
4. **Verify:** Form displays with email input field
5. **Verify:** Submit button is visible

#### Test Reset Password Page
1. Navigate to: https://trunc.site/reset-password?token=test123
2. **Verify:** Page loads without 404 error
3. **Verify:** Password reset form displays
4. **Verify:** New password and confirm password fields visible

### 3. Check Application Logs

```bash
# View recent logs
docker-compose -f docker/docker-compose.oracle.yml logs --tail=100

# Watch logs in real-time (Ctrl+C to stop)
docker-compose -f docker/docker-compose.oracle.yml logs -f

# Check for specific errors
docker-compose -f docker/docker-compose.oracle.yml logs | grep -i error
docker-compose -f docker/docker-compose.oracle.yml logs | grep -i "404"
```

**Look for:**
- ✓ No 404 errors for /forgot-password or /reset-password
- ✓ Server started successfully
- ✓ Database connection successful
- ✗ Any error messages or stack traces

### 4. Test Complete Password Reset Flow

#### A. Request Password Reset
```bash
# Test API endpoint
curl -X POST https://trunc.site/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "your-test-email@example.com"}'

# Expected response (200 OK):
# {"message":"If an account exists with that email, a password reset link has been sent"}
```

#### B. Check Email Delivery
1. Check inbox for password reset email
2. Verify email subject: "Reset Your Password - Trunc.site"
3. Verify email contains reset link
4. Copy the token from the link

#### C. Reset Password (Browser)
1. Click reset link from email or navigate to:
   ```
   https://trunc.site/reset-password?token=<TOKEN_FROM_EMAIL>
   ```
2. Enter new password (min 8 characters)
3. Confirm password
4. Click "Reset Password" button
5. **Verify:** Success message appears
6. **Verify:** Redirected to login page

#### D. Test New Password
1. Login with email and new password
2. **Verify:** Successfully authenticated
3. **Verify:** Redirected to dashboard/home

---

## Troubleshooting

### Issue: Static Files Not Found After Rebuild

**Check build logs:**
```bash
docker-compose -f docker/docker-compose.oracle.yml logs app | grep "Building"
docker-compose -f docker/docker-compose.oracle.yml logs app | grep "next build"
```

**Rebuild with verbose output:**
```bash
docker-compose -f docker/docker-compose.oracle.yml build --no-cache --progress=plain app
```

### Issue: Container Won't Start

**Check container logs:**
```bash
docker-compose -f docker/docker-compose.oracle.yml logs app
```

**Common issues:**
- Port already in use
- Environment variables missing
- Database not accessible

**Restart individual service:**
```bash
docker-compose -f docker/docker-compose.oracle.yml restart app
```

### Issue: Email Not Sending

**Check SMTP configuration:**
```bash
# Verify environment variables in container
docker-compose -f docker/docker-compose.oracle.yml exec app env | grep SMTP
```

**Required environment variables:**
- SMTP_HOST
- SMTP_PORT
- SMTP_USER
- SMTP_PASS
- SMTP_FROM

**Test SMTP connection:**
```bash
# Check application logs for SMTP errors
docker-compose -f docker/docker-compose.oracle.yml logs app | grep -i smtp
```

### Issue: Pages Still Show 404

**Verify Next.js build output:**
```bash
# Check if static export completed
docker exec -it $CONTAINER_ID ls -la /app/static/

# Check Next.js output directory
docker exec -it $CONTAINER_ID cat /app/frontend/out/forgot-password.html
```

**Check Go static file serving:**
```bash
# Verify static directory is correct
docker-compose -f docker/docker-compose.oracle.yml exec app ls -la ./static/
```

---

## Rollback Plan

If issues occur and you need to rollback:

### Method 1: Rollback via Portainer
1. Go to Images in Portainer
2. Find previous working image (check timestamp)
3. Go to Stacks → your stack → Editor
4. Change image tag to previous version
5. Click "Update the stack"

### Method 2: Rollback via Command Line

```bash
# View available images
docker images | grep trunc

# Stop current deployment
docker-compose -f docker/docker-compose.oracle.yml down

# Tag previous working image
docker tag <PREVIOUS_IMAGE_ID> <CURRENT_IMAGE_TAG>

# Or pull specific version from registry
# docker pull your-registry/trunc:previous-tag

# Start with previous image
docker-compose -f docker/docker-compose.oracle.yml up -d

# Verify rollback successful
docker-compose -f docker/docker-compose.oracle.yml ps
```

### Emergency Stop All Services

```bash
docker-compose -f docker/docker-compose.oracle.yml down --remove-orphans
```

---

## Environment Variables Checklist

Ensure these are set in your `.env` or docker-compose environment:

**Email Configuration:**
```env
SMTP_HOST=smtp.gmail.com          # or your SMTP server
SMTP_PORT=587                      # or 465 for SSL
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
SMTP_FROM=noreply@trunc.site
```

**Application URLs:**
```env
BASE_URL=https://trunc.site
FRONTEND_URL=https://trunc.site
```

**Database:**
```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=urlshortener
DB_PASSWORD=your-secure-password
DB_NAME=urlshortener
```

---

## Success Criteria Checklist

- [ ] Docker container rebuilt successfully
- [ ] No build errors in logs
- [ ] `/app/static/forgot-password.html` exists in container
- [ ] `/app/static/reset-password.html` exists in container
- [ ] https://trunc.site/forgot-password returns 200 status
- [ ] https://trunc.site/reset-password returns 200 status
- [ ] Forgot password form displays correctly
- [ ] Password reset email sent successfully
- [ ] Password reset form displays correctly
- [ ] User can reset password and login with new password
- [ ] No errors in container logs
- [ ] All containers showing "Up" status
- [ ] Application accessible at https://trunc.site

---

## Post-Deployment

### Monitor for 24 Hours
```bash
# Check error logs periodically
docker-compose -f docker/docker-compose.oracle.yml logs --since 1h | grep -i error

# Monitor container health
docker-compose -f docker/docker-compose.oracle.yml ps
```

### Update Documentation
- [ ] Update MEMORY.md with deployment date and outcome
- [ ] Document any issues encountered and solutions
- [ ] Note any configuration changes made

---

## Support Commands Reference

```bash
# Container Management
docker ps                                                          # List running containers
docker-compose -f docker/docker-compose.oracle.yml ps             # List stack containers
docker-compose -f docker/docker-compose.oracle.yml restart app    # Restart app container
docker-compose -f docker/docker-compose.oracle.yml logs -f app    # Follow app logs

# Execute Commands in Container
docker-compose -f docker/docker-compose.oracle.yml exec app sh    # Open shell in container
docker-compose -f docker/docker-compose.oracle.yml exec app env   # View environment variables

# Resource Cleanup
docker system prune                                                # Clean unused resources
docker image prune                                                 # Remove unused images
docker volume prune                                                # Remove unused volumes

# Health Checks
curl -i https://trunc.site/health                                 # Check API health
curl -i https://trunc.site/health/db-pool                         # Check DB connections
```

---

## Estimated Timeline

1. **SSH/Portainer Access:** 2 minutes
2. **Pull latest code:** 1 minute
3. **Docker rebuild:** 5-10 minutes (depends on server specs)
4. **Verification:** 5 minutes
5. **End-to-end testing:** 10 minutes

**Total Time: 20-30 minutes**

---

## Contact & Escalation

If issues persist after deployment:
1. Check container logs for specific errors
2. Verify all environment variables are set
3. Test email configuration separately
4. Verify static build completed successfully
5. Consider rollback if critical functionality affected

---

**Deployment prepared by:** Claude Code
**Last updated:** February 15, 2026
**Status:** Ready for execution
