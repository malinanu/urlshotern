package services

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/URLshorter/url-shortener/internal/models"
)

// Digital visiting cards for Microsites.
//
// Twelve card templates share one stylesheet set, one catalog and one set of fonts with the
// React renderer used by MicroPages. They are copied from
// frontend/src/components/micropages/cards by scripts/sync-card-assets.sh, and a test fails if
// the copies drift. This file turns a page's blocks into the same card markup the React
// renderer produces, so a card looks identical in both products.

var (
	//go:embed microsite_web/cards/catalog.json
	cardCatalogJSON []byte
	//go:embed microsite_web/cards/*.css
	cardCSSFS embed.FS
	//go:embed microsite_web/card-export.js
	micrositeCardExportJS []byte
	//go:embed microsite_web/html-to-image.js
	micrositeHTMLToImageJS []byte
)

// Static paths for the card page assets.
const (
	MicrositeCardExportJSPath  = "/m-static/card-export.js"
	MicrositeHTMLToImageJSPath = "/m-static/html-to-image.js"
)

// CardPalette is one curated colour set for a card template.
type CardPalette struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Vars map[string]string `json:"vars"`
}

// CardTemplate describes one of the twelve digital visiting card templates.
type CardTemplate struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Collection  string        `json:"collection"`
	Category    string        `json:"category"`
	Description string        `json:"description"`
	Audience    string        `json:"audience"`
	CTA         string        `json:"cta"`
	Logo        string        `json:"logo"` // default logo background: light | dark | none
	Fonts       []string      `json:"fonts"`
	Palettes    []CardPalette `json:"palettes"`
}

type cardCatalogFile struct {
	Collections []models.MicrositeCollection `json:"collections"`
	Templates   []CardTemplate               `json:"templates"`
}

var (
	cardCatalog = mustLoadCardCatalog()
	cardByID    = func() map[string]*CardTemplate {
		m := make(map[string]*CardTemplate, len(cardCatalog.Templates))
		for i := range cardCatalog.Templates {
			m[cardCatalog.Templates[i].ID] = &cardCatalog.Templates[i]
		}
		return m
	}()
)

// Palette values come from the embedded catalog, but they end up in a style attribute, so they
// are checked at start-up: only hex colours and rgba() are accepted.
var cardVarValue = regexp.MustCompile(`^(#[0-9a-fA-F]{3,8}|rgba?\([0-9., %]+\))$`)

// cardVarNames maps catalog palette keys to CSS custom properties, in a fixed output order.
var cardVarNames = []struct{ key, css string }{
	{"bg", "--vc-bg"}, {"ink", "--vc-ink"}, {"mute", "--vc-mute"}, {"accent", "--vc-accent"},
	{"accent2", "--vc-accent2"}, {"surface", "--vc-surface"}, {"line", "--vc-line"},
	{"onAccent", "--vc-on-accent"}, {"plate", "--vc-plate"}, {"qr", "--vc-qr"},
	{"g2", "--vc-g2"}, {"g3", "--vc-g3"}, {"glow", "--vc-glow"}, {"sky1", "--vc-sky1"}, {"sky2", "--vc-sky2"},
}

// cardDefaults are each template's stylesheet defaults: page colour, text, accent, text on accent.
var cardDefaults = map[string]struct{ BG, Ink, Accent, OnAccent string }{
	"zenith":  {"#0b0b0d", "#f4eee2", "#c9aa6a", "#17130a"},
	"nova":    {"#f7f2e9", "#2a2723", "#a98b5b", "#fbf9f4"},
	"atlas":   {"#f5ecdc", "#3a2e22", "#b48a3c", "#fffaf0"},
	"axiom":   {"#090c13", "#e8f1f8", "#22d3ee", "#031318"},
	"vectra":  {"#121418", "#eceff3", "#c9ced6", "#101216"},
	"pulse":   {"#0a0f2c", "#f1f3ff", "#3d6bff", "#ffffff"},
	"horizon": {"#ffffff", "#0f2748", "#1f5fd6", "#ffffff"},
	"verdant": {"#0e2a1e", "#f2eee1", "#c8a45e", "#0e2a1e"},
	"skyway":  {"#f3f9ff", "#0d2c4f", "#1672c4", "#ffffff"},
	"creo":    {"#1b0836", "#fdf7ff", "#ff7a59", "#2a0a1a"},
	"bravo":   {"#f6eedb", "#111111", "#ff6a13", "#111111"},
	"imagine": {"#17181c", "#f7f4ea", "#ffd60a", "#17181c"},
}

