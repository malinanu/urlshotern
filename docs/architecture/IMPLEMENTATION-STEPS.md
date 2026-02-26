# Password Reset Email Fix - Implementation Steps

**Status:** Ready to implement on production server
**Time Required:** 15-20 minutes
**Risk Level:** LOW (configuration only, no code changes)

---

## Quick Summary

**Problem:** Password reset emails are logged to console instead of being sent via SMTP.

**Root Cause:** SMTP environment variables are not configured on the production server.

**Solution:** Add SMTP configuration to Portainer and restart the stack.

**Evidence:**
```
========== EMAIL (CONSOLE) ==========
From: URL Shortener <noreply@urlshortener.com>
To: mjjayaweera@gmail.com
Subject: Reset Your Password - trunc.site
```

---

## Prerequisites

Before starting, you'll need:

1. **Access to Portainer** on your Oracle server
2. **Gmail account** with 2FA enabled (or other SMTP provider)
3. **Gmail app password** (16 characters)

---

## Step 1: Generate Gmail App Password (5 minutes)

### Option A: Gmail (Recommended for Quick Fix)

1. **Enable 2FA** (if not already):
   - Visit: https://myaccount.google.com/security
   - Enable "2-Step Verification"

2. **Generate App Password**:
   - Visit: https://myaccount.google.com/apppasswords
   - Select: "Mail" → "Other" → Enter "trunc.site SMTP"
   - Click "Generate"
   - **Copy the 16-character password** (remove spaces)
   - Example: `abcdefghijklmnop` (no spaces)

### Option B: Other SMTP Provider

See `SMTP-CONFIGURATION-GUIDE.md` for:
- SendGrid (free 100 emails/day)
- Mailgun (free 5,000 emails/month)
- AWS SES (very cheap)

---

## Step 2: Configure SMTP in Portainer (5 minutes)

### Access Portainer

1. **Open Portainer** in your browser
2. **Navigate to:** Stacks → Find your trunc.site stack
3. **Click:** Editor tab

### Add SMTP Environment Variables

Find the `app` service → `environment:` section in the docker-compose file:

```yaml
services:
  app:
    environment:
      - ENVIRONMENT=production
      - DB_HOST=postgres
      # ... existing variables ...
```

**Add these 6 lines** (replace with your actual values):

```yaml
      # SMTP Email Configuration
      - SMTP_HOST=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USERNAME=your.email@gmail.com
      - SMTP_PASSWORD=abcdefghijklmnop
      - FROM_EMAIL=noreply@trunc.site
      - FROM_NAME=trunc.site
```

### Update Stack

1. **Click:** "Update the stack" button (bottom of page)
2. **Wait:** 1-2 minutes for container to restart
3. **Verify:** Container is running (green status)

---

## Step 3: Verify Configuration (5 minutes)

### Check Email Health Endpoint

**Open in browser:**
```
https://trunc.site/health/email
```

**Expected response (SUCCESS):**
```json
{
  "status": "ok",
  "provider": "smtp",
  "environment": "production",
  "smtp_config": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your.email@gmail.com",
    "password": "ab****op",
    "from_email": "noreply@trunc.site",
    "from_name": "trunc.site"
  }
}
```

**If you see `"provider": "console"`, the SMTP variables were not loaded correctly.**

### SSH Verification (Optional)

If the health endpoint shows `console` provider, SSH into the server:

```bash
# SSH into Oracle server
ssh user@your-server-ip

# Find container ID
docker ps | grep trunc

# Check environment variables
docker exec <container-id> env | grep SMTP

# Expected output:
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USERNAME=your.email@gmail.com
# SMTP_PASSWORD=abcdefghijklmnop
```

**If SMTP variables are empty, they weren't added correctly in Portainer.**

### Check Logs

```bash
# View recent logs
docker logs <container-id> --tail=50 | grep -i email

# Expected (SUCCESS):
# INFO [EMAIL] Initializing email service (environment=production, provider=smtp)
# INFO [EMAIL] SMTP configuration: host=smtp.gmail.com, port=587

# NOT expected (FAILURE):
# ========== EMAIL (CONSOLE) ==========
```

---

## Step 4: Test Password Reset (5 minutes)

### Request Password Reset

1. **Open browser:** https://trunc.site/forgot-password
2. **Enter email:** Your email address
3. **Click:** "Send Reset Link"

### Watch Logs in Real-Time

```bash
# SSH into server
docker logs -f <container-id>

# You should see:
# INFO [EMAIL] Sending password reset email to your@email.com
# INFO [EMAIL] Password reset email sent successfully

# You should NOT see:
# ========== EMAIL (CONSOLE) ==========
```

### Check Email Inbox

1. **Wait:** 1-2 minutes
2. **Check inbox** (and spam folder)
3. **Look for email:**
   - From: `trunc.site <noreply@trunc.site>`
   - Subject: "Reset Your Password - trunc.site"
4. **Click reset link:** Should open `https://trunc.site/reset-password?token=...`
5. **Enter new password:** Submit and verify success

---

## Expected Results

