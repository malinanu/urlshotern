# SMTP Email Configuration Guide for Portainer

## Quick Fix: Password Reset Emails Not Sending

**Problem:** Emails are logging to console instead of being sent via SMTP
**Solution:** Add SMTP environment variables to Portainer deployment
**Time Required:** 10-15 minutes

---

## Prerequisites

### If Using Gmail (Recommended for Testing)

1. **Enable 2-Factor Authentication:**
   - Visit: https://myaccount.google.com/security
   - Enable 2-Step Verification if not already enabled

2. **Generate App Password:**
   - Visit: https://myaccount.google.com/apppasswords
   - Select "Mail" and "Other (Custom name)"
   - Enter name: `trunc.site SMTP`
   - Click "Generate"
   - **Copy the 16-character password** (format: `xxxx xxxx xxxx xxxx`)
   - **Remove spaces** when entering (final format: `xxxxxxxxxxxxxxxx`)

---

## Step 1: SSH into Oracle Server

```bash
# Connect to your Oracle server
ssh user@your-oracle-server-ip

# Verify current deployment
cd /path/to/URLshorter
docker ps | grep trunc
```

---

## Step 2: Configure Environment Variables in Portainer

### Option A: Via Portainer Web UI (Recommended)

1. **Open Portainer:**
   - Navigate to: `http://your-server-ip:9000` (or your Portainer URL)
   - Login to Portainer

2. **Find Your Stack:**
   - Go to: **Stacks** → Find your trunc.site stack
   - Click on the stack name

3. **Edit Stack:**
   - Click the **"Editor"** tab
   - Find the `app` service → `environment` section

4. **Add SMTP Variables:**

   Find this section in the docker-compose file:
   ```yaml
   environment:
     - ENVIRONMENT=production
     - DB_HOST=postgres
     # ... other variables ...
   ```

   **Add these lines** (replace with your actual values):
   ```yaml
     # SMTP Email Configuration
     - SMTP_HOST=smtp.gmail.com
     - SMTP_PORT=587
     - SMTP_USERNAME=your.email@gmail.com
     - SMTP_PASSWORD=your-app-password-here
     - FROM_EMAIL=noreply@trunc.site
     - FROM_NAME=trunc.site
   ```

5. **Update Stack:**
   - Click **"Update the stack"** button at the bottom
   - Enable: **"Re-pull image and redeploy"** (if using latest tag)
   - Click **"Update"**
   - Wait 1-2 minutes for container to restart

### Option B: Via Environment Variables Section

1. **In Portainer:**
   - Go to: **Stacks** → Your Stack → **Environment variables** tab

2. **Add each variable:**
   - Click **"+ Add environment variable"**
   - Add each of these:
     ```
     SMTP_HOST = smtp.gmail.com
     SMTP_PORT = 587
     SMTP_USERNAME = your.email@gmail.com
     SMTP_PASSWORD = your-app-password-here
     FROM_EMAIL = noreply@trunc.site
     FROM_NAME = trunc.site
     ```

3. **Update and Redeploy:**
   - Click **"Update the stack"**

---

## Step 3: Verify Configuration

### Check Container Environment Variables

```bash
# SSH into Oracle server
ssh user@your-oracle-server

# Find container ID
CONTAINER_ID=$(docker ps | grep trunc-site-app | awk '{print $1}')

# Verify SMTP variables are set
docker exec $CONTAINER_ID env | grep SMTP

# Expected output:
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USERNAME=your.email@gmail.com
# SMTP_PASSWORD=xxxxxxxxxxxxxxxx
```

### Check Application Logs

```bash
# Check for SMTP errors
docker logs $CONTAINER_ID 2>&1 | grep -i "failed to create email provider"

# If you see this error, SMTP is failing
# If no output, SMTP provider was created successfully

# Check for console email output (BAD - means SMTP not working)
docker logs $CONTAINER_ID 2>&1 | grep "EMAIL (CONSOLE)"

# If you see this, emails are still going to console (not fixed yet)
```

---

## Step 4: Test Email Delivery

### Test Password Reset Flow

1. **Go to Password Reset Page:**
   - Navigate to: https://trunc.site/forgot-password

2. **Enter Your Email:**
   - Use your own email for testing
   - Click "Send Reset Link"

3. **Check Server Logs:**
   ```bash
   # Watch logs in real-time
   docker logs -f $CONTAINER_ID

   # You should see:
   # INFO Password reset requested for email
   #
   # You should NOT see:
   # ========== EMAIL (CONSOLE) ==========
   ```

4. **Check Your Email Inbox:**
   - Wait 1-2 minutes
   - Check inbox (and spam folder)
   - Look for email from: `trunc.site <noreply@trunc.site>`
   - Subject: "Reset Your Password - trunc.site"

5. **Click Reset Link:**
   - Should open: `https://trunc.site/reset-password?token=...`
   - Enter new password
   - Submit and verify success

6. **Test Login:**
   - Login with the new password to confirm it works

---

## Gmail-Specific Configuration

### Correct Gmail Settings

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587                                    # Use 587 for STARTTLS
SMTP_USERNAME=your.actual.email@gmail.com       # FULL email address
SMTP_PASSWORD=abcdefghijklmnop                  # 16-char app password (no spaces)
FROM_EMAIL=noreply@trunc.site                   # Can be different from SMTP_USERNAME
FROM_NAME=trunc.site
```

### Common Gmail Issues

| Issue | Solution |
|-------|----------|
| "535 Authentication failed" | Need app password, not regular password |
| "Username and Password not accepted" | Remove spaces from app password |
| "Less secure apps" error | Use app password (required for Gmail) |
| Emails going to spam | Gmail's SMTP with different FROM_EMAIL may trigger spam filters |

### Testing Gmail SMTP Connection

```bash
# From Oracle server, test SMTP connection
telnet smtp.gmail.com 587

