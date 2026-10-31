package services

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	_ "time/tzdata" // IANA zones for event times, even on minimal container images

	"github.com/URLshorter/url-shortener/internal/models"
)

// Template styles. classic is the v1 look and renders byte-for-byte as before.
const (
	MicrositeStyleClassic     = "classic"
	MicrositeStyleLetterpress = "letterpress"
	MicrositeStyleBoard       = "board"
	MicrositeStyleFlyer       = "flyer"
	MicrositeStyleProgramme   = "programme"
)

// MicrositeStyles is the allowlist of template styles.
var MicrositeStyles = map[string]bool{
	MicrositeStyleClassic: true, MicrositeStyleLetterpress: true, MicrositeStyleBoard: true,
	MicrositeStyleFlyer: true, MicrositeStyleProgramme: true,
}

// MicrositeFontFace is one self-hosted font file.
type MicrositeFontFace struct {
	Family string
	File   string // under /m-static/fonts/
	Weight string // CSS font-weight value or range
}

// MicrositeFontFaces lists the files each self-hosted font key needs.
var MicrositeFontFaces = map[string][]MicrositeFontFace{
	"schibsted": {{Family: "Schibsted Grotesk", File: "schibsted-grotesk.woff2", Weight: "400 900"}},
	"martian":   {{Family: "Martian Mono", File: "martian-mono.woff2", Weight: "100 800"}},
	"barlow": {
		{Family: "Barlow Condensed", File: "barlow-condensed-600.woff2", Weight: "600"},
		{Family: "Barlow Condensed", File: "barlow-condensed-800.woff2", Weight: "800"},
	},
	"hanken": {{Family: "Hanken Grotesk", File: "hanken-grotesk.woff2", Weight: "100 900"}},
}

const (
	eventTimeLayout    = "2006-01-02T15:04"
	maxProductItems    = 12
	maxAgendaItems     = 40
	defaultMapZoom     = 15
	maxPriceRunes      = 24
	formKindNewsletter = "newsletter"
	formKindContact    = "contact"
)

// decodeV2BlockData handles the v2 block types. ok is false for unknown types.
func decodeV2BlockData(blockType string, raw json.RawMessage, field string, draft bool) (interface{}, bool, error) {
	switch blockType {
	case models.BlockTypeProductCard:
		var d models.ProductItem
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid product data")
		}
		d.ID = ""
		v, err := normalizeProduct(d, field, draft)
		return v, true, err
	case models.BlockTypeProductCarousel:
		var d models.ProductCarouselData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid product carousel data")
		}
		v, err := normalizeProductCarousel(d, field, draft)
		return v, true, err
	case models.BlockTypeEventDetails:
		var d models.EventDetailsData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid event data")
		}
		v, err := normalizeEventDetails(d, field, draft)
		return v, true, err
	case models.BlockTypeCountdown:
		var d models.CountdownData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid countdown data")
		}
		v, err := normalizeCountdown(d, field, draft)
		return v, true, err
	case models.BlockTypeAgenda:
		var d models.AgendaData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid agenda data")
		}
		v, err := normalizeAgenda(d, field, draft)
		return v, true, err
	case models.BlockTypeMap:
		var d models.MapData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid map data")
		}
		v, err := normalizeMap(d, field, draft)
		return v, true, err
	case models.BlockTypeVideo:
		var d models.VideoData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid video data")
		}
		v, err := normalizeVideo(d, field, draft)
		return v, true, err
	case models.BlockTypeAudio:
		var d models.AudioData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid audio data")
		}
		v, err := normalizeAudio(d, field, draft)
		return v, true, err
	case models.BlockTypeForm:
		var d models.FormData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, true, validationErr(field, "invalid form data")
		}
		v, err := normalizeForm(d, field)
		return v, true, err
	}
	return nil, false, nil
}

func normalizeProduct(d models.ProductItem, field string, draft bool) (models.ProductItem, error) {
	var err error
	texts := []struct {
		v     *string
		max   int
		multi bool
		name  string
	}{
		{&d.Name, 80, false, "name"},
		{&d.Price, maxPriceRunes, false, "price"},
		{&d.ComparePrice, maxPriceRunes, false, "compare_price"},
		{&d.Badge, 20, false, "badge"},
		{&d.Description, 300, true, "description"},
		{&d.ButtonText, 30, false, "button_text"},
	}
	for _, t := range texts {
		if *t.v, err = cleanText(*t.v, t.max, t.multi, field+"."+t.name); err != nil {
			return d, err
		}
	}
	if d.Name == "" && !draft {
		return d, validationErr(field+".name", "is required")
	}
	if d.ImageURL, err = SanitizeMicrositeImageURL(d.ImageURL, field+".image_url"); err != nil {
		return d, err
	}
	if d.BuyURL, err = optionalURL(d.BuyURL, true, field+".buy_url", SanitizeMicrositeLinkURL); err != nil {
		return d, err
	}
	return d, nil
}