### Before Fix
```bash
docker logs <container-id> | grep "EMAIL (CONSOLE)"
# Output:
# ========== EMAIL (CONSOLE) ==========
# From: URL Shortener <noreply@urlshortener.com>
# To: user@example.com
# Subject: Reset Your Password - trunc.site
```

### After Fix
```bash
docker logs <container-id> | grep "EMAIL (CONSOLE)"
# No output (emails sent via SMTP, not logged to console)

# Check for successful sends:
docker logs <container-id> | grep "email sent successfully"
# INFO [EMAIL] Password reset email sent successfully to user@example.com
```

---

## Troubleshooting

### Issue: Still Seeing "EMAIL (CONSOLE)"

**Problem:** SMTP variables not loaded

**Solution:**
1. Check Portainer → Stacks → Your Stack → Editor
2. Verify SMTP variables are in the `environment:` section
3. Click "Update the stack" again
4. Wait 2 minutes for restart
5. Check `/health/email` endpoint again

---

### Issue: "535 Authentication failed" in logs

**Problem:** Invalid Gmail credentials

**Solution:**
1. Verify you're using **app password** (not regular password)
2. Remove spaces from app password: `abcd efgh ijkl mnop` → `abcdefghijklmnop`
3. Generate new app password: https://myaccount.google.com/apppasswords
4. Update `SMTP_PASSWORD` in Portainer
5. Restart stack

---

### Issue: Connection timeout

**Problem:** Port 587 blocked by firewall

**Solution:**
```bash
# Test SMTP connectivity
telnet smtp.gmail.com 587
# Should connect immediately

# If timeout, try port 465 instead:
# Update in Portainer:
- SMTP_PORT=465
```

---

### Issue: Emails going to spam

**Problem:** Gmail SMTP with different FROM_EMAIL domain

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

## Verification Scripts

Use these scripts for automated verification:

### Linux/macOS
```bash
cd /path/to/URLshorter
chmod +x scripts/verify-smtp-config.sh
./scripts/verify-smtp-config.sh
```

### Windows (PowerShell)
```powershell
cd D:\github\URLshorter
.\scripts\verify-smtp-config.ps1
```

---

## Rollback Plan

If SMTP configuration causes issues:

### Quick Rollback
1. **Open Portainer** → Stacks → Your Stack → Editor
2. **Remove or comment out** the 6 SMTP variables:
   ```yaml
   # - SMTP_HOST=smtp.gmail.com
   # - SMTP_PORT=587
   # - SMTP_USERNAME=your.email@gmail.com
   # - SMTP_PASSWORD=abcdefghijklmnop
   # - FROM_EMAIL=noreply@trunc.site
   # - FROM_NAME=trunc.site
   ```
3. **Click:** "Update the stack"
4. **Result:** Emails will revert to console mode (logged, not sent)

### Temporary Disable SMTP
```yaml
# Keep variables but set SMTP_HOST to empty
- SMTP_HOST=
```

This will make the email service fall back to console provider without removing the configuration.

---

## Security Notes

✅ **DO:**
- Use app-specific passwords (Gmail app password, SendGrid API key)
- Keep SMTP credentials in Portainer environment variables only
- Never commit `.env` file to git (already in `.gitignore`)
- Enable 2FA on email accounts

❌ **DON'T:**
- Use regular Gmail password for SMTP
- Commit SMTP credentials to git repository
- Share SMTP credentials in plain text
- Use SMTP credentials for multiple services

---

## Success Criteria

You'll know the fix is successful when:

- ✅ `/health/email` shows `"provider": "smtp"`
- ✅ No "EMAIL (CONSOLE)" messages in logs
- ✅ Logs show "email sent successfully"
- ✅ Password reset emails arrive in inbox
- ✅ Reset links work correctly
- ✅ Users can reset their passwords

---

## Next Steps After Success

1. **Document credentials:** Save SMTP credentials in password manager
2. **Monitor email deliverability:** Check spam placement for different providers
3. **Consider professional email service:** For production (SendGrid, Mailgun, AWS SES)
4. **Configure DNS records:** SPF, DKIM, DMARC for better deliverability
5. **Set up email monitoring:** Track delivery rates and failures

---

## Additional Resources

- **Quick Start:** `SMTP-QUICKSTART.md` (15 min guide)
- **Full Guide:** `SMTP-CONFIGURATION-GUIDE.md` (comprehensive setup)
- **Verification Script (Linux):** `scripts/verify-smtp-config.sh`
- **Verification Script (Windows):** `scripts/verify-smtp-config.ps1`

---

## Support

### Common Commands

```bash
# Check container status
docker ps | grep trunc

# View logs (last 100 lines)
docker logs <container-id> --tail=100

# Follow logs in real-time
docker logs -f <container-id>

# Check environment variables
docker exec <container-id> env | grep SMTP

# Restart container
docker restart <container-id>

# Test SMTP connectivity
telnet smtp.gmail.com 587
```

### Getting Help

If you encounter issues:

1. **Check logs** for specific error messages
2. **Verify health endpoint:** `/health/email`
3. **Review** `SMTP-CONFIGURATION-GUIDE.md` for detailed troubleshooting
4. **Test SMTP connectivity** with telnet
5. **Check Portainer** environment variables are set correctly

---

**Ready to implement?** Start with Step 1: Generate Gmail App Password
