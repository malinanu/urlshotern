package services

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/URLshorter/url-shortener/internal/models"
)

// VCardPhoto is an embedded contact photo.
type VCardPhoto struct {
	Data     []byte
	MimeType string // image/jpeg | image/png
}

// BuildVCard produces a vCard 3.0 (RFC 2426) document. Text values are escaped,
// lines use CRLF and are folded at 75 octets without splitting UTF-8
// characters. photo, if non-nil, is embedded inline (preferred by iOS and
// Android); otherwise a remote https photo URL is referenced.
func BuildVCard(d models.VCardActionData, photo *VCardPhoto) []byte {
	var lines []string
	add := func(line string) { lines = append(lines, line) }

	add("BEGIN:VCARD")
	add("VERSION:3.0")

	fn := strings.TrimSpace(strings.Join([]string{d.FirstName, d.LastName}, " "))
	if fn == "" {
		fn = d.Organization
	}
	add("N:" + vcardEscape(d.LastName) + ";" + vcardEscape(d.FirstName) + ";;;")
	add("FN:" + vcardEscape(fn))
	if d.Organization != "" {
		add("ORG:" + vcardEscape(d.Organization))
	}
	if d.Title != "" {
		add("TITLE:" + vcardEscape(d.Title))
	}
	if p := telValue(d.PhoneMobile); p != "" {
		add("TEL;TYPE=CELL,VOICE:" + p)
	}
	if p := telValue(d.PhoneWork); p != "" {
		add("TEL;TYPE=WORK,VOICE:" + p)
	}
	if d.EmailWork != "" {
		add("EMAIL;TYPE=INTERNET,WORK:" + vcardEscape(d.EmailWork))
	}
	if d.EmailPersonal != "" {
		add("EMAIL;TYPE=INTERNET,HOME:" + vcardEscape(d.EmailPersonal))
	}
	if d.Website != "" {
		add("URL:" + vcardEscape(d.Website))
	}
	if d.Street != "" || d.City != "" || d.Region != "" || d.PostalCode != "" || d.Country != "" {
		add("ADR;TYPE=WORK:;;" + strings.Join([]string{
			vcardEscape(d.Street), vcardEscape(d.City), vcardEscape(d.Region),
			vcardEscape(d.PostalCode), vcardEscape(d.Country),
		}, ";"))
	}
	if d.Note != "" {
		add("NOTE:" + vcardEscape(d.Note))
	}
	switch {
	case photo != nil && len(photo.Data) > 0:
		typ := "JPEG"
		if photo.MimeType == "image/png" {
			typ = "PNG"
		}
		add("PHOTO;ENCODING=b;TYPE=" + typ + ":" + base64.StdEncoding.EncodeToString(photo.Data))
	case strings.HasPrefix(d.PhotoURL, "https://"):
		add("PHOTO;VALUE=URI:" + d.PhotoURL)
	}
	add("END:VCARD")

	var b strings.Builder
	for _, l := range lines {
		b.WriteString(vcardFold(l))
		b.WriteString("\r\n")
	}
	return []byte(b.String())
}

// vcardEscape escapes a text value per RFC 2426 section 4: backslash, comma,
// semicolon and newlines.
func vcardEscape(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case ',':
			b.WriteString(`\,`)
		case ';':
			b.WriteString(`\;`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// vcardFold folds a content line so no physical line exceeds 75 octets.
// Continuation lines start with a single space. Multi-byte characters are
// never split.
func vcardFold(line string) string {
	const limit = 75
	if len(line) <= limit {
		return line
	}
	var b strings.Builder
	width := 0
	lineMax := limit
	for _, r := range line {
		n := utf8.RuneLen(r)
		if width+n > lineMax {
			b.WriteString("\r\n ")
			width = 0
			lineMax = limit - 1 // the leading space counts toward the 75 octets
		}
		b.WriteRune(r)
		width += n
	}
	return b.String()
}

// VCardFilename returns a safe ASCII filename for the download.
func VCardFilename(d models.VCardActionData, fallback string) string {
	name := strings.TrimSpace(d.FirstName + " " + d.LastName)
	if name == "" {
		name = d.Organization
	}
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
				b.WriteRune('-')
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = fallback
	}
	if len(out) > 60 {
		out = out[:60]
	}
	return out + ".vcf"
}
