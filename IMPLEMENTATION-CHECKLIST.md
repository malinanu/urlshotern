# Password Reset Email Fix - Implementation Checklist

**Date:** _____________
**Implemented By:** _____________
**Time Started:** _____________

---

## Pre-Implementation

- [ ] Read `IMPLEMENTATION-STEPS.md`
- [ ] Backup current Portainer stack configuration
- [ ] Have access to Portainer
- [ ] Have access to Gmail account (or other SMTP provider)

---

## Step 1: Generate SMTP Credentials (5 min)

### Gmail App Password
- [ ] Go to https://myaccount.google.com/security
- [ ] Verify 2FA is enabled
- [ ] Go to https://myaccount.google.com/apppasswords
- [ ] Create app password: Mail → Other → "trunc.site SMTP"
- [ ] Copy 16-character password (no spaces): ___________________________
- [ ] Save password in password manager

---

## Step 2: Configure Portainer (5 min)

### Add SMTP Variables
- [ ] Open Portainer web interface
- [ ] Navigate to: Stacks → Find trunc.site stack
- [ ] Click: Editor tab
- [ ] Find `app` service → `environment:` section
- [ ] Add SMTP_HOST=smtp.gmail.com
- [ ] Add SMTP_PORT=587
- [ ] Add SMTP_USERNAME=(your email)
- [ ] Add SMTP_PASSWORD=(app password)
- [ ] Add FROM_EMAIL=noreply@trunc.site
- [ ] Add FROM_NAME=trunc.site
- [ ] Click: "Update the stack"
- [ ] Wait 1-2 minutes for restart
- [ ] Verify: Container status is green/running

---

## Step 3: Verify Configuration (5 min)

### Health Endpoint Check
- [ ] Open browser: https://trunc.site/health/email
- [ ] Verify: `"provider": "smtp"` (not "console")
- [ ] Verify: `"status": "ok"`
- [ ] Verify: SMTP config shows correct host and username

**If provider is still "console", go to Troubleshooting section**

### SSH Verification (Optional)
- [ ] SSH into Oracle server
- [ ] Run: `docker ps | grep trunc`
- [ ] Note container ID: ___________________________
- [ ] Run: `docker exec <container-id> env | grep SMTP`
- [ ] Verify: SMTP_HOST=smtp.gmail.com
- [ ] Verify: SMTP_USERNAME is correct
- [ ] Verify: SMTP_PASSWORD is set (not empty)

### Log Check
- [ ] Run: `docker logs <container-id> --tail=50 | grep -i email`
- [ ] Verify: See "Initializing email service (provider=smtp)"
- [ ] Verify: NO "EMAIL (CONSOLE)" messages
- [ ] Verify: See "SMTP configuration: host=smtp.gmail.com"

---

## Step 4: Test Email Delivery (5 min)

### Request Password Reset
- [ ] Open: https://trunc.site/forgot-password
- [ ] Enter email: ___________________________
- [ ] Click: "Send Reset Link"
- [ ] Note timestamp: ___________________________

### Monitor Logs
- [ ] Run: `docker logs -f <container-id>`
- [ ] Verify: See "Sending password reset email to..."
- [ ] Verify: See "Password reset email sent successfully"
- [ ] Verify: NO "EMAIL (CONSOLE)" output

### Check Email Inbox
- [ ] Check inbox (and spam folder)
- [ ] Verify: Email received within 2 minutes
- [ ] Verify: From: "trunc.site <noreply@trunc.site>"
- [ ] Verify: Subject: "Reset Your Password - trunc.site"
- [ ] Click: Reset link
- [ ] Verify: Opens https://trunc.site/reset-password?token=...
- [ ] Enter new password
- [ ] Submit form
- [ ] Verify: Success message displayed

### Verify Password Changed
- [ ] Go to: https://trunc.site/admin
- [ ] Login with new password
- [ ] Verify: Login successful

---

## Post-Implementation Verification

### Production Checks
- [ ] No errors in application logs
- [ ] Password reset flow works end-to-end
- [ ] Emails arrive in inbox (not spam)
- [ ] Reset links are valid and work
- [ ] Users can login after password reset

### Performance Checks
- [ ] Email delivery time < 2 minutes
- [ ] No SMTP connection errors in logs
- [ ] Application performance unchanged

---

## Documentation

- [ ] Save SMTP credentials in password manager
- [ ] Document which email service is used (Gmail/SendGrid/etc)
- [ ] Update team documentation with SMTP provider info
- [ ] Note any configuration changes made

---

## Troubleshooting (If Needed)

### Issue: Provider still shows "console"

- [ ] Checked SMTP variables in Portainer
- [ ] Variables are in the correct `environment:` section
- [ ] Clicked "Update the stack" in Portainer
- [ ] Waited 2+ minutes for restart
- [ ] Container restarted successfully
- [ ] Re-checked `/health/email` endpoint

**Resolution:** ___________________________

---

### Issue: Authentication failed

- [ ] Verified using app password (not regular password)
- [ ] Removed spaces from app password
- [ ] Generated new app password
- [ ] Updated SMTP_PASSWORD in Portainer
- [ ] Restarted stack

**Resolution:** ___________________________

---

### Issue: Connection timeout

- [ ] Tested: `telnet smtp.gmail.com 587`
- [ ] Verified port 587 is not blocked
- [ ] Tried port 465 instead
- [ ] Contacted hosting provider about firewall

**Resolution:** ___________________________

---

### Issue: Emails going to spam

- [ ] Changed FROM_EMAIL to Gmail address
- [ ] Tested with multiple email providers
- [ ] Considered professional email service

**Resolution:** ___________________________

---

## Rollback (If Needed)

If issues occur:

- [ ] Opened Portainer → Stacks → Editor
- [ ] Commented out SMTP variables
- [ ] Clicked "Update the stack"
- [ ] Verified: Container restarted
- [ ] Checked: Service still accessible
- [ ] Noted rollback reason: ___________________________

---

## Sign-Off

### Implementation Complete
- [ ] All verification steps passed
- [ ] No errors in production logs
- [ ] Email delivery confirmed working
- [ ] Users can reset passwords successfully
- [ ] Documentation updated

**Completion Time:** _____________

**Implemented By:** _____________

**Verified By:** _____________

**Date:** _____________

---

## Notes

Use this space to note any issues, observations, or additional steps taken:

___________________________________________________________________________

___________________________________________________________________________

___________________________________________________________________________

___________________________________________________________________________

___________________________________________________________________________

---

## Success Criteria Met

✅ `/health/email` shows `"provider": "smtp"`
✅ No "EMAIL (CONSOLE)" in logs
✅ Password reset emails delivered to inbox
✅ Reset links work correctly
✅ Users can reset passwords

**Status: COMPLETE / INCOMPLETE**

---

## Next Steps (After Success)

- [ ] Monitor email deliverability for 24 hours
- [ ] Check spam placement across different providers
- [ ] Consider professional email service for production
- [ ] Configure DNS records (SPF, DKIM, DMARC)
- [ ] Set up email delivery monitoring/alerting

---

**For detailed implementation guide, see:** `IMPLEMENTATION-STEPS.md`
**For troubleshooting help, see:** `SMTP-CONFIGURATION-GUIDE.md`
