# 🚀 Deployment Ready - Password Reset Pages Fix

**Status:** Ready for Production Deployment
**Date:** February 15, 2026
**Issue:** 404 errors for /forgot-password and /reset-password in production
**Solution:** Rebuild Docker container to include new static pages

---

## Quick Start

You have **3 files** to help you deploy:

1. **`DEPLOYMENT-CHECKLIST.md`** - Simple checklist to follow step-by-step
2. **`deployment-guide.md`** - Detailed guide with all commands and troubleshooting
3. **`scripts/verify-deployment.sh`** - Automated verification script (run on server)

---

## What's The Problem?

The `/forgot-password` and `/reset-password` pages exist in your GitHub repo but return 404 in production because:

- These pages were added **after** the current Docker container was built
- Next.js uses static export → pages must be built into HTML files
- The production container doesn't have these HTML files in `/app/static/`

**Solution:** Rebuild the container to generate the new static files.

---

## How to Fix It (2 Options)

### Option 1: Portainer UI (Easiest - 5 minutes)

1. Login to Portainer
2. Go to **Stacks** → **trunc.site**
3. Click **Editor**
4. Enable **"☑ Re-pull image and redeploy"**
5. Click **"Update the stack"**
6. Wait for green status
7. Run verification (see below)

### Option 2: SSH Command Line (10 minutes)

```bash
# Connect to server
ssh user@your-oracle-server

# Navigate to project
cd /path/to/URLshorter

# Pull latest code
git pull origin main

# Rebuild containers
docker-compose -f docker/docker-compose.oracle.yml down
docker-compose -f docker/docker-compose.oracle.yml build --no-cache
docker-compose -f docker/docker-compose.oracle.yml up -d

# Verify
bash scripts/verify-deployment.sh
```

---

## Verify Deployment

### Automated Verification (Recommended)

```bash
# Run the verification script on the server
bash scripts/verify-deployment.sh
```

This checks:
- ✓ Containers running
- ✓ Static files exist
- ✓ Pages return 200 status
- ✓ API endpoints working
- ✓ No errors in logs

### Manual Verification

**In Browser:**
1. https://trunc.site/forgot-password (should load without 404)
2. https://trunc.site/reset-password?token=test (should load without 404)

**In Terminal:**
```bash
# Check static files exist
CONTAINER_ID=$(docker-compose -f docker/docker-compose.oracle.yml ps -q app)
docker exec -it $CONTAINER_ID ls -la /app/static/forgot-password.html
docker exec -it $CONTAINER_ID ls -la /app/static/reset-password.html
```

---

## Test End-to-End Flow

1. **Request reset:** Go to https://trunc.site/forgot-password
2. **Enter email:** Use a real email you can access
3. **Check inbox:** Verify reset email arrives
4. **Click link:** Should open reset-password page with token
5. **Reset password:** Enter new password and submit
6. **Login:** Test login with new password

---

## Files Changed (Already Pushed to GitHub ✓)

- ✓ `frontend/src/app/forgot-password/page.tsx`
- ✓ `frontend/src/app/reset-password/page.tsx`
- ✓ `internal/handlers/auth.go`
- ✓ `internal/services/auth.go`
- ✓ `internal/storage/postgres.go`

**No code changes needed** - just rebuild the container!

---

## If Something Goes Wrong

### Rollback to Previous Version

```bash
# Stop containers
docker-compose -f docker/docker-compose.oracle.yml down

# View previous images
docker images | grep trunc

# Rollback (replace with actual image ID)
docker tag <PREVIOUS_IMAGE_ID> <CURRENT_IMAGE_TAG>
docker-compose -f docker/docker-compose.oracle.yml up -d
```

### Common Issues

**Issue:** Build fails
**Fix:** Check build logs for specific error, ensure all dependencies available

**Issue:** Pages still show 404
**Fix:** Verify static files exist in container (see verification above)

**Issue:** Email not sending
**Fix:** Check SMTP environment variables are set correctly

---

## Environment Variables Required

Make sure these are set in your production `.env` or docker-compose:

```env
# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
SMTP_FROM=noreply@trunc.site

# Application URLs
BASE_URL=https://trunc.site
FRONTEND_URL=https://trunc.site
```

---

## Expected Timeline

- **Deployment:** 5-10 minutes
- **Verification:** 5 minutes
- **Testing:** 10 minutes
- **Total:** ~20-25 minutes

---

## Success Checklist

After deployment, all these should pass:

- [ ] Container rebuilt successfully
- [ ] forgot-password.html exists in /app/static/
- [ ] reset-password.html exists in /app/static/
- [ ] https://trunc.site/forgot-password returns 200
- [ ] https://trunc.site/reset-password returns 200
- [ ] Password reset email sends successfully
- [ ] User can reset password and login
- [ ] No errors in container logs

---

## Next Steps

**Ready to deploy?**

1. **Choose your method:** Portainer UI (easier) or SSH (more control)
2. **Follow the checklist:** Open `DEPLOYMENT-CHECKLIST.md`
3. **Deploy:** Execute the steps
4. **Verify:** Run `scripts/verify-deployment.sh`
5. **Test:** Complete end-to-end password reset flow
6. **Monitor:** Watch logs for 15 minutes

**Need more details?**

- See `deployment-guide.md` for comprehensive instructions
- See `DEPLOYMENT-CHECKLIST.md` for step-by-step checklist

---

## Questions?

- **What if I don't have SSH access?** Use Portainer UI method
- **Will there be downtime?** Minimal (1-2 minutes during rebuild)
- **Can I test before production?** Yes, test locally with `docker-compose -f docker/docker-compose.local.yml up --build`
- **What if it fails?** Follow rollback instructions above

---

**You're ready to deploy!** 🎉

Choose your deployment method and follow the checklist. The verification script will confirm everything is working correctly.

Good luck! 🚀