func normalizeProductCarousel(d models.ProductCarouselData, field string, draft bool) (models.ProductCarouselData, error) {
	var err error
	if d.Title, err = cleanText(d.Title, 60, false, field+".title"); err != nil {
		return d, err
	}
	if len(d.Items) > maxProductItems {
		return d, validationErr(field+".items", "at most %d products per strip", maxProductItems)
	}
	seen := map[string]bool{}
	out := make([]models.ProductItem, 0, len(d.Items))
	for i, it := range d.Items {
		f := fmt.Sprintf("%s.items[%d]", field, i)
		if it.ID, err = normalizeItemID(it.ID, seen, f); err != nil {
			return d, err
		}
		id := it.ID
		if it, err = normalizeProduct(it, f, draft); err != nil {
			return d, err
		}
		it.ID = id
		out = append(out, it)
	}
	d.Items = out
	return d, nil
}

// parseEventTime parses a wall-clock time in an IANA zone.
func parseEventTime(value, tz string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" || strings.EqualFold(tz, "local") {
		return time.Time{}, fmt.Errorf("unknown time zone")
	}
	return time.ParseInLocation(eventTimeLayout, value, loc)
}

func normalizeTimeZone(tz, field string, required bool) (string, error) {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		if required {
			return "", validationErr(field, "is required")
		}
		return "", nil
	}
	if strings.EqualFold(tz, "local") || len(tz) > 64 {
		return "", validationErr(field, "must be a time zone like Asia/Colombo")
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return "", validationErr(field, "must be a time zone like Asia/Colombo")
	}
	return tz, nil
}

func normalizeWallTime(v, tz, field string, required bool) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		if required {
			return "", validationErr(field, "is required")
		}
		return "", nil
	}
	if len(v) == len("2006-01-02T15:04:05") {
		v = v[:len(eventTimeLayout)]
	}
	if tz == "" {
		tz = "UTC"
	}
	if _, err := parseEventTime(v, tz); err != nil {
		return "", validationErr(field, "must be a date and time like 2026-11-20T18:30")
	}
	return v, nil
}

func normalizeEventDetails(d models.EventDetailsData, field string, draft bool) (models.EventDetailsData, error) {
	var err error
	texts := []struct {
		v    *string
		max  int
		name string
	}{
		{&d.Title, 120, "title"}, {&d.Venue, 120, "venue"}, {&d.Address, 200, "address"}, {&d.TicketText, 30, "ticket_text"},
	}
	for _, t := range texts {
		if *t.v, err = cleanText(*t.v, t.max, false, field+"."+t.name); err != nil {
			return d, err
		}
	}
	if d.Title == "" && !draft {
		return d, validationErr(field+".title", "is required")
	}
	if d.Timezone, err = normalizeTimeZone(d.Timezone, field+".timezone", !draft); err != nil {
		return d, err
	}
	if d.Start, err = normalizeWallTime(d.Start, d.Timezone, field+".start", !draft); err != nil {
		return d, err
	}
	if d.End, err = normalizeWallTime(d.End, d.Timezone, field+".end", false); err != nil {
		return d, err
	}
	if d.Start != "" && d.End != "" && d.Timezone != "" {
		s, _ := parseEventTime(d.Start, d.Timezone)
		e, _ := parseEventTime(d.End, d.Timezone)
		if !e.After(s) {
			return d, validationErr(field+".end", "must be after the start")
		}
	}
	if d.TicketURL, err = optionalURL(d.TicketURL, true, field+".ticket_url", SanitizeMicrositeLinkURL); err != nil {
		return d, err
	}
	return d, nil
}

func normalizeCountdown(d models.CountdownData, field string, draft bool) (models.CountdownData, error) {
	var err error
	if d.Label, err = cleanText(d.Label, 40, false, field+".label"); err != nil {
		return d, err
	}
	if d.DoneText, err = cleanText(d.DoneText, 60, false, field+".done_text"); err != nil {
		return d, err
	}
	if d.Timezone, err = normalizeTimeZone(d.Timezone, field+".timezone", !draft); err != nil {
		return d, err
	}
	if d.Target, err = normalizeWallTime(d.Target, d.Timezone, field+".target", !draft); err != nil {
		return d, err
	}
	return d, nil
}

func normalizeAgenda(d models.AgendaData, field string, draft bool) (models.AgendaData, error) {
	var err error
	if d.Title, err = cleanText(d.Title, 60, false, field+".title"); err != nil {
		return d, err
	}
	if len(d.Items) > maxAgendaItems {
		return d, validationErr(field+".items", "at most %d sessions", maxAgendaItems)
	}
	seen := map[string]bool{}
	out := make([]models.AgendaItem, 0, len(d.Items))
	for i, it := range d.Items {
		f := fmt.Sprintf("%s.items[%d]", field, i)
		if it.ID, err = normalizeItemID(it.ID, seen, f); err != nil {
			return d, err
		}
		texts := []struct {
			v     *string
			max   int
			multi bool
			name  string
		}{
			{&it.Time, 20, false, "time"}, {&it.Title, 100, false, "title"}, {&it.Speaker, 80, false, "speaker"},
			{&it.Track, 30, false, "track"}, {&it.Detail, 500, true, "detail"},
		}
		for _, t := range texts {
			if *t.v, err = cleanText(*t.v, t.max, t.multi, f+"."+t.name); err != nil {
				return d, err
			}
		}
		if it.Title == "" && !draft {
			return d, validationErr(f+".title", "is required")
		}
		out = append(out, it)
	}
	d.Items = out
	return d, nil
}

