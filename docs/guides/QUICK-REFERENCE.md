# Password Reset Email Fix - Quick Reference Card

---

## 🎯 The Fix (One Page)

### Problem
```
========== EMAIL (CONSOLE) ==========  ← Emails logged to console, not sent
```

### Solution
Add 6 environment variables to Portainer → Update stack → Done

---

## 📋 Implementation (15 minutes)

### 1. Get Gmail App Password (5 min)
```
1. https://myaccount.google.com/apppasswords
2. Create: Mail → Other → "trunc.site SMTP"
3. Copy 16-char password: abcdefghijklmnop
```

### 2. Add to Portainer (5 min)
```yaml
Portainer → Stacks → trunc.site → Editor → app service → environment:

- SMTP_HOST=smtp.gmail.com
- SMTP_PORT=587
- SMTP_USERNAME=your.email@gmail.com
- SMTP_PASSWORD=abcdefghijklmnop
- FROM_EMAIL=noreply@trunc.site
- FROM_NAME=trunc.site

Click: "Update the stack" → Wait 2 minutes
```

### 3. Verify (5 min)
```
Browser: https://trunc.site/health/email
✅ Should show: "provider": "smtp"
❌ Still "console"? Check Portainer variables

Test: https://trunc.site/forgot-password
✅ Email arrives in inbox within 2 minutes
```

---

## 🔍 Quick Diagnostics

### Check Email Provider
```bash
curl https://trunc.site/health/email | jq
# Should show: "provider": "smtp"
```

### Check Container Variables
```bash
docker ps | grep trunc                     # Get container ID
docker exec <id> env | grep SMTP           # Check variables
docker logs <id> --tail=50 | grep -i email # Check logs
```

### Expected Logs (Success)
```
INFO [EMAIL] Initializing email service (provider=smtp)
INFO [EMAIL] SMTP configuration: host=smtp.gmail.com
INFO [EMAIL] Password reset email sent successfully
```

### Bad Logs (Still Broken)
```
========== EMAIL (CONSOLE) ==========     # Still using console
```

---

## 🚨 Troubleshooting (1-Liners)

| Issue | Quick Fix |
|-------|-----------|
| Provider still "console" | Update stack again, wait 2 min, check variables |
| "535 Authentication failed" | Use app password, not regular password |
| App password has spaces | Remove spaces: `abcd efgh` → `abcdefgh` |
| Connection timeout | Test: `telnet smtp.gmail.com 587` |
| Emails to spam | Change FROM_EMAIL to your Gmail address |

---

## ✅ Success Indicators

```
1. curl https://trunc.site/health/email
   → "provider": "smtp" ✅

2. docker logs <id> | grep "EMAIL (CONSOLE)"
   → No output ✅

3. https://trunc.site/forgot-password
   → Email arrives in inbox ✅

4. Reset link works ✅
```

---

## 🔄 Quick Rollback

```yaml
Portainer → Editor → Comment out SMTP vars:

# - SMTP_HOST=smtp.gmail.com
# - SMTP_PORT=587
# - SMTP_USERNAME=...
# - SMTP_PASSWORD=...
# - FROM_EMAIL=...
# - FROM_NAME=...

Update stack → Reverts to console mode
```

---

## 📚 Full Documentation

- **Step-by-step:** `IMPLEMENTATION-STEPS.md`
- **Checklist:** `IMPLEMENTATION-CHECKLIST.md`
- **Quick start:** `SMTP-QUICKSTART.md`
- **Full guide:** `SMTP-CONFIGURATION-GUIDE.md`

---

## 🔐 Security Notes

✅ **DO:**
- Use Gmail app password (16 chars)
- Enable 2FA on Gmail account
- Save password in password manager

❌ **DON'T:**
- Use regular Gmail password
- Commit credentials to git
- Share credentials in plain text

---

## 🎓 Key Commands

```bash
# Container status
docker ps | grep trunc

# Check variables
docker exec <id> env | grep SMTP

# View logs
docker logs <id> --tail=100 | grep -i email

# Follow logs
docker logs -f <id>

# Test SMTP
telnet smtp.gmail.com 587

# Restart container
docker restart <id>
```

---

## 💡 Important Notes

1. **No code changes needed** - This is pure configuration
2. **Risk is LOW** - Can rollback in 1 minute
3. **Time required** - 15 minutes total
4. **Already deployed** - Code is ready, just needs config
5. **Health endpoint** - Use `/health/email` to verify

---

## 📞 Support

**Still stuck?** Check logs for exact error:
```bash
docker logs <container-id> 2>&1 | grep -i "failed\|error" | tail -20
```

The error message will tell you exactly what's wrong.

---

**Print this page for quick reference during implementation** 🖨️
