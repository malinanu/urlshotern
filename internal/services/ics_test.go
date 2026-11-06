package services

import (
	"strings"
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

func TestBuildICS(t *testing.T) {
	d := models.EventDetailsData{
		Title: "Launch night; drinks, demos", Start: "2026-11-20T18:30", End: "2026-11-20T21:00",
		Timezone: "Asia/Colombo", Venue: "Hall A", Address: "1 Main St, Colombo",
	}
	out, err := BuildICS(d, "blk_e-5@trunc.site", "https://trunc.site/m/launch", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	lines := unfoldVCard(t, s)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{
		"BEGIN:VCALENDAR", "VERSION:2.0", "BEGIN:VEVENT",
		"DTSTART:20261120T130000Z", // 18:30 in Colombo (UTC+5:30)
		"DTEND:20261120T153000Z",
		`SUMMARY:Launch night\; drinks\, demos`,
		`LOCATION:Hall A\, 1 Main St\, Colombo`,
		"END:VEVENT", "END:VCALENDAR",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in\n%s", want, joined)
		}
	}
	for _, physical := range strings.Split(strings.TrimSuffix(s, "\r\n"), "\r\n") {
		if len(physical) > 75 {
			t.Errorf("unfolded line > 75 octets: %q", physical)
		}
	}

	d.End = ""
	out, _ = BuildICS(d, "u", "", time.Now())
	if !strings.Contains(string(out), "DTEND:20261120T140000Z") {
		t.Error("missing end should default to one hour after start")
	}
	if _, err := BuildICS(models.EventDetailsData{Start: "nope", Timezone: "UTC"}, "u", "", time.Now()); err == nil {
		t.Error("invalid start should error")
	}
}

func TestEventValidation(t *testing.T) {
	ok := []models.EventDetailsData{
		{Title: "x", Start: "2026-11-20T18:30", Timezone: "Europe/London"},
		{Title: "x", Start: "2026-11-20T18:30:00", Timezone: "America/New_York", End: "2026-11-20T20:00"},
	}
	for _, d := range ok {
		if _, err := normalizeEventDetails(d, "b", false); err != nil {
			t.Errorf("rejected %+v: %v", d, err)
		}
	}
	bad := []models.EventDetailsData{
		{Title: "x", Start: "2026-11-20T18:30", Timezone: "Mars/Olympus"},
		{Title: "x", Start: "2026-11-20T18:30", Timezone: "Local"},
		{Title: "x", Start: "20/11/2026 18:30", Timezone: "UTC"},
		{Title: "x", Start: "2026-11-20T18:30", End: "2026-11-20T17:00", Timezone: "UTC"},
		{Title: "x", Start: "2026-11-20T18:30", Timezone: "UTC", TicketURL: "javascript:alert(1)"},
		{Title: "", Start: "2026-11-20T18:30", Timezone: "UTC"},
	}
	for _, d := range bad {
		if _, err := normalizeEventDetails(d, "b", false); err == nil {
			t.Errorf("accepted %+v", d)
		}
	}
	// Drafts may leave required fields empty but not set invalid ones.
	if _, err := normalizeEventDetails(models.EventDetailsData{}, "b", true); err != nil {
		t.Errorf("empty draft rejected: %v", err)
	}
	if _, err := normalizeEventDetails(models.EventDetailsData{Timezone: "Nope/Nope"}, "b", true); err == nil {
		t.Error("draft with invalid zone accepted")
	}
}

func TestV2BlockValidation(t *testing.T) {
	lat, lng := 6.9271, 79.8612
	badLat := 95.0
	cases := []struct {
		name  string
		typ   string
		data  interface{}
		valid bool
	}{
		{"product ok", "product_card", models.ProductItem{Name: "Mug", Price: "$12", BuyURL: "https://shop.example/mug"}, true},
		{"product js url", "product_card", models.ProductItem{Name: "Mug", BuyURL: "javascript:alert(1)"}, false},
		{"product long price", "product_card", models.ProductItem{Name: "Mug", Price: strings.Repeat("9", 30)}, false},
		{"carousel too many", "product_carousel", models.ProductCarouselData{Items: make([]models.ProductItem, 13)}, false},
		{"map ok", "map", models.MapData{Address: "Colombo", Lat: &lat, Lng: &lng}, true},
		{"map lat only", "map", models.MapData{Address: "x", Lat: &lat}, false},
		{"map out of range", "map", models.MapData{Address: "x", Lat: &badLat, Lng: &lng}, false},
		{"video ok", "video", models.VideoData{URL: "https://youtu.be/dQw4w9WgXcQ"}, true},
		{"video other host", "video", models.VideoData{URL: "https://evil.example/v.mp4"}, false},
		{"audio ok", "audio", models.AudioData{URL: "https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC"}, true},
		{"audio other host", "audio", models.AudioData{URL: "https://soundcloud.com/x"}, false},
		{"form ok", "form", models.FormData{Kind: "contact"}, true},
		{"form bad kind", "form", models.FormData{Kind: "survey"}, false},
		{"countdown ok", "countdown", models.CountdownData{Target: "2027-01-01T00:00", Timezone: "UTC"}, true},
		{"agenda ok", "agenda", models.AgendaData{Items: []models.AgendaItem{{Title: "Keynote", Time: "09:00"}}}, true},
	}
	for _, c := range cases {
		_, err := DecodeMicrositeBlockData(c.typ, mustJSON(t, c.data), "b", false)
		if (err == nil) != c.valid {
			t.Errorf("%s: err=%v, want valid=%v", c.name, err, c.valid)
		}
	}
	// Video URLs are normalized to the canonical watch URL.
	v, _ := DecodeMicrositeBlockData("video", mustJSON(t, models.VideoData{URL: "https://youtu.be/dQw4w9WgXcQ?si=tracking"}), "b", false)
	if v.(models.VideoData).URL != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Errorf("video URL not normalized: %q", v.(models.VideoData).URL)
	}
}