// cardSamples is the starter content of each template's sample page.
var cardSamples = map[string]struct {
	Role string
	Tags []string
}{
	"zenith": {"Chief Executive Officer", nil}, "nova": {"Marketing Director", nil},
	"atlas": {"Financial Adviser", nil}, "axiom": {"Founder and CTO", nil},
	"vectra": {"Head of Product", nil}, "pulse": {"Growth Strategist", nil},
	"horizon": {"Senior Account Executive", nil}, "verdant": {"Sustainability Lead", nil},
	"skyway": {"Travel Consultant", nil}, "creo": {"Creative Director", []string{"Branding", "Strategy", "Identity"}},
	"bravo": {"Business Development Manager", nil}, "imagine": {"Brand Strategist", []string{"Brand", "Content", "Campaigns"}},
}

func mustLoadCardCatalog() cardCatalogFile {
	var c cardCatalogFile
	if err := json.Unmarshal(cardCatalogJSON, &c); err != nil {
		panic(fmt.Sprintf("card catalog: %v", err))
	}
	for _, t := range c.Templates {
		if len(t.Palettes) == 0 {
			panic("card catalog: " + t.ID + " has no palettes")
		}
		if _, ok := cardDefaults[t.ID]; !ok {
			panic("card catalog: no defaults for " + t.ID)
		}
		for _, p := range t.Palettes {
			for k, v := range p.Vars {
				if !cardVarValue.MatchString(v) {
					panic(fmt.Sprintf("card catalog: %s/%s %s has an unsafe value %q", t.ID, p.ID, k, v))
				}
			}
		}
	}
	return c
}

func init() {
	// Card styles are valid theme styles.
	for _, t := range cardCatalog.Templates {
		MicrositeStyles[t.ID] = true
	}
}

// CardTemplateByID returns a card template.
func CardTemplateByID(id string) (*CardTemplate, bool) {
	t, ok := cardByID[id]
	return t, ok
}

// IsCardStyle reports whether a theme style is one of the twelve card templates.
func IsCardStyle(style string) bool {
	_, ok := cardByID[style]
	return ok
}

// CardCollections lists the four template collections.
func CardCollections() []models.MicrositeCollection { return cardCatalog.Collections }

func (t *CardTemplate) palette(id string) (CardPalette, bool) {
	for _, p := range t.Palettes {
		if p.ID == id {
			return p, true
		}
	}
	return CardPalette{}, false
}

func (t *CardTemplate) paletteAccent(id string) string {
	if p, ok := t.palette(id); ok && p.Vars["accent"] != "" {
		return p.Vars["accent"]
	}
	return cardDefaults[t.ID].Accent
}

