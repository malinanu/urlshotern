package services

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/URLshorter/url-shortener/internal/models"
)

// unfoldVCard reverses RFC 2426 line folding and splits into content lines.
func unfoldVCard(t *testing.T, vcf string) []string {
	t.Helper()
	if !strings.HasSuffix(vcf, "\r\n") {
		t.Fatal("vCard must end with CRLF")
	}
	if strings.Contains(strings.ReplaceAll(vcf, "\r\n", ""), "\n") {
		t.Fatal("bare LF found; all line breaks must be CRLF")
	}
	unfolded := strings.ReplaceAll(vcf, "\r\n ", "")
	return strings.Split(strings.TrimSuffix(unfolded, "\r\n"), "\r\n")
}

func TestBuildVCardFields(t *testing.T) {
	d := models.VCardActionData{
		FirstName: "Sasha", LastName: "Fields", Organization: "Betty Baker, Inc; Ltd", Title: "Brand Director",
		PhoneMobile: "+1 (303) 333-4444", PhoneWork: "303.555.0000", EmailWork: "sasha.f@betty-baker.com",
		Website: "https://betty-baker.com", Street: "1 Main St", City: "Denver", Region: "CO", PostalCode: "80202",
		Country: "USA", Note: "Line one\nLine two \\ done",
	}
	lines := unfoldVCard(t, string(BuildVCard(d, nil)))
	want := []string{
		"BEGIN:VCARD",
		"VERSION:3.0",
		"N:Fields;Sasha;;;",
		"FN:Sasha Fields",
		`ORG:Betty Baker\, Inc\; Ltd`,
		"TITLE:Brand Director",
		"TEL;TYPE=CELL,VOICE:+13033334444",
		"TEL;TYPE=WORK,VOICE:3035550000",
		"EMAIL;TYPE=INTERNET,WORK:sasha.f@betty-baker.com",
		"URL:https://betty-baker.com",
		"ADR;TYPE=WORK:;;1 Main St;Denver;CO;80202;USA",
		`NOTE:Line one\nLine two \\ done`,
		"END:VCARD",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("vCard lines =\n%s\nwant\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

func TestBuildVCardFoldingAndPhoto(t *testing.T) {
	d := models.VCardActionData{FirstName: strings.Repeat("Ä", 60), Note: strings.Repeat("ü", 100)}
	photo := &VCardPhoto{Data: make([]byte, 600), MimeType: "image/jpeg"}
	vcf := string(BuildVCard(d, photo))

	for _, physical := range strings.Split(strings.TrimSuffix(vcf, "\r\n"), "\r\n") {
		if len(physical) > 75 {
			t.Errorf("line exceeds 75 octets (%d): %q", len(physical), physical)
		}
		if !utf8.ValidString(physical) {
			t.Errorf("fold split a UTF-8 character: %q", physical)
		}
	}
	lines := unfoldVCard(t, vcf)
	var gotPhoto bool
	for _, l := range lines {
		if strings.HasPrefix(l, "PHOTO;ENCODING=b;TYPE=JPEG:") {
			gotPhoto = true
		}
		if strings.HasPrefix(l, "FN:") && l != "FN:"+strings.Repeat("Ä", 60) {
			t.Errorf("FN did not round-trip: %q", l)
		}
	}
	if !gotPhoto {
		t.Error("embedded photo missing")
	}
}

func TestBuildVCardOrganizationOnlyAndRemotePhoto(t *testing.T) {
	lines := unfoldVCard(t, string(BuildVCard(models.VCardActionData{Organization: "Acme", PhotoURL: "https://cdn.example.com/a.jpg"}, nil)))
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "FN:Acme") || !strings.Contains(joined, "PHOTO;VALUE=URI:https://cdn.example.com/a.jpg") {
		t.Errorf("unexpected vCard:\n%s", joined)
	}
}

func TestVCardFilename(t *testing.T) {
	cases := map[string]models.VCardActionData{
		"sasha-fields.vcf": {FirstName: "Sasha", LastName: "Fields"},
		"acme-co.vcf":      {Organization: "Acme  Co."},
		"fallback.vcf":     {FirstName: "日本"},
	}
	for want, d := range cases {
		if got := VCardFilename(d, "fallback"); got != want {
			t.Errorf("VCardFilename(%+v) = %q, want %q", d, got, want)
		}
	}
}
