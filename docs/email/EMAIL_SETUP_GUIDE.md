# Email Setup Guide for trunc.site

This guide explains how to configure and use the email functionality in your URL shortener application.

## Email Configuration

### Environment Variables

Add these variables to your `.env` file:

```bash
# SMTP Configuration
SMTP_HOST=mail.trunc.site           # Your SMTP server hostname
SMTP_PORT=587                        # SMTP port (587 for STARTTLS, 465 for TLS)
SMTP_USERNAME=noreply@trunc.site    # SMTP username
SMTP_PASSWORD=your_smtp_password     # SMTP password

# Email Sender Information
FROM_EMAIL=noreply@trunc.site       # Email address to send from
FROM_NAME=trunc.site                # Display name for sender

# Frontend URL (for email links)
BASE_URL=https://trunc.site     # Your frontend URL
```

### Email Addresses

Your configured email addresses:

- **noreply@trunc.site** - Password resets, verification codes, OTPs
- **hello@trunc.site** - General contact (already created)
- **contact@trunc.site** - Contact form submissions
- **support@trunc.site** - User support inquiries

## Available Email Templates

### 1. Email Verification

Sent when users sign up and need to verify their email address.

```go
emailService.SendEmailVerification(userEmail, verificationToken)
```

**Features:**
- Beautiful branded template with trunc.site branding
- 24-hour expiration notice
- Verification button with fallback link

---

### 2. Password Reset

Sent when users request a password reset.

```go
emailService.SendPasswordReset(userEmail, resetToken)
```

**Features:**
- Security warning notice
- 1-hour expiration
- Reset button with fallback link

---

### 3. Team Invitation

Sent when users are invited to join a team.

```go
emailService.SendTeamInvitation(userEmail, teamName, inviterName, inviteToken)
```

**Features:**
- Personalized invitation message
- 7-day expiration
- Accept invitation button

---

### 4. Contact Form Submission

Sent to `contact@trunc.site` when someone submits the contact form.

```go
emailService.SendContactFormEmail(name, email, message)
```

**Features:**
- Contact details clearly displayed
- Message formatted for easy reading
- Clickable reply-to email link

---

### 5. Support Inquiry

Sent to `support@trunc.site` when users submit support requests.

```go
emailService.SendSupportInquiryEmail(name, email, subject, issue)
```

**Features:**
- Priority notice (respond within 24 hours)
- Issue details formatted
- Quick reply button
- User contact information

---

### 6. One-Time Password (OTP)

Sent for two-factor authentication and verification codes.

```go
emailService.SendOTPEmail(userEmail, otpCode, expiryMinutes)
```

**Features:**
- Large, easy-to-read OTP code
- Expiration time clearly displayed
- Security notice about not sharing the code

## Testing Email Configuration

### 1. Development Mode (Console)

In development, emails are logged to the console instead of being sent:

```bash
ENVIRONMENT=development
```

This will print email contents to your terminal for testing.

### 2. Production Mode (SMTP)

In production, ensure all SMTP settings are configured:

```bash
ENVIRONMENT=production
SMTP_HOST=mail.trunc.site
SMTP_PORT=587
SMTP_USERNAME=noreply@trunc.site
SMTP_PASSWORD=your_actual_password
FROM_EMAIL=noreply@trunc.site
FROM_NAME=trunc.site
```

## SMTP Port Selection

- **Port 587** - STARTTLS (recommended for most providers)
- **Port 465** - Direct TLS/SSL connection
- **Port 25** - Plain SMTP (not recommended, often blocked)

## Email Security Best Practices

1. **Use strong SMTP passwords** - Generate a secure password for your SMTP account
2. **Enable SPF, DKIM, DMARC** - Configure these DNS records for your domain
3. **Rate limiting** - The system includes built-in rate limiting for verification emails (1 per 5 minutes)
4. **Secure storage** - Never commit your `.env` file to version control
5. **Monitor delivery** - Check email delivery rates and bounces regularly

## Common SMTP Providers

### Gmail/Google Workspace

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=app-specific-password
```

**Note:** You need to create an App Password in your Google Account settings.

### Custom Domain Email (cPanel, Plesk, etc.)

```bash
SMTP_HOST=mail.yourdomain.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your_password
```

### SendGrid

```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your_sendgrid_api_key
```

### Mailgun

```bash
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USERNAME=postmaster@yourdomain.com
SMTP_PASSWORD=your_mailgun_password
```

## Troubleshooting

### Emails not sending

1. Check SMTP credentials are correct
2. Verify SMTP port is not blocked by firewall
3. Check application logs for error messages
4. Test SMTP connection with a tool like `telnet` or `openssl`

### Emails going to spam

1. Configure SPF record: `v=spf1 include:your-smtp-provider.com ~all`
2. Set up DKIM signing through your email provider
3. Add DMARC record: `v=DMARC1; p=none; rua=mailto:admin@trunc.site`
4. Ensure reverse DNS (PTR) record is set

### Rate limiting issues

The system limits verification email resends to 1 per 5 minutes per user. This is configurable in `internal/services/email.go`:

```go
if err == nil && time.Since(lastVerification.CreatedAt) < 5*time.Minute {
    return nil, fmt.Errorf("please wait before requesting another verification email")
}
```

## Database Tables

The email system uses the following database table:

### `email_verifications`

```sql
CREATE TABLE email_verifications (
    id UUID PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);
```

## Example Usage in Handlers

```go
// Example: Send verification email after user registration
func (h *AuthHandlers) Register(c *gin.Context) {
    // ... user registration logic ...

    // Send verification email
    verification, err := h.emailService.SendVerificationEmail(user.ID, user.Email)
    if err != nil {
        log.Printf("Failed to send verification email: %v", err)
        // Continue registration even if email fails
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Registration successful. Please check your email to verify your account.",
        "user": user,
    })
}

// Example: Contact form submission
func (h *ContactHandlers) SubmitContactForm(c *gin.Context) {
    var req struct {
        Name    string `json:"name" binding:"required"`
        Email   string `json:"email" binding:"required,email"`
        Message string `json:"message" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Send contact form email
    err := h.emailService.SendContactFormEmail(req.Name, req.Email, req.Message)
    if err != nil {
        log.Printf("Failed to send contact form email: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to send message. Please try again later.",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Thank you for your message. We'll get back to you soon!",
    })
}
```

## Monitoring and Logging

All email operations are logged. Check your application logs for:

- Successful email sends
- Failed email sends with error details
- SMTP connection issues
- Rate limiting events

Example log entries:

```
INFO: Email verification sent to user@example.com
ERROR: Failed to send email: dial tcp: lookup smtp.example.com: no such host
WARN: Verification email rate limit exceeded for user ID 12345
```

## Next Steps

1. Configure your SMTP credentials in production
2. Set up SPF, DKIM, and DMARC DNS records
3. Test all email templates
4. Monitor email delivery rates
5. Consider using a dedicated email service (SendGrid, Mailgun, etc.) for better deliverability

---

For more information, see the email service implementation in:
- `internal/services/email.go` - Main email service
- `internal/services/email/templates.go` - Email templates
- `internal/services/email/smtp.go` - SMTP provider
- `configs/config.go` - Configuration loading
