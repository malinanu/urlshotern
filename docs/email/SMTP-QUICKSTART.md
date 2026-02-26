# SMTP Email Fix - Quick Start Guide

**Issue:** Password reset emails are being logged to console instead of sent via SMTP
**Time:** 15 minutes
**Difficulty:** Easy (configuration only, no code changes)

---

## TL;DR - 3 Steps to Fix

### 1. Generate Gmail App Password (5 minutes)

1. Go to: https://myaccount.google.com/apppasswords
2. Enable 2FA if not enabled: https://myaccount.google.com/security
3. Create app password: Select "Mail" → "Other (trunc.site SMTP)" → Generate
4. **Copy the 16-character password** (remove spaces): `xxxxxxxxxxxxxxxx`

### 2. Add SMTP Variables in Portainer (5 minutes)

1. Open Portainer → **Stacks** → Find your trunc.site stack
2. Click **"Editor"** tab
3. Find the `app` service `environment:` section
4. **Add these 6 lines** (replace values with yours):

```yaml
# SMTP Email Configuration
- SMTP_HOST=smtp.gmail.com
- SMTP_PORT=587
- SMTP_USERNAME=your.email@gmail.com
- SMTP_PASSWORD=xxxxxxxxxxxxxxxx
- FROM_EMAIL=noreply@trunc.site
- FROM_NAME=trunc.site
```

5. Click **"Update the stack"** button
6. Wait 1-2 minutes for restart

### 3. Verify & Test (5 minutes)

**SSH into server:**
```bash
ssh user@your-server

# Run verification script
cd /path/to/URLshorter
chmod +x scripts/verify-smtp-config.sh
./scripts/verify-smtp-config.sh
```

**Test in browser:**
1. Go to: https://trunc.site/forgot-password
2. Enter your email
3. Check inbox (and spam folder)
4. Click reset link and test

---

## Expected Results

### Before Fix
```
docker logs <container> | grep "EMAIL (CONSOLE)"
========== EMAIL (CONSOLE) ==========
To: user@example.com
Subject: Reset Your Password
...
```

### After Fix
```
# No console output
# Email arrives in inbox with:
From: trunc.site <noreply@trunc.site>
Subject: Reset Your Password - trunc.site
```

---

## Troubleshooting

### Still seeing console output?

**Check environment variables are set:**
```bash
docker exec <container-id> env | grep SMTP
```

**Should see:**
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your.email@gmail.com
SMTP_PASSWORD=xxxxxxxxxxxxxxxx
```

If missing, add them in Portainer and update stack.

---

### Getting "Authentication failed" error?

**Causes:**
- Using regular Gmail password instead of app password
- App password has spaces
- 2FA not enabled

**Fix:**
1. Generate new app password: https://myaccount.google.com/apppasswords
2. Copy without spaces
3. Update `SMTP_PASSWORD` in Portainer
4. Restart stack

---

### Emails going to spam?

**Short-term fix:**
```yaml
# Use Gmail address as FROM_EMAIL
- FROM_EMAIL=your.email@gmail.com
- FROM_NAME=trunc.site
```

**Long-term fix:**
- Use professional email service (SendGrid, Mailgun)
- Configure DNS records (SPF, DKIM, DMARC)

---

## Files Created

This implementation created:

- ✅ `docker/docker-compose.oracle.yml` - Production compose file with SMTP config
- ✅ `SMTP-CONFIGURATION-GUIDE.md` - Detailed configuration guide
- ✅ `SMTP-QUICKSTART.md` - This file (quick reference)
- ✅ `scripts/verify-smtp-config.sh` - Verification script (Linux)
- ✅ `scripts/verify-smtp-config.ps1` - Verification script (Windows)

---

## Need More Details?

See the comprehensive guide: **SMTP-CONFIGURATION-GUIDE.md**

Covers:
- Alternative SMTP providers (SendGrid, Mailgun, AWS SES)
- Detailed troubleshooting
- DNS configuration
- Security best practices
- Production considerations

---

## Support

**Common Commands:**

```bash
# Check logs
docker logs <container-id> --tail=100 | grep -i email

# Check environment variables
docker exec <container-id> env | grep SMTP

# Test SMTP connectivity
telnet smtp.gmail.com 587

# Restart container
docker restart <container-id>
```

**Still stuck?** Check logs for specific error messages:
```bash
docker logs <container-id> 2>&1 | grep "failed to create email provider"
```

The error message will tell you exactly what's wrong.
