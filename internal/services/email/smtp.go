package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPProvider implements email sending via SMTP
type SMTPProvider struct {
	config EmailConfig
}

// NewSMTPProvider creates a new SMTP provider
func NewSMTPProvider(config EmailConfig) *SMTPProvider {
	return &SMTPProvider{
		config: config,
	}
}

// SendEmail sends an email via SMTP
func (s *SMTPProvider) SendEmail(to, subject, htmlBody, textBody string) error {
	// Build email message
	msg := s.buildMessage(to, subject, htmlBody, textBody)

	// Setup authentication
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// For TLS connections (port 465)
	if s.config.SMTPPort == 465 {
		return s.sendWithTLS(addr, auth, to, msg)
	}

	// For STARTTLS connections (port 587, 25)
	return smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(msg))
}

// sendWithTLS sends email using direct TLS connection
func (s *SMTPProvider) sendWithTLS(addr string, auth smtp.Auth, to string, msg string) error {
	// Create TLS config
	tlsconfig := &tls.Config{
		ServerName: s.config.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}

	// Connect to server
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err = client.Mail(s.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data connection: %w", err)
	}

	return nil
}

// buildMessage builds the email message with MIME multipart
func (s *SMTPProvider) buildMessage(to, subject, htmlBody, textBody string) string {
	fromHeader := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)

	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	// Multipart alternative boundary
	boundary := "----=_Part_0_1234567890.1234567890"
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// Plain text part
	if textBody != "" {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
		msg.WriteString("\r\n\r\n")
	}

	// HTML part
	if htmlBody != "" {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(htmlBody)
		msg.WriteString("\r\n\r\n")
	}

	// End boundary
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return msg.String()
}
