package email

import (
	"fmt"
)

// Provider defines the interface for email providers
type Provider interface {
	SendEmail(to, subject, htmlBody, textBody string) error
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Provider     string // "smtp", "sendgrid", "console"
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// NewProvider creates an email provider based on configuration
func NewProvider(config EmailConfig) (Provider, error) {
	switch config.Provider {
	case "smtp":
		return NewSMTPProvider(config), nil
	case "console":
		return NewConsoleProvider(config), nil
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", config.Provider)
	}
}
