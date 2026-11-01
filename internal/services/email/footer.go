package email

import (
	"fmt"
	"html"
	"strings"
)

// Sender identification required in every commercial email (CAN-SPAM 15 U.S.C. §7704(a)(5)).
const (
	CompanyName          = "3logiq Technologies, Inc."
	CompanyPostalAddress = "123 Tech Boulevard, San Francisco, CA 94105, USA"
	DefaultSiteURL       = "https://trunc.site"
	// DMCAAgentEmail is the Designated Agent's address for copyright notices (17 U.S.C. §512(c)(2)).
	DMCAAgentEmail = "dmca@trunc.site"
)

// FooterMarker marks where a template wants the compliance footer. Bodies without
// it get the footer before </body> (HTML) or at the end (text).
const FooterMarker = "<!--EMAIL_FOOTER-->"

// FooterOptions describes the footer for one recipient.
type FooterOptions struct {
	SiteURL        string // e.g. https://trunc.site
	UnsubscribeURL string // per-recipient one-click unsubscribe link
	Promotional    bool   // marketing / campaign message
}

func (o FooterOptions) privacyURL() string {
	site := strings.TrimRight(o.SiteURL, "/")
	if site == "" {
		site = DefaultSiteURL
	}
	return site + "/privacy"
}

// FooterHTML renders the footer block: postal address, privacy policy and unsubscribe link.
func FooterHTML(o FooterOptions) string {
	var b strings.Builder
	b.WriteString(`<div style="color: #999999; margin: 0; font-size: 12px; line-height: 1.6; text-align: center;">`)
	if o.Promotional {
		b.WriteString(`<p style="margin: 0 0 8px;">This is a promotional message from Trunc.</p>`)
	}
	fmt.Fprintf(&b, `<p style="margin: 0 0 8px;">%s &middot; %s</p>`,
		html.EscapeString(CompanyName), html.EscapeString(CompanyPostalAddress))
	fmt.Fprintf(&b, `<p style="margin: 0;"><a href="%s" style="color: #999999; text-decoration: underline;">Privacy Policy</a>`,
		html.EscapeString(o.privacyURL()))
	if o.UnsubscribeURL != "" {
		fmt.Fprintf(&b, ` &middot; <a href="%s" style="color: #999999; text-decoration: underline;">Unsubscribe from marketing emails</a>`,
			html.EscapeString(o.UnsubscribeURL))
	}
	b.WriteString(`</p></div>`)
	return b.String()
}

// FooterText renders the plain-text version of the footer.
func FooterText(o FooterOptions) string {
	var b strings.Builder
	b.WriteString("--\n")
	if o.Promotional {
		b.WriteString("This is a promotional message from Trunc.\n")
	}
	fmt.Fprintf(&b, "%s, %s\n", CompanyName, CompanyPostalAddress)
	fmt.Fprintf(&b, "Privacy Policy: %s\n", o.privacyURL())
	if o.UnsubscribeURL != "" {
		fmt.Fprintf(&b, "Unsubscribe from marketing emails: %s\n", o.UnsubscribeURL)
	}
	return b.String()
}

// ApplyFooter adds the footer to an HTML/text body pair.
func ApplyFooter(htmlBody, textBody string, o FooterOptions) (string, string) {
	footer := FooterHTML(o)
	switch {
	case strings.Contains(htmlBody, FooterMarker):
		htmlBody = strings.Replace(htmlBody, FooterMarker, footer, 1)
	case strings.Contains(strings.ToLower(htmlBody), "</body>"):
		i := strings.LastIndex(strings.ToLower(htmlBody), "</body>")
		htmlBody = htmlBody[:i] + footer + htmlBody[i:]
	default:
		htmlBody += footer
	}
	return htmlBody, strings.TrimRight(textBody, "\n") + "\n\n" + FooterText(o)
}
