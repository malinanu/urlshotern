package services

import (
	"embed"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

var (
	//go:embed microsite_web/styles/*.css
	micrositeStyleFS embed.FS
	//go:embed microsite_web/blocks.css
	micrositeBlocksCSS string
	//go:embed microsite_web/embed.js
	micrositeEmbedJS []byte
	//go:embed microsite_web/countdown.js
	micrositeCountdownJS []byte
	//go:embed microsite_web/fonts/*.woff2
	micrositeFontFS embed.FS
)

// Static paths for v2 page assets.
const (
	MicrositeEmbedJSPath     = "/m-static/embed.js"
	MicrositeCountdownJSPath = "/m-static/countdown.js"
	MicrositeFontPathPrefix  = "/m-static/fonts/"
)

// MicrositeStaticScript returns an embedded page script by path.
func MicrositeStaticScript(path string) ([]byte, bool) {
	switch path {
	case MicrositeEmbedJSPath:
		return micrositeEmbedJS, true
	case MicrositeCountdownJSPath:
		return micrositeCountdownJS, true
	case MicrositeBeaconPath:
		return MicrositeBeaconJS, true
	}
	return nil, false
}

// MicrositeFontFile returns an embedded woff2 by file name.
func MicrositeFontFile(name string) ([]byte, bool) {
	if strings.ContainsAny(name, "/\\") || !strings.HasSuffix(name, ".woff2") {
		return nil, false
	}
	b, err := micrositeFontFS.ReadFile("microsite_web/fonts/" + name)
	return b, err == nil
}

type blockGroup struct {
	Class  string
	Blocks []micrositeBlockView
}

type productView struct {
	BlockID      string
	ItemID       string
	DomID        string
	Image        string
	Name         string
	Description  string
	Badge        string
	PriceText    string // whole price when it doesn't split into parts
	PriceLead    string
	PriceMain    string
	PriceSep     string
	PriceCents   string
	PriceTail    string
	ComparePrice string
	Href         template.URL
	External     bool
	ButtonText   string
}

type carouselView struct {
	Title string
	Items []productView
}

type eventView struct {
	BlockID        string
	Title          string
	StartISO       string
	DateText       string
	TimeText       string
	Zone           string
	Venue          string
	Address        string
	TicketHref     template.URL
	TicketExternal bool
	TicketText     string
	CalendarHref   string
}

type countdownView struct {
	TargetISO string
	Label     string
	DoneText  string
	Done      bool
	Days      int
	Hours     string
	Minutes   string
}

type agendaView struct {
	Title string
	Items []agendaItemView
}

type agendaItemView struct {
	models.AgendaItem
	TrackIndex int
}

type mapView struct {
	BlockID        string
	Label          string
	Address        string
	EmbedURL       string
	OpenURL        string
	DirectionsHref string
}

type embedView struct {
	BlockID      string
	Provider     string
	ProviderName string
	EmbedURL     string
	WatchURL     string
	Height       int
	Label        string
	Poster       string
}

type formView struct {
	BlockID     string
	Title       string
	Description string
	Contact     bool
	ButtonText  string
	SuccessText string
	ConsentText string
	Action      string
	Token       string
	Sent        bool
	Error       string
	Disabled    bool
}

// FormErrorMessages maps submit error codes to visitor-facing text.
var FormErrorMessages = map[string]string{
	"invalid_email":   "Please enter a valid email address.",
	"missing_message": "Please write a message.",
	"too_fast":        "That was quick! Please wait a moment and send again.",
	"expired":         "This form expired. Please send it again.",
	"rate_limited":    "Too many messages right now. Please try again later.",
	"closed":          "This form isn't accepting messages right now.",
	"invalid":         "Something was wrong with that submission. Please try again.",
}

// buildV2BlockView fills bv for a v2 block. It returns false when the block
// has nothing to show (incomplete, or forms disabled for the owner's plan).
func buildV2BlockView(bv *micrositeBlockView, data interface{}, base string, opts MicrositeRenderOptions) bool {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	switch d := data.(type) {
	case models.ProductItem:
		pv, ok := buildProductView(bv.ID, "", d, base)
		if !ok {
			return false
		}
		bv.Product = &pv

	case models.ProductCarouselData:
		cv := &carouselView{Title: d.Title}
		for _, it := range d.Items {
			if pv, ok := buildProductView(bv.ID, it.ID, it, base); ok {
				cv.Items = append(cv.Items, pv)
			}
		}
		if len(cv.Items) == 0 {
			return false
		}
		bv.Carousel = cv

	case models.EventDetailsData:
		if d.Title == "" || d.Start == "" || d.Timezone == "" {
			return false
		}
		start, err := parseEventTime(d.Start, d.Timezone)
		if err != nil {
			return false
		}
		ev := &eventView{
			BlockID: bv.ID, Title: d.Title, StartISO: start.Format(time.RFC3339),
			DateText: start.Format("Mon 2 Jan 2006"), TimeText: start.Format("15:04"),
			Zone: gmtOffset(start), Venue: d.Venue, Address: d.Address,
			TicketText: firstNonEmpty(d.TicketText, "Get tickets"),
		}
		if d.End != "" {
			if end, err := parseEventTime(d.End, d.Timezone); err == nil {
				if end.YearDay() == start.YearDay() && end.Year() == start.Year() {
					ev.TimeText += " – " + end.Format("15:04")
				} else {
					ev.DateText += " – " + end.Format("Mon 2 Jan 2006")
				}
			}
		}
		if d.TicketURL != "" {
			if href, err := SanitizeMicrositeLinkURL(d.TicketURL, "ticket_url"); err == nil {
				ev.TicketHref = template.URL(href) // #nosec G203 -- scheme allowlisted
				ev.TicketExternal = strings.HasPrefix(href, "http")
			}
		}
		if d.ShowAddToCalendar && opts.Slug != "" {
			ev.CalendarHref = "/m/" + url.PathEscape(opts.Slug) + "/ics?b=" + url.QueryEscape(bv.ID)
		}
		bv.Event = ev

	case models.CountdownData:
		if d.Target == "" || d.Timezone == "" {
			return false
		}
		target, err := parseEventTime(d.Target, d.Timezone)
		if err != nil {
			return false
		}
		cv := &countdownView{TargetISO: target.Format(time.RFC3339), Label: d.Label, DoneText: firstNonEmpty(d.DoneText, "It's happening now")}
		left := target.Sub(now)
		if left <= 0 {
			cv.Done = true
		} else {
			mins := int(left.Minutes())
			cv.Days, cv.Hours, cv.Minutes = mins/1440, fmt.Sprintf("%02d", (mins/60)%24), fmt.Sprintf("%02d", mins%60)
		}
		bv.Countdown = cv

	case models.AgendaData:
		av := &agendaView{Title: d.Title}
		tracks := map[string]int{}
		for _, it := range d.Items {
			if it.Title == "" {
				continue
			}
			idx := 0
			if it.Track != "" {
				key := strings.ToLower(it.Track)
				if _, seen := tracks[key]; !seen {
					tracks[key] = len(tracks) % 4
				}
				idx = tracks[key]
			}
			av.Items = append(av.Items, agendaItemView{AgendaItem: it, TrackIndex: idx})
		}
		if len(av.Items) == 0 {
			return false
		}
		bv.Agenda = av

	case models.MapData:
		if d.Address == "" && d.Lat == nil {
			return false
		}
		mv := &mapView{BlockID: bv.ID, Label: d.Label, Address: d.Address}
		dest := d.Address
		if d.Lat != nil && d.Lng != nil {
			lat, lng := *d.Lat, *d.Lng
			dest = fmt.Sprintf("%.6f,%.6f", lat, lng)
			dLng := 0.012 * math.Pow(2, float64(defaultMapZoom-d.Zoom))
			dLat := dLng * 0.6
			mv.EmbedURL = fmt.Sprintf("https://www.openstreetmap.org/export/embed.html?bbox=%.6f%%2C%.6f%%2C%.6f%%2C%.6f&layer=mapnik&marker=%.6f%%2C%.6f",
				lng-dLng, lat-dLat, lng+dLng, lat+dLat, lat, lng)
			mv.OpenURL = fmt.Sprintf("https://www.openstreetmap.org/?mlat=%.6f&mlon=%.6f#map=%d/%.6f/%.6f", lat, lng, d.Zoom, lat, lng)
		}
		mv.DirectionsHref = "https://www.google.com/maps/dir/?api=1&destination=" + url.QueryEscape(dest)
		bv.Map = mv

	case models.VideoData, models.AudioData:
		var (
			ref    *EmbedRef
			err    error
			title  string
			poster string
		)
		if v, ok := d.(models.VideoData); ok {
			ref, err = ParseVideoURL(v.URL)
			title, poster = v.Title, absoluteAssetURL(v.PosterURL, base)
		} else {
			a := d.(models.AudioData)
			ref, err = ParseAudioURL(a.URL)
			title = a.Title
		}
		if err != nil {
			return false
		}
		bv.Embed = &embedView{
			BlockID: bv.ID, Provider: ref.Provider, ProviderName: ref.ProviderName(),
			EmbedURL: ref.EmbedURL(), WatchURL: ref.WatchURL(), Height: ref.EmbedHeight(),
			Label: firstNonEmpty(title, "Play on "+ref.ProviderName()), Poster: poster,
		}

	case models.FormData:
		if !opts.FormsEnabled {
			return false
		}
		fv := &formView{
			BlockID: bv.ID, Title: d.Title, Description: d.Description, Contact: d.Kind == formKindContact,
			ButtonText: d.ButtonText, SuccessText: d.SuccessText, ConsentText: d.ConsentText,
			Sent: opts.SentBlock == bv.ID,
		}
		if opts.Slug == "" || opts.Preview {
			fv.Action, fv.Disabled = "#", true
		} else {
			fv.Action = "/m/" + url.PathEscape(opts.Slug) + "/submit"
		}
		if opts.FormToken != nil {
			fv.Token = opts.FormToken(bv.ID)
		}
		if opts.ErrorBlock == bv.ID {
			fv.Error = firstNonEmpty(FormErrorMessages[opts.ErrorCode], FormErrorMessages["invalid"])
		}
		bv.Form = fv

	default:
		return false
	}
	return true
}

func buildProductView(blockID, itemID string, d models.ProductItem, base string) (productView, bool) {
	if d.Name == "" {
		return productView{}, false
	}
	pv := productView{
		BlockID: blockID, ItemID: itemID, DomID: blockID, Image: absoluteAssetURL(d.ImageURL, base), Name: d.Name,
		Description: d.Description, Badge: d.Badge, ComparePrice: d.ComparePrice,
		ButtonText: firstNonEmpty(d.ButtonText, "Buy now"),
	}
	if itemID != "" {
		pv.DomID = blockID + "-" + itemID
	}
	pv.PriceLead, pv.PriceMain, pv.PriceSep, pv.PriceCents, pv.PriceTail = splitPrice(d.Price)
	if pv.PriceMain == "" && d.Price != "" {
		pv.PriceText = d.Price
	}
	if d.BuyURL != "" {
		if href, err := SanitizeMicrositeLinkURL(d.BuyURL, "buy_url"); err == nil {
			pv.Href = template.URL(href) // #nosec G203 -- scheme allowlisted
			pv.External = strings.HasPrefix(href, "http")
		}
	}
	return pv, true
}

var pricePattern = regexp.MustCompile(`^(\D*?)\s?(\d[\d,.' ]*?)(?:([.,])(\d{2}))?(\D*)$`)

// splitPrice splits "$24.99" into "$", "24", ".", "99", "" so prices can be
// typeset with small cents. Anything that doesn't look like a price returns
// empty parts and is shown as written.
func splitPrice(s string) (lead, main, sep, cents, tail string) {
	m := pricePattern.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return "", "", "", "", ""
	}
	return strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), m[3], m[4], strings.TrimSpace(m[5])
}