# Should see:
# 220 smtp.gmail.com ESMTP ...
# (Press Ctrl+] then type 'quit' to exit)

# Or use OpenSSL
openssl s_client -connect smtp.gmail.com:587 -starttls smtp
```

---

## Alternative SMTP Providers (For Production)

Gmail has limitations (500 emails/day). For production, consider:

### SendGrid (Recommended)
- **Free Tier:** 100 emails/day (forever)
- **Configuration:**
  ```
  SMTP_HOST=smtp.sendgrid.net
  SMTP_PORT=587
  SMTP_USERNAME=apikey
  SMTP_PASSWORD=<your-sendgrid-api-key>
  ```

### Mailgun
- **Free Tier:** 5,000 emails/month (3 months)
- **Configuration:**
  ```
  SMTP_HOST=smtp.mailgun.org
  SMTP_PORT=587
  SMTP_USERNAME=<your-username>
  SMTP_PASSWORD=<your-password>
  ```

### AWS SES
- **Free Tier:** 62,000 emails/month (if sending from EC2)
- **Configuration:**
  ```
  SMTP_HOST=email-smtp.us-east-1.amazonaws.com
  SMTP_PORT=587
  SMTP_USERNAME=<your-ses-username>
  SMTP_PASSWORD=<your-ses-password>
  ```

---

## Troubleshooting

### Issue: Emails Still Logging to Console

**Diagnosis:**
```bash
# Check if SMTP_HOST is set
docker exec $CONTAINER_ID env | grep SMTP_HOST

# If empty or not set, add it to Portainer
```

**Fix:** Ensure all SMTP variables are set in Portainer environment section.

---

### Issue: "Failed to create email provider" Error

**Diagnosis:**
```bash
# Check logs for specific error
docker logs $CONTAINER_ID 2>&1 | grep "failed to create email provider"
```

**Common Causes:**
- Wrong SMTP credentials → Check username/password
- Port blocked by firewall → Test with `telnet smtp.gmail.com 587`
- Invalid SMTP host → Verify `SMTP_HOST` value

---

### Issue: Connection Timeout

**Diagnosis:**
```bash
# Test SMTP port connectivity
telnet smtp.gmail.com 587
```

**Fix:**
- Port 587 might be blocked by firewall
- Try port 465 (direct TLS): `SMTP_PORT=465`
- Contact hosting provider to open port 587

---

### Issue: Emails Going to Spam

**Causes:**
- Using Gmail SMTP with different `FROM_EMAIL` domain
- No SPF/DKIM/DMARC records configured

**Short-term Fix:**
```bash
# Use Gmail address as FROM_EMAIL
FROM_EMAIL=your.email@gmail.com
FROM_NAME=trunc.site
```

**Long-term Fix:**
- Use professional email service (SendGrid, Mailgun)
- Configure DNS records (SPF, DKIM, DMARC) for your domain

---

## Verification Checklist

- [ ] Gmail app password generated (if using Gmail)
- [ ] All 6 SMTP environment variables added to Portainer
- [ ] Stack updated and container restarted
- [ ] `docker exec` confirms SMTP variables are set
- [ ] No "failed to create email provider" errors in logs
- [ ] No "EMAIL (CONSOLE)" output in logs
- [ ] Test password reset email sent successfully
- [ ] Email received in inbox (not spam)
- [ ] Reset link works correctly
- [ ] Can login with new password

---

## Rollback Plan

If SMTP configuration causes issues:

```bash
# Remove SMTP environment variables in Portainer
# Or set SMTP_HOST to empty string
SMTP_HOST=

# This will revert to console mode (no emails sent, but no errors)
```

---

## Security Best Practices

1. **Never commit SMTP credentials to git:**
   ```bash
   # Add to .gitignore
   .env
   .env.production
   ```

2. **Use environment variables only:**
   - Store in Portainer environment variables
   - Or in .env file on server (not in git)

3. **Use app-specific passwords:**
   - Gmail: app password (not regular password)
   - SendGrid: API key (not account password)

4. **Configure DNS records** (for production):
   ```
   SPF:   v=spf1 include:_spf.google.com ~all
   DKIM:  (generated by email provider)
   DMARC: v=DMARC1; p=quarantine; rua=mailto:postmaster@trunc.site
   ```

---

## Next Steps After Email Works

1. **Monitor deliverability:**
   - Check inbox vs spam placement
   - Test with multiple email providers (Gmail, Outlook, Yahoo)

2. **Consider professional email service:**
   - Better deliverability than Gmail
   - Dedicated IP address
   - Detailed analytics and bounce handling
   - Higher sending limits

3. **Configure email templates:**
   - Customize email design in `internal/services/email/templates.go`
   - Add company branding

4. **Set up monitoring:**
   - Track email sending errors
   - Alert on delivery failures
   - Monitor bounce rates

---

## Support

If you encounter issues not covered in this guide:

1. **Check application logs:**
   ```bash
   docker logs $CONTAINER_ID --tail=200 | grep -i email
   ```

2. **Test SMTP manually:**
   ```bash
   # Install swaks for testing
   docker exec -it $CONTAINER_ID sh
   apk add perl-net-smtp perl-io-socket-ssl

   swaks --to test@example.com \
         --from noreply@trunc.site \
         --server smtp.gmail.com:587 \
         --auth LOGIN \
         --auth-user your.email@gmail.com \
         --auth-password your-app-password \
         --tls
   ```

3. **Review error messages:**
   - SMTP errors are detailed and usually indicate the exact problem
   - Common errors: authentication, connection timeout, TLS issues
