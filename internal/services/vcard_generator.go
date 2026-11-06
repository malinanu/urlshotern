package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

// EscapeVCardValue escapes characters per RFC 2426 specification
func EscapeVCardValue(val string) string {
	val = strings.ReplaceAll(val, `\`, `\\`)
	val = strings.ReplaceAll(val, `,`, `\,`)
	val = strings.ReplaceAll(val, `;`, `\;`)
	val = strings.ReplaceAll(val, "\r\n", `\n`)
	val = strings.ReplaceAll(val, "\r", `\n`)
	val = strings.ReplaceAll(val, "\n", `\n`)
	return strings.TrimSpace(val)
}

// GenerateVCard generates an RFC 2426 (vCard 3.0) compliant string
func GenerateVCard(c *models.VCardContact) string {
	if c == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("BEGIN:VCARD\r\n")
	sb.WriteString("VERSION:3.0\r\n")

	// Full name and structured name
	fn := strings.TrimSpace(fmt.Sprintf("%s %s", c.FirstName, c.LastName))
	if fn == "" {
		fn = c.Organization
	}
	if fn != "" {
		sb.WriteString(fmt.Sprintf("FN:%s\r\n", EscapeVCardValue(fn)))
	}
	if c.LastName != "" || c.FirstName != "" {
		sb.WriteString(fmt.Sprintf("N:%s;%s;;;\r\n", EscapeVCardValue(c.LastName), EscapeVCardValue(c.FirstName)))
	} else if c.Organization != "" {
		sb.WriteString(fmt.Sprintf("N:%s;;;;\r\n", EscapeVCardValue(c.Organization)))
	} else {
		sb.WriteString("N:;;;;\r\n")
	}

	// Organization & Title
	if c.Organization != "" {
		sb.WriteString(fmt.Sprintf("ORG:%s\r\n", EscapeVCardValue(c.Organization)))
	}
	if c.Title != "" {
		sb.WriteString(fmt.Sprintf("TITLE:%s\r\n", EscapeVCardValue(c.Title)))
	}

	// Telephones
	if c.Phone != "" {
		sb.WriteString(fmt.Sprintf("TEL;TYPE=CELL,VOICE:%s\r\n", EscapeVCardValue(c.Phone)))
	}
	if c.CellPhone != "" && c.CellPhone != c.Phone {
		sb.WriteString(fmt.Sprintf("TEL;TYPE=CELL,VOICE:%s\r\n", EscapeVCardValue(c.CellPhone)))
	}
	if c.WorkPhone != "" {
		sb.WriteString(fmt.Sprintf("TEL;TYPE=WORK,VOICE:%s\r\n", EscapeVCardValue(c.WorkPhone)))
	}

	// Email
	if c.Email != "" {
		sb.WriteString(fmt.Sprintf("EMAIL;TYPE=INTERNET,PREF:%s\r\n", EscapeVCardValue(c.Email)))
	}

	// Website / URL
	if c.Website != "" {
		sb.WriteString(fmt.Sprintf("URL:%s\r\n", EscapeVCardValue(c.Website)))
	}

	// Address: ADR;TYPE=WORK:;;Street;City;State;PostalCode;Country
	if c.Address != "" || c.City != "" || c.State != "" || c.PostalCode != "" || c.Country != "" {
		sb.WriteString(fmt.Sprintf("ADR;TYPE=WORK:;;%s;%s;%s;%s;%s\r\n",
			EscapeVCardValue(c.Address),
			EscapeVCardValue(c.City),
			EscapeVCardValue(c.State),
			EscapeVCardValue(c.PostalCode),
			EscapeVCardValue(c.Country),
		))
	}

	// Note
	if c.Note != "" {
		sb.WriteString(fmt.Sprintf("NOTE:%s\r\n", EscapeVCardValue(c.Note)))
	}

	// Revision timestamp
	sb.WriteString(fmt.Sprintf("REV:%s\r\n", time.Now().UTC().Format("20060102T150405Z")))
	sb.WriteString("END:VCARD\r\n")

	return sb.String()
}