var hexColor6 = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// readableOn returns near-black for light colours and white for dark ones (WCAG luminance).
func readableOn(hex string) string {
	if !hexColor6.MatchString(hex) {
		return "#ffffff"
	}
	n, _ := strconv.ParseUint(hex[1:], 16, 32)
	lin := func(c uint64) float64 {
		s := float64(c) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	l := 0.2126*lin((n>>16)&255) + 0.7152*lin((n>>8)&255) + 0.0722*lin(n&255)
	if l > 0.4 {
		return "#14110c"
	}
	return "#ffffff"
}

// cardStyleAttr builds the inline custom properties for a card: the palette, then a custom
// accent when the page's accent differs from the palette's.
func cardStyleAttr(t *CardTemplate, theme models.MicrositeTheme) template.CSS {
	var b strings.Builder
	if p, ok := t.palette(theme.Palette); ok {
		for _, n := range cardVarNames {
			if v := p.Vars[n.key]; v != "" {
				fmt.Fprintf(&b, "%s:%s;", n.css, v)
			}
		}
	}
	if hexColor6.MatchString(theme.Accent) && !strings.EqualFold(theme.Accent, t.paletteAccent(theme.Palette)) {
		fmt.Fprintf(&b, "--vc-accent:%s;--vc-on-accent:%s;", strings.ToLower(theme.Accent), readableOn(theme.Accent))
	}
	return template.CSS(b.String()) // #nosec G203 -- values are embedded catalog entries and a validated #rrggbb colour
}

// cardFontFaces maps catalog font file keys to their @font-face data.
var cardFontFaces = map[string]MicrositeFontFace{
	"cormorant-garamond":  {Family: "Cormorant Garamond", File: "cormorant-garamond.woff2", Weight: "300 700"},
	"jost":                {Family: "Jost", File: "jost.woff2", Weight: "100 900"},
	"playfair-display":    {Family: "Playfair Display", File: "playfair-display.woff2", Weight: "400 900"},
	"source-sans-3":       {Family: "Source Sans 3", File: "source-sans-3.woff2", Weight: "200 900"},
	"chakra-petch-500":    {Family: "Chakra Petch", File: "chakra-petch-500.woff2", Weight: "500"},
	"chakra-petch-700":    {Family: "Chakra Petch", File: "chakra-petch-700.woff2", Weight: "700"},
	"sora":                {Family: "Sora", File: "sora.woff2", Weight: "100 800"},
	"bricolage-grotesque": {Family: "Bricolage Grotesque", File: "bricolage-grotesque.woff2", Weight: "300 800"},
	"lora":                {Family: "Lora", File: "lora.woff2", Weight: "400 700"},
	"red-hat-display":     {Family: "Red Hat Display", File: "red-hat-display.woff2", Weight: "300 900"},
	"syne":                {Family: "Syne", File: "syne.woff2", Weight: "400 800"},
	"archivo-black":       {Family: "Archivo Black", File: "archivo-black.woff2", Weight: "400"},
	"familjen-grotesk":    {Family: "Familjen Grotesk", File: "familjen-grotesk.woff2", Weight: "400 700"},
	"work-sans":           {Family: "Work Sans", File: "work-sans.woff2", Weight: "100 900"},
	"hanken-grotesk":      {Family: "Hanken Grotesk", File: "hanken-grotesk.woff2", Weight: "100 900"},
}

// cardCSS returns the stylesheet text for a card: shared base, the template, and the page chrome.
func cardCSS(t *CardTemplate) string {
	var b strings.Builder
	for _, name := range []string{"base", t.ID, "page"} {
		if data, err := cardCSSFS.ReadFile("microsite_web/cards/" + name + ".css"); err == nil {
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// ---- QR code ----------------------------------------------------------------

// cardQRShape builds a QR code as an SVG path (horizontal runs of dark modules) with a
// 4-module quiet zone, matching the React renderer. Error correction level M.
func cardQRShape(text string) (viewBox, path string) {
	if text == "" {
		return "", ""
	}
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return "", ""
	}
	q.DisableBorder = true
	bm := q.Bitmap()
	n := len(bm)
	var b strings.Builder
	for y := 0; y < n; y++ {
		x := 0
		for x < n {
			if !bm[y][x] {
				x++
				continue
			}
			w := 1
			for x+w < n && bm[y][x+w] {
				w++
			}
			fmt.Fprintf(&b, "M%d %dh%dv1h-%dz", x, y, w, w)
			x += w
		}
	}
	return fmt.Sprintf("-4 -4 %d %d", n+8, n+8), b.String()
}

// ---- card view --------------------------------------------------------------

type cardRow struct {
	Kind     string
	Label    string
	Value    string
	Href     template.URL
	Icon     template.HTML
	External bool
}

type cardSocial struct {
	Name    string
	Href    template.URL
	Glyph   template.HTML
	BlockID string
	ID      string
}

// cardView is everything the card markup needs.
type cardView struct {
	Template  string
	NameClass string
	LogoPlate string
	StyleAttr template.CSS
	IDNumber  string
	Name      string
	Role      string
	Company   string
	About     string
	Photo     string
	Logo      string
	Initials  string
	Tags      []string
	Rows      []cardRow
	Socials   []cardSocial
	QRViewBox string
	QRPath    string
	CTALabel  string
	CTAHref   string
	HasCTA    bool
	Slug      string
	Actions   bool
}

// nameLengthClass picks the name-size step used by the stylesheets.
func nameLengthClass(name string) string {
	switch n := utf8.RuneCountInString(name); {
	case n <= 14:
		return "s"
	case n <= 22:
		return "m"
	case n <= 32:
		return "l"
	default:
		return "xl"
	}
}

const (
	cardIconPhone    = `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2"/></svg>`
	cardIconWhatsApp = `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 21l1.65-3.8a9 9 0 1 1 3.4 2.9L3 21"/><path d="M9 10a.5.5 0 0 0 1 0V9a.5.5 0 0 0-1 0v1a5 5 0 0 0 5 5h1a.5.5 0 0 0 0-1h-1a.5.5 0 0 0 0 1"/></svg>`
	cardIconEmail    = `<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="5" width="18" height="14" rx="2"/><path d="M3 7l9 6 9-6"/></svg>`
	cardIconWebsite  = `<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M3.6 9h16.8M3.6 15h16.8M11.5 3a17 17 0 0 0 0 18M12.5 3a17 17 0 0 1 0 18"/></svg>`
	cardIconAddress  = `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 21s7-6.2 7-11a7 7 0 1 0-14 0c0 4.8 7 11 7 11z"/><circle cx="12" cy="10" r="2.5"/></svg>`
	cardIconArrow    = `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h14M13 6l6 6-6 6"/></svg>`
)

// cardSocialPaths are filled brand glyphs (24px grid). Unknown platforms use the globe.
var cardSocialPaths = map[string]string{
	"linkedin":  "M19 3a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14m-.5 15.5v-5.3a3.26 3.26 0 0 0-3.26-3.26c-.85 0-1.84.52-2.28 1.3v-1.11h-2.79v8.37h2.79v-4.93c0-.77.62-1.4 1.39-1.4a1.4 1.4 0 0 1 1.4 1.4v4.93h2.75M6.46 10.9h2.79v8.37H6.46v-8.37M7.86 6.88a1.45 1.45 0 0 0-1.46 1.46 1.46 1.46 0 0 0 1.46 1.46 1.46 1.46 0 0 0 1.46-1.46c0-.81-.66-1.46-1.46-1.46z",
	"x":         "M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z",
	"youtube":   "M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z",
	"instagram": "M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838a6.162 6.162 0 1 0 0 12.324 6.162 6.162 0 0 0 0-12.324zm0 10.162a3.999 3.999 0 1 1 0-7.998 3.999 3.999 0 0 1 0 7.998zm6.406-11.845a1.44 1.44 0 1 0 0 2.881 1.44 1.44 0 0 0 0-2.881z",
	"github":    "M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.87 1.52 2.34 1.07 2.91.83.1-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.92 0-1.11.38-2 1.03-2.71-.1-.25-.45-1.29.1-2.64 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.35.2 2.39.1 2.64.65.71 1.03 1.6 1.03 2.71 0 3.82-2.34 4.66-4.57 4.91.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z",
	"telegram":  "M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69.01-.03.01-.14-.06-.19-.06-.06-.17-.04-.24-.02-.11.02-1.78 1.13-5.02 3.32-.47.33-.9.49-1.29.48-.42-.01-1.24-.24-1.84-.44-.75-.24-1.34-.37-1.29-.79.03-.22.33-.44.9-.68 3.52-1.53 5.87-2.54 7.05-3.03 3.36-1.4 4.06-1.64 4.51-1.65.1 0 .32.02.46.14.12.1.15.24.17.34-.01.07.01.21 0 .28z",
}

const cardGlobePath = "M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm6.93 9h-3.2a15.7 15.7 0 0 0-1.2-5.2A8 8 0 0 1 18.93 11zM12 4c.83 1.2 1.5 3.1 1.73 5h-3.46C10.5 7.1 11.17 5.2 12 4zM4.26 13h3.2c.1 1.9.5 3.7 1.2 5.2A8 8 0 0 1 4.26 13zm3.2-2h-3.2a8 8 0 0 1 4.4-5.2c-.7 1.5-1.1 3.3-1.2 5.2zM12 20c-.83-1.2-1.5-3.1-1.73-5h3.46c-.23 1.9-.9 3.8-1.73 5zm2.27-7H9.73a14 14 0 0 1 0-2h4.54a14 14 0 0 1 0 2zm.27 5.2c.7-1.5 1.1-3.3 1.2-5.2h3.2a8 8 0 0 1-4.4 5.2z"

func cardSocialGlyph(platform string) template.HTML {
	p := strings.ToLower(platform)
	if p == "twitter" {
		p = "x"
	}
	d, ok := cardSocialPaths[p]
	if !ok {
		d = cardGlobePath
	}
	return template.HTML(`<svg viewBox="0 0 24 24" aria-hidden="true"><path d="` + d + `"/></svg>`) // #nosec G203 -- constant SVG path data
}

func onlyDigitsPlus(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '+' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// buildCardView consumes the blocks a card shows (profile, contact card, contact buttons,
// social icons, QR code) and returns the remaining blocks, which render under the card.
func buildCardView(t *CardTemplate, theme models.MicrositeTheme, blocks []micrositeBlockView, base string, opts MicrositeRenderOptions) (*cardView, []micrositeBlockView) {
	cv := &cardView{
		Template:  t.ID,
		StyleAttr: cardStyleAttr(t, theme),
		IDNumber:  fmt.Sprintf("TC-%06d", opts.PageID),
		Slug:      opts.Slug,
		Actions:   !opts.Preview && opts.Slug != "",
	}
	var (
		profile  *profileView
		vcard    *vcardButtonView
		contacts []contactView
		socials  []socialView
		qrURL    string
		haveQR   bool
		rest     = make([]micrositeBlockView, 0, len(blocks))
	)
	for _, b := range blocks {
		switch {
		case b.Type == models.BlockTypeProfileHeader && profile == nil && b.Profile != nil:
			profile = b.Profile
		case b.Type == models.BlockTypeVCardAction && vcard == nil && b.VCard != nil:
			vcard = b.VCard
		case b.Type == models.BlockTypeContactButtons && contacts == nil:
			contacts = b.Contacts
		case b.Type == models.BlockTypeSocialIcons && socials == nil:
			socials = b.Socials
		case b.Type == models.BlockTypeQRCode && !haveQR:
			qrURL, haveQR = b.QRURL, true
		default:
			rest = append(rest, b)
		}
	}

	cv.Name = "Your Name"
	plate := ""
	if profile != nil {
		if profile.DisplayName != "" {
			cv.Name = profile.DisplayName
		}
		cv.Role, cv.Company, cv.About = profile.Title, profile.Company, profile.Bio
		cv.Photo, cv.Logo = profile.AvatarURL, profile.CompanyLogoURL
		cv.Tags = profile.Tags
		plate = profile.LogoPlate
	}
	if plate == "" || plate == "auto" {
		plate = t.Logo
	}
	cv.LogoPlate = plate
	cv.NameClass = nameLengthClass(cv.Name)
	cv.Initials = initialsOf(cv.Name)

	// Contact rows: phone, WhatsApp, email, website, address.
	var phone, email, website, address, whatsapp, whatsappHref string
	if vcard != nil {
		phone, email, website, address = vcard.Phone, vcard.Email, vcard.Website, vcard.Address
	}
	for _, c := range contacts {
		switch c.Kind {
		case "call":
			if phone == "" {
				phone = c.Value
			}
		case "email":
			if email == "" {
				email = c.Value
			}
		case "whatsapp":
			if whatsapp == "" {
				whatsapp, whatsappHref = c.Value, string(c.Href)
			}
		}
	}
	if phone != "" {
		cv.Rows = append(cv.Rows, cardRow{Kind: "phone", Label: "Phone", Value: phone, Icon: cardIconPhone, Href: template.URL("tel:" + onlyDigitsPlus(phone))}) // #nosec G203 -- digits and + only
	}
	if whatsapp != "" && whatsappHref != "" {
		cv.Rows = append(cv.Rows, cardRow{Kind: "whatsapp", Label: "WhatsApp", Value: whatsapp, Icon: cardIconWhatsApp, Href: template.URL(whatsappHref), External: true}) // #nosec G203 -- built by contactHref from a validated number
	}
	if email != "" {
		cv.Rows = append(cv.Rows, cardRow{Kind: "email", Label: "Email", Value: email, Icon: cardIconEmail, Href: template.URL("mailto:" + email)}) // #nosec G203 -- validated email address
	}
	if website != "" {
		if href, err := SanitizeMicrositeLinkURL(website, "website"); err == nil && strings.HasPrefix(href, "http") {
			cv.Rows = append(cv.Rows, cardRow{Kind: "website", Label: "Web", Value: websiteLabel(website), Icon: cardIconWebsite, Href: template.URL(href), External: true}) // #nosec G203 -- scheme allowlisted by SanitizeMicrositeLinkURL
		}
	}
	if address != "" {
		cv.Rows = append(cv.Rows, cardRow{Kind: "address", Label: "Address", Value: address, Icon: cardIconAddress, Href: template.URL("https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(address)), External: true}) // #nosec G203 -- fixed https URL with an escaped query
	}

	for _, s := range socials {
		cv.Socials = append(cv.Socials, cardSocial{Name: s.Name, Href: template.URL(s.Href), Glyph: cardSocialGlyph(s.Platform), BlockID: s.BlockID, ID: s.ID}) // #nosec G203 -- href already sanitized in buildBlockView
	}

	if vcard != nil {
		cv.HasCTA = true
		cv.CTALabel, cv.CTAHref = vcard.ButtonText, vcard.Href
		if cv.CTALabel == "" {
			cv.CTALabel = t.CTA
		}
	}

	target := qrURL
	if target == "" {
		target = base
		if opts.Slug != "" {
			target = base + "/m/" + url.PathEscape(opts.Slug)
		}
	}
	cv.QRViewBox, cv.QRPath = cardQRShape(target)
	return cv, rest
}

func initialsOf(name string) string {
	var out []rune
	for _, w := range strings.Fields(name) {
		r, _ := utf8.DecodeRuneInString(w)
		out = append(out, []rune(strings.ToUpper(string(r)))...)
		if len(out) == 2 {
			break
		}
	}
	return string(out)
}

func websiteLabel(website string) string {
	s := strings.TrimSpace(website)
	for _, p := range []string{"https://", "http://"} {
		s = strings.TrimPrefix(s, p)
	}
	return strings.TrimSuffix(s, "/")
}

// ---- catalog for the API ----------------------------------------------------

// CardTemplateInfos lists the card templates as picker entries.
func CardTemplateInfos() []models.MicrositeTemplateInfo {
	out := make([]models.MicrositeTemplateInfo, 0, len(cardCatalog.Templates))
	for _, t := range cardCatalog.Templates {
		out = append(out, models.MicrositeTemplateInfo{
			ID: CardTemplateID(t.ID), Name: t.Name, Description: t.Description, Collection: t.Collection,
			Category: t.Category, Audience: t.Audience, Style: t.ID, CTA: t.CTA, Accent: cardDefaults[t.ID].Accent,
		})
	}
	return out
}

// CardTemplateID is the template id of a card style ("card_zenith").
func CardTemplateID(style string) string { return "card_" + style }

// cardStyleOfTemplate returns the card style for a template id, if it is a card template.
func cardStyleOfTemplate(templateID string) (*CardTemplate, bool) {
	if !strings.HasPrefix(templateID, "card_") {
		return nil, false
	}
	return CardTemplateByID(strings.TrimPrefix(templateID, "card_"))
}

// CardPaletteIDs lists the palette ids of a card style, in order.
func CardPaletteIDs(style string) []string {
	t, ok := CardTemplateByID(style)
	if !ok {
		return nil
	}
	ids := make([]string, 0, len(t.Palettes))
	for _, p := range t.Palettes {
		ids = append(ids, p.ID)
	}
	return ids
}

// cardSampleContent is the starter page of a card template.
func cardSampleContent(t *CardTemplate, displayName string) models.MicrositeContent {
	d := cardDefaults[t.ID]
	sample := cardSamples[t.ID]
	theme := styledTheme(t.ID, "system", d.Accent, d.BG, d.Ink, d.OnAccent)
	theme.Palette = t.Palettes[0].ID
	return models.MicrositeContent{
		Theme: theme,
		Blocks: []models.MicrositeBlock{
			mustBlock("blk_profile", models.BlockTypeProfileHeader, models.ProfileHeaderData{
				DisplayName: displayName,
				Title:       sample.Role,
				Company:     "Your Company",
				Bio:         "One or two lines about you or your company.",
				Tags:        sample.Tags,
			}),
			mustBlock("blk_vcard", models.BlockTypeVCardAction, models.VCardActionData{
				ButtonText:   t.CTA,
				FirstName:    displayName,
				Organization: "Your Company",
				Title:        sample.Role,
				PhoneMobile:  "+10000000000",
				EmailWork:    "you@example.com",
				Website:      "https://example.com/",
				City:         "Your City",
				Country:      "Your Country",
			}),
			mustBlock("blk_contact", models.BlockTypeContactButtons, models.ContactButtonsData{Items: []models.ContactButton{
				{ID: "itm_whatsapp", Kind: "whatsapp", Value: "+10000000000", Label: "WhatsApp"},
			}}),
			mustBlock("blk_social", models.BlockTypeSocialIcons, models.SocialIconsData{Items: []models.SocialLink{
				{ID: "itm_linkedin", Platform: "linkedin", URL: "https://www.linkedin.com/"},
				{ID: "itm_instagram", Platform: "instagram", URL: "https://www.instagram.com/"},
				{ID: "itm_x", Platform: "x", URL: "https://x.com/"},
				{ID: "itm_youtube", Platform: "youtube", URL: "https://www.youtube.com/"},
			}}),
			mustBlock("blk_qr", models.BlockTypeQRCode, models.QRCodeData{}),
		},
	}
}
