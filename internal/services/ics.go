package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

// BuildICS produces an iCalendar (RFC 5545) document with one event. Times
// are converted to UTC so no VTIMEZONE block is needed; calendar apps show
// them in the visitor's own zone. Text uses the same escaping and 75-octet
// folding as vCard.
func BuildICS(d models.EventDetailsData, uid, pageURL string, now time.Time) ([]byte, error) {
	start, err := parseEventTime(d.Start, d.Timezone)
	if err != nil {
		return nil, fmt.Errorf("event start: %w", err)
	}
	end := start.Add(time.Hour)
	if d.End != "" {
		if e, err := parseEventTime(d.End, d.Timezone); err == nil && e.After(start) {
			end = e
		}
	}
	const utc = "20060102T150405Z"
	location := joinNonEmpty(", ", d.Venue, d.Address)

	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//Trunc//Pages//EN",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"UID:" + vcardEscape(uid),
		"DTSTAMP:" + now.UTC().Format(utc),
		"DTSTART:" + start.UTC().Format(utc),
		"DTEND:" + end.UTC().Format(utc),
		"SUMMARY:" + vcardEscape(d.Title),
	}
	if location != "" {
		lines = append(lines, "LOCATION:"+vcardEscape(location))
	}
	if pageURL != "" {
		lines = append(lines, "URL:"+pageURL, "DESCRIPTION:"+vcardEscape(pageURL))
	}
	lines = append(lines, "END:VEVENT", "END:VCALENDAR")

	var b strings.Builder
	for _, l := range lines {
		b.WriteString(vcardFold(l))
		b.WriteString("\r\n")
	}
	return []byte(b.String()), nil
}

// ICSFilename returns a safe ASCII filename for the calendar download.
func ICSFilename(title, fallback string) string {
	name := VCardFilename(models.VCardActionData{Organization: title}, fallback)
	return strings.TrimSuffix(name, ".vcf") + ".ics"
}
