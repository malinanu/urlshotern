package email

import (
	"strings"
	"testing"
	"time"
)

func TestDMCANoticeEscapesSubmittedText(t *testing.T) {
	tpl := DMCANoticeEmailTemplate(DMCANotice{
		ReferenceID:     "DMCA-1",
		Name:            `<script>alert(1)</script>`,
		Email:           "a@example.com",
		Address:         "1 Main St",
		CopyrightedWork: `<img src=x onerror=alert(1)>`,
		InfringingURLs:  "https://trunc.site/abc",
		Signature:       "Jane",
		SubmittedAt:     time.Unix(0, 0),
	})
	if strings.Contains(tpl.HTMLBody, "<script>") || strings.Contains(tpl.HTMLBody, "<img src=x") {
		t.Fatal("submitted text must be HTML-escaped")
	}
	if !strings.Contains(tpl.HTMLBody, FooterMarker) {
		t.Fatal("notice should carry the footer marker")
	}
}

func TestDMCAReceiptDoesNotEchoSubmission(t *testing.T) {
	tpl := DMCAReceiptEmailTemplate("DMCA-1", time.Unix(0, 0))
	if !strings.Contains(tpl.TextBody, "DMCA-1") || !strings.Contains(tpl.HTMLBody, FooterMarker) {
		t.Fatal("receipt should include the reference and footer marker")
	}
}
