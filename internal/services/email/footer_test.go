package email

import (
	"strings"
	"testing"
)

func TestApplyFooterUsesMarker(t *testing.T) {
	tpl := WelcomeEmailTemplate("Jane", "jane@example.com")
	opts := FooterOptions{SiteURL: "https://trunc.site", UnsubscribeURL: "https://trunc.site/unsubscribe?token=abc"}
	html, text := ApplyFooter(tpl.HTMLBody, tpl.TextBody, opts)

	for _, want := range []string{CompanyName, CompanyPostalAddress, "https://trunc.site/privacy", "https://trunc.site/unsubscribe?token=abc"} {
		if !strings.Contains(html, want) {
			t.Errorf("html footer missing %q", want)
		}
		if !strings.Contains(text, want) {
			t.Errorf("text footer missing %q", want)
		}
	}
	if strings.Contains(html, FooterMarker) {
		t.Error("marker should be replaced")
	}
	if strings.Index(html, CompanyPostalAddress) > strings.Index(html, "</body>") {
		t.Error("footer should sit inside the body")
	}
}

func TestApplyFooterWithoutMarker(t *testing.T) {
	html, _ := ApplyFooter("<div>hello</div>", "hello", FooterOptions{Promotional: true, UnsubscribeURL: "u"})
	if !strings.HasPrefix(html, "<div>hello</div>") || !strings.Contains(html, "promotional message") {
		t.Errorf("footer not appended: %s", html)
	}

	html, _ = ApplyFooter("<html><BODY><p>x</p></BODY></html>", "x", FooterOptions{})
	if !strings.HasSuffix(html, "</BODY></html>") || !strings.Contains(html, CompanyPostalAddress) {
		t.Errorf("footer not placed before </body>: %s", html)
	}
}

func TestEveryTemplateHasFooterMarker(t *testing.T) {
	templates := []EmailTemplate{
		VerificationEmailTemplate("u", "e"),
		PasswordResetEmailTemplate("u", "e"),
		TeamInvitationEmailTemplate("u", "t", "i", "e"),
		ContactFormEmailTemplate("n", "e", "s", "m"),
		SupportInquiryEmailTemplate("n", "e", "s", "i"),
		OTPEmailTemplate("123456", "e", 5),
		WelcomeEmailTemplate("n", "e"),
	}
	for i, tpl := range templates {
		if !strings.Contains(tpl.HTMLBody, FooterMarker) {
			t.Errorf("template %d is missing the footer marker", i)
		}
	}
}