// linkVia names where a link goes, for the board's "via" column.
func linkVia(href string) string {
	switch {
	case strings.HasPrefix(href, "mailto:"):
		return "email"
	case strings.HasPrefix(href, "tel:"):
		return "phone"
	case strings.HasPrefix(href, "sms:"):
		return "text"
	}
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
}

// gmtOffset renders a zone offset as "GMT+5:30".
func gmtOffset(t time.Time) string {
	_, off := t.Zone()
	sign := "+"
	if off < 0 {
		sign, off = "-", -off
	}
	h, m := off/3600, (off%3600)/60
	if m == 0 {
		return fmt.Sprintf("GMT%s%d", sign, h)
	}
	return fmt.Sprintf("GMT%s%d:%02d", sign, h, m)
}

// applyV2Layout adds style, fonts, block CSS and scripts. For a classic page
// with only v1 blocks it changes nothing, so the output stays byte-identical.
func applyV2Layout(view *micrositePageView, theme models.MicrositeTheme, base string) {
	var css strings.Builder
	hasV2, hasEmbed, hasCountdown, hasAgenda := false, false, false, false
	for _, b := range view.Blocks {
		switch b.Type {
		case models.BlockTypeProductCard, models.BlockTypeProductCarousel, models.BlockTypeEventDetails,
			models.BlockTypeMap, models.BlockTypeForm:
			hasV2 = true
		case models.BlockTypeCountdown:
			hasV2, hasCountdown = true, true
		case models.BlockTypeAgenda:
			hasV2, hasAgenda = true, true
		case models.BlockTypeVideo, models.BlockTypeAudio:
			hasV2, hasEmbed = true, true
		}
		if b.Map != nil && b.Map.EmbedURL != "" {
			hasEmbed = true
		}
	}

	styled := theme.Style != "" && theme.Style != MicrositeStyleClassic
	if faces, ok := MicrositeFontFaces[theme.Font]; ok {
		for _, f := range faces {
			src := base + MicrositeFontPathPrefix + f.File
			fmt.Fprintf(&css, `@font-face{font-family:"%s";src:url("%s") format("woff2");font-weight:%s;font-style:normal;font-display:swap}`, f.Family, src, f.Weight)
			view.FontPreloads = append(view.FontPreloads, src)
		}
	}
	if styled || hasAgenda {
		css.WriteString(styleTokenCSS(theme))
	}
	if hasV2 {
		css.WriteString(micrositeBlocksCSS)
	}
	if styled {
		if b, err := micrositeStyleFS.ReadFile("microsite_web/styles/" + theme.Style + ".css"); err == nil {
			css.Write(b)
		}
		view.BodyClass = "s-" + theme.Style
		view.Groups = groupBlocks(theme.Style, view.Blocks)
	}
	view.ExtraCSS = template.CSS(css.String()) // #nosec G203 -- embedded files plus tokens from validated hex/enum values
	if !view.Preview && view.BeaconSrc != "" {
		if hasEmbed {
			view.Scripts = append(view.Scripts, MicrositeEmbedJSPath)
		}
		if hasCountdown {
			view.Scripts = append(view.Scripts, MicrositeCountdownJSPath)
		}
	}
}

