# SMTP Email Configuration - Deployment Checklist

Use this checklist to track your progress deploying the SMTP email fix.

**Estimated Time:** 15-20 minutes
**Difficulty:** Easy (configuration only)

---

## Pre-Deployment Checklist

### Prerequisites
- [ ] Have access to Portainer (production server)
- [ ] Have SSH access to Oracle server
- [ ] Know which email address to use for SMTP (Gmail recommended for testing)
- [ ] Have 2FA enabled on Google account (if using Gmail)

---

## Phase 1: Gmail App Password (5 minutes)

**Skip this section if using a different SMTP provider**

- [ ] Visit: https://myaccount.google.com/security
- [ ] Verify 2-Factor Authentication is enabled
  - If not enabled: Click "2-Step Verification" and follow setup
- [ ] Visit: https://myaccount.google.com/apppasswords
- [ ] Create new app password:
  - [ ] Select "Mail" from dropdown
  - [ ] Select "Other (Custom name)"
  - [ ] Enter name: "trunc.site SMTP"
  - [ ] Click "Generate"
- [ ] Copy the 16-character password
  - Format: `xxxx xxxx xxxx xxxx`
  - **Remove spaces** for final value: `xxxxxxxxxxxxxxxx`
- [ ] Save password securely (you won't see it again)

**App Password:** `____________________________________`

---

## Phase 2: Portainer Configuration (5 minutes)

### Step 1: Access Portainer
- [ ] Open browser and navigate to Portainer
- [ ] Login with credentials
- [ ] Navigate to **Stacks** section
- [ ] Find and click on your trunc.site stack

### Step 2: Edit Stack Configuration
- [ ] Click the **"Editor"** tab
- [ ] Find the `app` service in the YAML
- [ ] Locate the `environment:` section
- [ ] Scroll to find where environment variables are defined

### Step 3: Add SMTP Variables
- [ ] Add these 6 lines to the `environment:` section:

```yaml
# SMTP Email Configuration
- SMTP_HOST=smtp.gmail.com
- SMTP_PORT=587
- SMTP_USERNAME=your.email@gmail.com
- SMTP_PASSWORD=xxxxxxxxxxxxxxxx
- FROM_EMAIL=noreply@trunc.site
- FROM_NAME=trunc.site
```

- [ ] Replace `your.email@gmail.com` with your actual Gmail address
- [ ] Replace `xxxxxxxxxxxxxxxx` with your app password (from Phase 1)
- [ ] Optionally customize `FROM_EMAIL` and `FROM_NAME`

### Your Configuration:
```
SMTP_HOST: smtp.gmail.com
SMTP_PORT: 587
SMTP_USERNAME: ____________________________________
SMTP_PASSWORD: ____________________________________
FROM_EMAIL: ____________________________________
FROM_NAME: ____________________________________
```

### Step 4: Deploy Changes
- [ ] Scroll to bottom of page
- [ ] Click **"Update the stack"** button
- [ ] Optional: Enable "Re-pull image and redeploy" if using latest tag
- [ ] Click **"Update"** to confirm
- [ ] Wait for deployment to complete (1-2 minutes)
- [ ] Verify container restarted successfully

**Deployment Time:** ______________ (note time for reference)

---

## Phase 3: Verification (5 minutes)

### Step 1: SSH Verification
- [ ] SSH into Oracle server:
  ```bash
  ssh user@your-oracle-server-ip
  ```

- [ ] Navigate to project directory:
  ```bash
  cd /path/to/URLshorter
  ```

- [ ] Make verification script executable (if not already):
  ```bash
  chmod +x scripts/verify-smtp-config.sh
  ```

- [ ] Run verification script:
  ```bash
  ./scripts/verify-smtp-config.sh
  ```

- [ ] Review output and confirm all checks pass:
  - [ ] ✓ Container found
  - [ ] ✓ ENVIRONMENT: production
  - [ ] ✓ SMTP_HOST: smtp.gmail.com
  - [ ] ✓ SMTP_PORT: 587
  - [ ] ✓ SMTP_USERNAME: (your email)
  - [ ] ✓ SMTP_PASSWORD: **************** (Set)
  - [ ] ✓ FROM_EMAIL: (your from email)
  - [ ] ✓ FROM_NAME: (your from name)
  - [ ] ✓ No SMTP provider creation errors
  - [ ] ✓ No console email logs found
  - [ ] ✓ SMTP server is reachable

**If any checks fail, see Troubleshooting section at bottom**

### Step 2: Manual Log Verification
- [ ] Check container logs for errors:
  ```bash
  docker logs $(docker ps | grep app | awk '{print $1}') --tail=50 | grep -i email
  ```

- [ ] Verify NO "EMAIL (CONSOLE)" output:
  ```bash
  docker logs $(docker ps | grep app | awk '{print $1}') 2>&1 | grep "EMAIL (CONSOLE)"
  ```
  - **Expected:** No output (if output found, SMTP not working)

- [ ] Verify NO "failed to create email provider" errors:
  ```bash
  docker logs $(docker ps | grep app | awk '{print $1}') 2>&1 | grep "failed to create email provider"
  ```
  - **Expected:** No output (if errors found, check SMTP credentials)

---

## Phase 4: End-to-End Testing (5 minutes)

### Test 1: Request Password Reset
- [ ] Open browser and navigate to: https://trunc.site/forgot-password
- [ ] Enter your email address (use your own for testing)
- [ ] Click **"Send Reset Link"** button
- [ ] Verify success message appears
- [ ] Note timestamp: ______________

### Test 2: Check Server Logs
- [ ] SSH into server (if not already connected)
- [ ] Watch logs in real-time:
  ```bash
  docker logs -f $(docker ps | grep app | awk '{print $1}')
  ```

- [ ] Verify log output:
  - [ ] ✓ See: `INFO Password reset requested for email=...`
  - [ ] ✗ DO NOT see: `========== EMAIL (CONSOLE) ==========`
  - If you see console output, SMTP is NOT working - go to Troubleshooting

- [ ] Press Ctrl+C to exit log tail

### Test 3: Check Email Inbox
- [ ] Open email inbox (the one you used in Test 1)
- [ ] Wait 1-2 minutes for email to arrive
- [ ] Check inbox folder
- [ ] If not in inbox, check spam/junk folder
- [ ] Verify email received:
  - [ ] From: `trunc.site <noreply@trunc.site>` (or your configured FROM_EMAIL)
  - [ ] Subject: `Reset Your Password - trunc.site`
  - [ ] Contains reset link button
  - [ ] Link format: `https://trunc.site/reset-password?token=...`

**Email Received Time:** ______________ (note delivery time)

**If email not received after 5 minutes, go to Troubleshooting section**

### Test 4: Password Reset Flow
- [ ] Click the reset link in the email
- [ ] Verify page loads: https://trunc.site/reset-password?token=...
- [ ] Enter new password in both fields
- [ ] Click **"Reset Password"** button
- [ ] Verify success message appears

### Test 5: Login with New Password
- [ ] Navigate to: https://trunc.site/login
- [ ] Enter your email
- [ ] Enter the NEW password (from Test 4)
- [ ] Click **"Login"** button
- [ ] Verify successful login

---

## Phase 5: Final Verification

### Production Readiness Checks
- [ ] All verification script checks pass
- [ ] No console email output in logs
- [ ] Password reset email received successfully
- [ ] Email delivery time < 2 minutes
- [ ] Email NOT in spam folder (note if it is)
- [ ] Reset link works correctly
- [ ] Can login with new password
- [ ] Complete end-to-end flow works

### Documentation
- [ ] SMTP credentials saved securely
- [ ] Deployment timestamp recorded: ______________
- [ ] Configuration backed up (if needed)
- [ ] Team notified of email fix (if applicable)

---

## Troubleshooting

### Issue: Verification Script Shows Errors

**Missing Environment Variables:**
```
✗ SMTP_HOST: NOT SET (Required)
```

**Action:**
1. Go back to Portainer
2. Verify variables were added correctly
3. Update stack again
4. Wait for container restart
5. Run verification script again

---

### Issue: "Failed to create email provider" in Logs

**Possible causes:**
- Wrong SMTP credentials
- Network/firewall blocking SMTP port

**Action:**
1. Check SMTP credentials are correct
2. Test SMTP connectivity:
   ```bash
   telnet smtp.gmail.com 587
   ```
3. If connection fails, port 587 might be blocked
4. Try port 465 instead: `SMTP_PORT=465`

---

### Issue: Still Seeing "EMAIL (CONSOLE)" Output

**Cause:** SMTP_HOST not set or empty

**Action:**
1. Check environment variable in container:
   ```bash
   docker exec $(docker ps | grep app | awk '{print $1}') env | grep SMTP_HOST
   ```
2. If empty, add in Portainer and update stack
3. Restart container:
   ```bash
   docker restart $(docker ps | grep app | awk '{print $1}')
   ```

---

### Issue: "535 Authentication failed"

**Cause:** Using regular Gmail password instead of app password

**Action:**
1. Generate new app password: https://myaccount.google.com/apppasswords
2. Update SMTP_PASSWORD in Portainer
3. Update stack
4. Test again

---

### Issue: Email Not Received After 5 Minutes

**Check:**
1. Email in spam/junk folder?
2. Correct email address entered?
3. Check logs for SMTP send errors

**Action:**
```bash
# Check for SMTP send errors
docker logs $(docker ps | grep app | awk '{print $1}') 2>&1 | grep -i "smtp\|email" | tail -20
```

If errors found, note the error message and:
- Google the specific error
- Check SMTP-CONFIGURATION-GUIDE.md for solutions
- Verify SMTP credentials are correct

---

### Issue: Emails Going to Spam

**Short-term fix:**
```yaml
# Use Gmail address as FROM_EMAIL
- FROM_EMAIL=your.email@gmail.com
- FROM_NAME=trunc.site
```

**Long-term fix:**
- Switch to professional email service (SendGrid, Mailgun)
- Configure DNS records (SPF, DKIM, DMARC)
- See SMTP-CONFIGURATION-GUIDE.md for details

---

## Success! ✅

If all checks pass:

- ✅ SMTP email delivery is working
- ✅ Password reset flow is functional
- ✅ Production ready

**Next steps:**
1. Monitor email deliverability over next few days
2. Consider switching to professional email service for better deliverability
3. Configure DNS records (SPF, DKIM, DMARC) if using custom domain
4. Set up email monitoring/alerts

---

## Rollback Plan

If SMTP causes issues and you need to revert:

1. **Option A: Remove SMTP variables**
   - Open Portainer → Stacks → Your Stack → Editor
   - Remove or comment out all SMTP_ variables
   - Update stack
   - System will revert to console mode (no emails sent, but no errors)

2. **Option B: Temporarily disable**
   - Set `SMTP_HOST=` (empty string)
   - Update stack
   - System will fall back to console mode

---

## Notes & Observations

Use this space to note any issues, observations, or customizations:

```
Deployment Notes:
- Deployment started: ______________
- Issues encountered: ______________
- Resolution: ______________
- Final status: ______________





```

---

## Completion Sign-Off

- [ ] All phases completed successfully
- [ ] All tests passed
- [ ] No errors in logs
- [ ] Production verified working
- [ ] Documentation updated (if needed)

**Deployed by:** ____________________
**Date:** ____________________
**Time:** ____________________
**Status:** ☐ Success  ☐ Issues (see notes)

---

**For detailed troubleshooting, see:** SMTP-CONFIGURATION-GUIDE.md
**For quick reference, see:** SMTP-QUICKSTART.md
