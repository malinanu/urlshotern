package email

import (
	"fmt"
)

// EmailTemplate represents an email with subject, HTML, and text body
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// VerificationEmailTemplate generates email verification template
func VerificationEmailTemplate(verificationURL, recipientEmail string) EmailTemplate {
	subject := "Verify Your Email - URL Shortener"

	textBody := fmt.Sprintf(`Hello,

Thank you for signing up for URL Shortener!

Please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you didn't create this account, you can safely ignore this email.

Best regards,
URL Shortener Team
`, verificationURL)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 40px 20px; text-align: center;">
                            <h1 style="color: #ffffff; margin: 0; font-size: 28px;">URL Shortener</h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="color: #333333; margin: 0 0 20px; font-size: 24px;">Verify Your Email Address</h2>

                            <p style="color: #666666; line-height: 1.6; margin: 0 0 20px; font-size: 16px;">
                                Thank you for signing up! Please verify your email address by clicking the button below:
                            </p>

                            <table cellpadding="0" cellspacing="0" style="margin: 30px 0;">
                                <tr>
                                    <td align="center" style="border-radius: 4px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);">
                                        <a href="%s" style="display: inline-block; padding: 14px 40px; color: #ffffff; text-decoration: none; font-weight: bold; font-size: 16px;">
                                            Verify Email Address
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="color: #666666; line-height: 1.6; margin: 20px 0; font-size: 14px;">
                                Or copy and paste this link into your browser:
                            </p>

                            <p style="color: #667eea; line-height: 1.6; margin: 0 0 20px; font-size: 14px; word-break: break-all;">
                                %s
                            </p>

                            <p style="color: #999999; line-height: 1.6; margin: 30px 0 0; font-size: 14px;">
                                This link will expire in 24 hours. If you didn't create this account, you can safely ignore this email.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f9fa; padding: 20px 30px; text-align: center; border-top: 1px solid #e9ecef;">
                            <p style="color: #999999; margin: 0; font-size: 12px;">
                                &copy; 2024 URL Shortener. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`, verificationURL, verificationURL)

	return EmailTemplate{
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}
}

// PasswordResetEmailTemplate generates password reset template
func PasswordResetEmailTemplate(resetURL, recipientEmail string) EmailTemplate {
	subject := "Reset Your Password - URL Shortener"

	textBody := fmt.Sprintf(`Hello,

We received a request to reset your password for your URL Shortener account.

Click the link below to reset your password:

%s

This link will expire in 1 hour.

If you didn't request a password reset, you can safely ignore this email. Your password will remain unchanged.

Best regards,
URL Shortener Team
`, resetURL)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); padding: 40px 20px; text-align: center;">
                            <h1 style="color: #ffffff; margin: 0; font-size: 28px;">URL Shortener</h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="color: #333333; margin: 0 0 20px; font-size: 24px;">Reset Your Password</h2>

                            <p style="color: #666666; line-height: 1.6; margin: 0 0 20px; font-size: 16px;">
                                We received a request to reset your password. Click the button below to create a new password:
                            </p>

                            <table cellpadding="0" cellspacing="0" style="margin: 30px 0;">
                                <tr>
                                    <td align="center" style="border-radius: 4px; background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%);">
                                        <a href="%s" style="display: inline-block; padding: 14px 40px; color: #ffffff; text-decoration: none; font-weight: bold; font-size: 16px;">
                                            Reset Password
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="color: #666666; line-height: 1.6; margin: 20px 0; font-size: 14px;">
                                Or copy and paste this link into your browser:
                            </p>

                            <p style="color: #f5576c; line-height: 1.6; margin: 0 0 20px; font-size: 14px; word-break: break-all;">
                                %s
                            </p>

                            <div style="background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 30px 0;">
                                <p style="color: #856404; margin: 0; font-size: 14px; line-height: 1.6;">
                                    <strong>Security Notice:</strong> This link will expire in 1 hour. If you didn't request a password reset, please ignore this email or contact support if you're concerned about your account security.
                                </p>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f9fa; padding: 20px 30px; text-align: center; border-top: 1px solid #e9ecef;">
                            <p style="color: #999999; margin: 0; font-size: 12px;">
                                &copy; 2024 URL Shortener. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`, resetURL, resetURL)

	return EmailTemplate{
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}
}

// TeamInvitationEmailTemplate generates team invitation template
func TeamInvitationEmailTemplate(inviteURL, teamName, inviterName, recipientEmail string) EmailTemplate {
	subject := fmt.Sprintf("You've been invited to join %s on URL Shortener", teamName)

	textBody := fmt.Sprintf(`Hello,

%s has invited you to join the team "%s" on URL Shortener!

Click the link below to accept the invitation:

%s

This invitation will expire in 7 days.

If you don't want to accept this invitation, you can safely ignore this email.

Best regards,
URL Shortener Team
`, inviterName, teamName, inviteURL)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Team Invitation</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f4f4f4; padding: 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #4facfe 0%%, #00f2fe 100%%); padding: 40px 20px; text-align: center;">
                            <h1 style="color: #ffffff; margin: 0; font-size: 28px;">URL Shortener</h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="color: #333333; margin: 0 0 20px; font-size: 24px;">Team Invitation</h2>

                            <p style="color: #666666; line-height: 1.6; margin: 0 0 20px; font-size: 16px;">
                                <strong>%s</strong> has invited you to join the team <strong>"%s"</strong> on URL Shortener!
                            </p>

                            <table cellpadding="0" cellspacing="0" style="margin: 30px 0;">
                                <tr>
                                    <td align="center" style="border-radius: 4px; background: linear-gradient(135deg, #4facfe 0%%, #00f2fe 100%%);">
                                        <a href="%s" style="display: inline-block; padding: 14px 40px; color: #ffffff; text-decoration: none; font-weight: bold; font-size: 16px;">
                                            Accept Invitation
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="color: #666666; line-height: 1.6; margin: 20px 0; font-size: 14px;">
                                Or copy and paste this link into your browser:
                            </p>

                            <p style="color: #4facfe; line-height: 1.6; margin: 0 0 20px; font-size: 14px; word-break: break-all;">
                                %s
                            </p>

                            <p style="color: #999999; line-height: 1.6; margin: 30px 0 0; font-size: 14px;">
                                This invitation will expire in 7 days. If you don't want to accept this invitation, you can safely ignore this email.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f8f9fa; padding: 20px 30px; text-align: center; border-top: 1px solid #e9ecef;">
                            <p style="color: #999999; margin: 0; font-size: 12px;">
                                &copy; 2024 URL Shortener. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`, inviterName, teamName, inviteURL, inviteURL)

	return EmailTemplate{
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}
}