// styleTokenCSS derives extra color tokens styles use (all from validated hex).
func styleTokenCSS(t models.MicrositeTheme) string {
	return fmt.Sprintf(":root{--accent-12:%s;--accent-30:%s;--accent-deep:%s;--trk-0:%s;--trk-1:%s;--trk-2:%s;--trk-3:%s}",
		blendHex(t.Accent, "#ffffff", 0.12), blendHex(t.Accent, "#ffffff", 0.30), blendHex(t.Accent, "#000000", 0.70),
		t.Accent, blendHex(t.Accent, t.Text, 0.45), blendHex(t.Accent, "#ffffff", 0.45), t.Text)
}

// groupBlocks splits a styled page into its lead and body regions.
func groupBlocks(style string, blocks []micrositeBlockView) []blockGroup {
	if len(blocks) == 0 {
		return nil
	}
	lead := 0
	switch style {
	case MicrositeStyleLetterpress:
		lead = leadingRun(blocks, models.BlockTypeProfileHeader)
		for i, b := range blocks {
			if b.Type == models.BlockTypeVCardAction {
				lead = i + 1
				break
			}
		}
	case MicrositeStyleProgramme:
		lead = leadingRun(blocks, models.BlockTypeProfileHeader, models.BlockTypeEventDetails, models.BlockTypeCountdown)
	default:
		lead = leadingRun(blocks, models.BlockTypeProfileHeader)
	}
	names := map[string][2]string{
		MicrositeStyleLetterpress: {"lp-front", "lp-back"},
		MicrositeStyleBoard:       {"bd-plate", "bd-board"},
		MicrositeStyleFlyer:       {"fl-mast", "fl-sheet"},
		MicrositeStyleProgramme:   {"pg-cover", "pg-body"},
	}[style]
	var groups []blockGroup
	if lead > 0 {
		groups = append(groups, blockGroup{Class: "grp " + names[0], Blocks: blocks[:lead]})
	}
	if lead < len(blocks) {
		groups = append(groups, blockGroup{Class: "grp " + names[1], Blocks: blocks[lead:]})
	}
	return groups
}

func leadingRun(blocks []micrositeBlockView, types ...string) int {
	allowed := map[string]bool{}
	for _, t := range types {
		allowed[t] = true
	}
	n := 0
	for n < len(blocks) && allowed[blocks[n].Type] {
		n++
	}
	return n
}
