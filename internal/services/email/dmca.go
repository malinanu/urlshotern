package email

import (
	"fmt"
	"html"
	"strings"
	"time"
)

// DMCANotice is a copyright takedown notice submitted through /dmca.
type DMCANotice struct {
	ReferenceID     string
	Name            string
	Company         string
	Role            string // "owner" or "agent"
	Email           string
	Phone           string
	Address         string
	CopyrightedWork string
	OriginalURL     string
	InfringingURLs  string
	Signature       string
	SubmittedAt     time.Time
	IPAddress       string
}

func (n DMCANotice) roleLabel() string {
	if n.Role == "agent" {
		return "Authorized agent of the copyright owner"
	}
	return "Copyright owner"
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(not provided)"
	}
	return s
}

// DMCANoticeEmailTemplate formats a takedown notice for the Designated Agent.
func DMCANoticeEmailTemplate(n DMCANotice) EmailTemplate {
	subject := fmt.Sprintf("[DMCA] Takedown notice %s from %s", n.ReferenceID, n.Name)
	submitted := n.SubmittedAt.UTC().Format(time.RFC1123)

	fields := [][2]string{
		{"Reference", n.ReferenceID},
		{"Submitted", submitted},
		{"Submitter IP", orNone(n.IPAddress)},
		{"Complaining party", n.Name},
		{"Company / organization", orNone(n.Company)},
		{"Capacity", n.roleLabel()},
		{"Email", n.Email},
		{"Phone", orNone(n.Phone)},
		{"Postal address", n.Address},
		{"Copyrighted work", n.CopyrightedWork},
		{"Original work location", orNone(n.OriginalURL)},
		{"Allegedly infringing trunc.site URLs", n.InfringingURLs},
		{"Good-faith statement", "Affirmed: use is not authorized by the copyright owner, its agent, or the law."},
		{"Accuracy statement", "Affirmed under penalty of perjury: information is accurate and submitter is the owner or authorized to act for the owner."},
		{"Electronic signature", n.Signature},
	}

	var text, rows strings.Builder
	text.WriteString("A DMCA takedown notice was submitted through trunc.site/dmca.\n\n")
	for _, f := range fields {
		fmt.Fprintf(&text, "%s:\n%s\n\n", f[0], f[1])
		fmt.Fprintf(&rows, `<tr><td style="padding: 8px 12px; border-bottom: 1px solid #e9ecef; color: #666666; font-size: 13px; vertical-align: top; width: 35%%;">%s</td><td style="padding: 8px 12px; border-bottom: 1px solid #e9ecef; color: #333333; font-size: 14px; white-space: pre-wrap;">%s</td></tr>`,
			html.EscapeString(f[0]), html.EscapeString(f[1]))
	}
	text.WriteString("Review the links, disable infringing ones expeditiously, notify the link owner, and record the strike under the repeat-infringer policy.\n")

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>DMCA takedown notice</title></head>
<body style="margin: 0; padding: 20px; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table width="640" cellpadding="0" cellspacing="0" style="margin: 0 auto; background-color: #ffffff; border-radius: 8px;">
        <tr><td style="padding: 24px 24px 8px;"><h2 style="margin: 0; color: #333333;">DMCA takedown notice %s</h2>
        <p style="color: #555555; font-size: 14px;">Submitted through trunc.site/dmca. Review the links, disable infringing ones expeditiously, notify the link owner, and record the strike under the repeat-infringer policy.</p></td></tr>
        <tr><td style="padding: 0 24px 24px;"><table width="100%%" cellpadding="0" cellspacing="0">%s</table></td></tr>
        <tr><td style="padding: 0 24px 24px;">%s</td></tr>
    </table>
</body>
</html>`, html.EscapeString(n.ReferenceID), rows.String(), FooterMarker)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: text.String()}
}

// DMCAReceiptEmailTemplate confirms receipt to the complaining party. It deliberately
// doesn't echo the submitted text, so the public form can't be used to relay
// arbitrary content to arbitrary inboxes.
func DMCAReceiptEmailTemplate(referenceID string, submittedAt time.Time) EmailTemplate {
	subject := fmt.Sprintf("We received your DMCA notice (%s)", referenceID)
	submitted := submittedAt.UTC().Format(time.RFC1123)

	textBody := fmt.Sprintf(`Hello,

We received a DMCA takedown notice submitted with this email address on %s.
Reference: %s

Our Designated Agent will review it and may contact you at this address if more information is needed.
If you didn't submit this notice, you can ignore this email.

trunc.site Copyright Agent
%s
`, submitted, referenceID, DMCAAgentEmail)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>DMCA notice received</title></head>
<body style="margin: 0; padding: 20px; font-family: Arial, sans-serif; background-color: #f4f4f4;">
    <table width="600" cellpadding="0" cellspacing="0" style="margin: 0 auto; background-color: #ffffff; border-radius: 8px;">
        <tr><td style="padding: 30px;">
            <h2 style="margin-top: 0; color: #333333;">We received your DMCA notice</h2>
            <p style="color: #555555; line-height: 1.6;">We received a DMCA takedown notice submitted with this email address on %s.</p>
            <p style="color: #555555; line-height: 1.6;"><strong>Reference:</strong> %s</p>
            <p style="color: #555555; line-height: 1.6;">Our Designated Agent will review it and may contact you at this address if more information is needed. If you didn't submit this notice, you can ignore this email.</p>
            <p style="color: #555555; line-height: 1.6;">trunc.site Copyright Agent<br>%s</p>
        </td></tr>
        <tr><td style="padding: 0 30px 24px;">%s</td></tr>
    </table>
</body>
</html>`, html.EscapeString(submitted), html.EscapeString(referenceID), DMCAAgentEmail, FooterMarker)

	return EmailTemplate{Subject: subject, HTMLBody: htmlBody, TextBody: textBody}
}
