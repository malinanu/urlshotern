package email

import (
	"fmt"
)

// ConsoleProvider implements email sending via console output (for development)
type ConsoleProvider struct {
	config EmailConfig
}

// NewConsoleProvider creates a new console provider
func NewConsoleProvider(config EmailConfig) *ConsoleProvider {
	return &ConsoleProvider{
		config: config,
	}
}

// SendEmail outputs the email to console
func (c *ConsoleProvider) SendEmail(to, subject, htmlBody, textBody string) error {
	fmt.Println("========== EMAIL (CONSOLE) ==========")
	fmt.Printf("From: %s <%s>\n", c.config.FromName, c.config.FromEmail)
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Println("-------------------------------------")
	if textBody != "" {
		fmt.Println("TEXT BODY:")
		fmt.Println(textBody)
		fmt.Println("-------------------------------------")
	}
	if htmlBody != "" {
		fmt.Println("HTML BODY:")
		fmt.Println(htmlBody)
	}
	fmt.Println("=====================================")
	return nil
}