func normalizeMap(d models.MapData, field string, draft bool) (models.MapData, error) {
	var err error
	if d.Label, err = cleanText(d.Label, 80, false, field+".label"); err != nil {
		return d, err
	}
	if d.Address, err = cleanText(d.Address, 200, false, field+".address"); err != nil {
		return d, err
	}
	if d.Address == "" && d.Lat == nil && !draft {
		return d, validationErr(field+".address", "is required")
	}
	if (d.Lat == nil) != (d.Lng == nil) {
		return d, validationErr(field+".lat", "latitude and longitude go together")
	}
	if d.Lat != nil {
		lat, lng := *d.Lat, *d.Lng
		if math.IsNaN(lat) || math.IsNaN(lng) || lat < -85 || lat > 85 || lng < -180 || lng > 180 {
			return d, validationErr(field+".lat", "coordinates are out of range")
		}
		lat, lng = math.Round(lat*1e6)/1e6, math.Round(lng*1e6)/1e6
		d.Lat, d.Lng = &lat, &lng
	}
	if d.Zoom == 0 {
		d.Zoom = defaultMapZoom
	}
	if d.Zoom < 3 || d.Zoom > 19 {
		return d, validationErr(field+".zoom", "must be between 3 and 19")
	}
	return d, nil
}

func normalizeVideo(d models.VideoData, field string, draft bool) (models.VideoData, error) {
	var err error
	if d.Title, err = cleanText(d.Title, 100, false, field+".title"); err != nil {
		return d, err
	}
	if d.PosterURL, err = SanitizeMicrositeImageURL(d.PosterURL, field+".poster_url"); err != nil {
		return d, err
	}
	d.URL = strings.TrimSpace(d.URL)
	if d.URL == "" {
		if draft {
			return d, nil
		}
		return d, validationErr(field+".url", "is required")
	}
	ref, perr := ParseVideoURL(d.URL)
	if perr != nil {
		return d, validationErr(field+".url", "%s", perr.Error())
	}
	d.URL = ref.WatchURL()
	return d, nil
}

func normalizeAudio(d models.AudioData, field string, draft bool) (models.AudioData, error) {
	var err error
	if d.Title, err = cleanText(d.Title, 100, false, field+".title"); err != nil {
		return d, err
	}
	d.URL = strings.TrimSpace(d.URL)
	if d.URL == "" {
		if draft {
			return d, nil
		}
		return d, validationErr(field+".url", "is required")
	}
	ref, perr := ParseAudioURL(d.URL)
	if perr != nil {
		return d, validationErr(field+".url", "%s", perr.Error())
	}
	d.URL = ref.WatchURL()
	return d, nil
}

func normalizeForm(d models.FormData, field string) (models.FormData, error) {
	if d.Kind == "" {
		d.Kind = formKindNewsletter
	}
	if d.Kind != formKindNewsletter && d.Kind != formKindContact {
		return d, validationErr(field+".kind", "must be newsletter or contact")
	}
	var err error
	texts := []struct {
		v     *string
		max   int
		multi bool
		name  string
	}{
		{&d.Title, 80, false, "title"}, {&d.Description, 300, true, "description"}, {&d.ButtonText, 30, false, "button_text"},
		{&d.SuccessText, 200, false, "success_text"}, {&d.ConsentText, 300, false, "consent_text"},
	}
	for _, t := range texts {
		if *t.v, err = cleanText(*t.v, t.max, t.multi, field+"."+t.name); err != nil {
			return d, err
		}
	}
	if d.ButtonText == "" {
		if d.Kind == formKindContact {
			d.ButtonText = "Send message"
		} else {
			d.ButtonText = "Subscribe"
		}
	}
	if d.SuccessText == "" {
		d.SuccessText = "Thanks! We got it."
	}
	return d, nil
}

// contentHasBlockType reports whether a document contains a block type.
func contentHasBlockType(c models.MicrositeContent, blockType string) bool {
	for _, b := range c.Blocks {
		if b.Type == blockType {
			return true
		}
	}
	return false
}

// formBlockIDs returns the IDs of form blocks in a document.
func formBlockIDs(c models.MicrositeContent) map[string]bool {
	ids := map[string]bool{}
	for _, b := range c.Blocks {
		if b.Type == models.BlockTypeForm {
			ids[b.ID] = true
		}
	}
	return ids
}
