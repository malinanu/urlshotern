package services

import (
	"bytes"
	_ "embed" // go:embed for page template, stylesheet and beacon
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/URLshorter/url-shortener/internal/models"
)

var (
	//go:embed microsite_web/page.html.tmpl
	micrositePageTemplate string
	//go:embed microsite_web/page.css
	micrositeBaseCSS string
	// MicrositeBeaconJS is the click-tracking script served at /m-static/beacon.js.
	//go:embed microsite_web/beacon.js
	MicrositeBeaconJS []byte

	micrositeTmpl = template.Must(template.New("page").Parse(micrositePageTemplate))
)

// MicrositeBeaconPath is where the beacon script is served.
const MicrositeBeaconPath = "/m-static/beacon.js"

// Static inline SVG icons (constant markup, never user-controlled).
var micrositeIcons = map[string]template.HTML{
	"call":     `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>`,
	"sms":      `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`,
	"email":    `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>`,
	"whatsapp": `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"/></svg>`,
	"website":  `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>`,
}

var micrositeMonograms = map[string]string{
	"linkedin": "in", "github": "GH", "x": "X", "instagram": "IG", "facebook": "f",
	"youtube": "YT", "tiktok": "TT", "threads": "@",
}

var contactLabels = map[string]string{"call": "Call", "sms": "Text", "email": "Email", "whatsapp": "WhatsApp"}

// MicrositeRenderOptions controls how a page is rendered.
type MicrositeRenderOptions struct {
	BaseURL   string // absolute site URL, e.g. https://trunc.site
	Slug      string
	Title     string
	ShowBadge bool
	Preview   bool // builder preview: no tracking script, links open nothing
}

type micrositePageView struct {
	Title        string
	Description  string
	CanonicalURL string
	OGImage      string
	Favicon      string
	NoIndex      bool
	Preview      bool
	BaseCSS      template.CSS
	ThemeCSS     template.CSS
	BgImage      string
	Blocks       []micrositeBlockView
	ShowBadge    bool
	BadgeURL     string
	BeaconSrc    string
	EventURL     string
}

type micrositeBlockView struct {
	ID       string
	Type     string
	Profile  *profileView
	VCard    *vcardButtonView
	Links    []linkView
	Contacts []contactView
	Socials  []socialView
	Text     *models.TextData
	Image    *imageView
	Divider  string
}

type profileView struct {
	models.ProfileHeaderData
	Initial string
	Role    string
}

type vcardButtonView struct {
	ButtonText string
	Href       string
}

type linkView struct {
	BlockID   string
	ID        string
	Label     string
	Href      template.URL
	Highlight bool
	Thumb     string
	External  bool
}

type contactView struct {
	BlockID  string
	ID       string
	Kind     string
	Label    string
	Href     template.URL
	Icon     template.HTML
	External bool
}

type socialView struct {
	BlockID  string
	ID       string
	Name     string
	Href     string
	Icon     template.HTML
	Monogram string
}

type imageView struct {
	BlockID string
	Src     string
	Alt     string
	Href    template.URL
}

// RenderMicrosite renders a page document to HTML. Every block is re-validated
// at render time; blocks that fail validation are skipped rather than emitted.
func RenderMicrosite(content models.MicrositeContent, opts MicrositeRenderOptions) ([]byte, error) {
	base := strings.TrimRight(opts.BaseURL, "/")
	theme, err := normalizeTheme(content.Theme, true)
	if err != nil {
		theme = DefaultMicrositeTheme()
	}
	settings, err := normalizeSettings(content.Settings)
	if err != nil {
		settings = models.MicrositeSettings{}
	}

	view := micrositePageView{
		Preview:   opts.Preview,
		NoIndex:   opts.Preview || !settings.AllowIndexing,
		BaseCSS:   template.CSS(micrositeBaseCSS),         // #nosec G203 -- constant embedded stylesheet
		ThemeCSS:  template.CSS(micrositeThemeCSS(theme)), // #nosec G203 -- built only from validated hex colors and enum values
		ShowBadge: opts.ShowBadge,
		BadgeURL:  base + "/?ref=m",
		Favicon:   absoluteAssetURL(settings.FaviconURL, base),
	}
	if opts.Slug != "" {
		view.CanonicalURL = base + "/m/" + url.PathEscape(opts.Slug)
	}
	if theme.Background.Type == "image" {
		view.BgImage = absoluteAssetURL(theme.Background.ImageURL, base)
	}
	if !opts.Preview && opts.Slug != "" {
		view.BeaconSrc = MicrositeBeaconPath
		view.EventURL = "/api/v1/m/" + url.PathEscape(opts.Slug) + "/event"
	}

	var firstProfile *models.ProfileHeaderData
	for _, b := range content.Blocks {
		if b.Hidden {
			continue
		}
		bv, profile, ok := buildBlockView(b, base, opts)
		if !ok {
			continue
		}
		if profile != nil && firstProfile == nil {
			firstProfile = profile
		}
		view.Blocks = append(view.Blocks, bv)
	}

	view.Title = firstNonEmpty(settings.SEOTitle, opts.Title)
	if view.Title == "" && firstProfile != nil {
		view.Title = firstProfile.DisplayName
	}
	if view.Title == "" {
		view.Title = "Trunc Page"
	}
	view.Description = settings.SEODescription
	view.OGImage = absoluteAssetURL(settings.OGImageURL, base)
	if firstProfile != nil {
		if view.Description == "" {
			view.Description = firstNonEmpty(truncateRunes(firstProfile.Bio, 200), joinNonEmpty(" · ", firstProfile.Title, firstProfile.Company))
		}
		if view.OGImage == "" {
			view.OGImage = absoluteAssetURL(firstProfile.AvatarURL, base)
		}
	}

	var buf bytes.Buffer
	if err := micrositeTmpl.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("render microsite: %w", err)
	}
	return buf.Bytes(), nil
}

func buildBlockView(b models.MicrositeBlock, base string, opts MicrositeRenderOptions) (micrositeBlockView, *models.ProfileHeaderData, bool) {
	data, err := DecodeMicrositeBlockData(b.Type, b.Data, "block", true)
	if err != nil || !idPattern.MatchString(b.ID) {
		return micrositeBlockView{}, nil, false
	}
	bv := micrositeBlockView{ID: b.ID, Type: b.Type}

	switch d := data.(type) {
	case models.ProfileHeaderData:
		pv := &profileView{ProfileHeaderData: d, Role: joinNonEmpty(" · ", d.Title, d.Company)}
		pv.AvatarURL = absoluteAssetURL(d.AvatarURL, base)
		pv.CompanyLogoURL = absoluteAssetURL(d.CompanyLogoURL, base)
		if pv.AvatarURL == "" && d.DisplayName != "" {
			r, _ := utf8.DecodeRuneInString(d.DisplayName)
			pv.Initial = strings.ToUpper(string(r))
		}
		bv.Profile = pv
		return bv, &d, true

	case models.VCardActionData:
		href := "#"
		if opts.Slug != "" {
			href = "/m/" + url.PathEscape(opts.Slug) + "/vcard?b=" + url.QueryEscape(b.ID)
		}
		bv.VCard = &vcardButtonView{ButtonText: d.ButtonText, Href: href}

	case models.LinkListData:
		for _, it := range d.Items {
			href, err := SanitizeMicrositeLinkURL(it.URL, "url")
			if err != nil {
				continue
			}
			bv.Links = append(bv.Links, linkView{
				BlockID: b.ID, ID: it.ID, Label: it.Label, Highlight: it.Highlight,
				Href:     template.URL(href), // #nosec G203 -- scheme allowlisted by SanitizeMicrositeLinkURL
				Thumb:    absoluteAssetURL(it.ThumbnailURL, base),
				External: strings.HasPrefix(href, "http"),
			})
		}

	case models.ContactButtonsData:
		for _, it := range d.Items {
			href := contactHref(it)
			if href == "" {
				continue
			}
			bv.Contacts = append(bv.Contacts, contactView{
				BlockID: b.ID, ID: it.ID, Kind: it.Kind,
				Label:    firstNonEmpty(it.Label, contactLabels[it.Kind]),
				Href:     template.URL(href), // #nosec G203 -- built from validated phone/email values
				Icon:     micrositeIcons[it.Kind],
				External: it.Kind == "whatsapp",
			})
		}

	case models.SocialIconsData:
		for _, it := range d.Items {
			href, err := sanitizeWebURL(it.URL, "url")
			if err != nil {
				continue
			}
			bv.Socials = append(bv.Socials, socialView{
				BlockID: b.ID, ID: it.ID, Href: href,
				Name:     MicrositeSocialPlatforms[it.Platform],
				Icon:     micrositeIcons[it.Platform],
				Monogram: micrositeMonograms[it.Platform],
			})
		}

	case models.TextData:
		if d.Text == "" {
			return bv, nil, false
		}
		bv.Text = &d

	case models.ImageData:
		if d.URL == "" {
			return bv, nil, false
		}
		iv := &imageView{BlockID: b.ID, Src: absoluteAssetURL(d.URL, base), Alt: d.Alt}
		if d.LinkURL != "" {
			if href, err := SanitizeMicrositeLinkURL(d.LinkURL, "link_url"); err == nil {
				iv.Href = template.URL(href) // #nosec G203 -- scheme allowlisted
			}
		}
		bv.Image = iv

	case models.DividerData:
		bv.Divider = d.Style
	}
	return bv, nil, true
}

// contactHref builds the protocol-handler URL for a contact button.
func contactHref(it models.ContactButton) string {
	switch it.Kind {
	case "call":
		if p := telValue(it.Value); p != "" {
			return "tel:" + p
		}
	case "sms":
		if p := telValue(it.Value); p != "" {
			return "sms:" + p
		}
	case "email":
		if _, err := cleanEmail(it.Value, "value"); err == nil && it.Value != "" {
			return "mailto:" + it.Value
		}
	case "whatsapp":
		if d := PhoneDigits(it.Value); len(d) >= 7 && len(d) <= 15 {
			return "https://wa.me/" + d
		}
	}
	return ""
}

// telValue keeps a leading + and digits only.
func telValue(v string) string {
	digits := PhoneDigits(v)
	if digits == "" {
		return ""
	}
	if strings.HasPrefix(strings.TrimSpace(v), "+") {
		return "+" + digits
	}
	return digits
}

// micrositeThemeCSS turns a validated theme into CSS custom properties.
func micrositeThemeCSS(t models.MicrositeTheme) string {
	radius := map[string]string{"sharp": "4px", "rounded": "14px", "pill": "999px"}[t.ButtonRadius]
	cardRadius := map[string]string{"sharp": "6px", "rounded": "24px", "pill": "32px"}[t.ButtonRadius]
	shadow := map[string]string{
		"none":   "none",
		"soft":   "0 10px 30px rgba(0,0,0,.12)",
		"strong": "0 20px 50px rgba(0,0,0,.30)",
	}[t.Shadow]
	avatarBorder := map[string]string{"none": "0", "thin": "3px", "thick": "6px"}[t.AvatarBorder]

	bgColor := "#1e3c72"
	bgImage := "none"
	overlay := "0"
	switch t.Background.Type {
	case "solid":
		bgColor = t.Background.Color
	case "linear":
		bgColor = t.Background.Stops[0]
		bgImage = fmt.Sprintf("linear-gradient(%ddeg, %s)", t.Background.Angle, strings.Join(t.Background.Stops, ", "))
	case "radial":
		bgColor = t.Background.Stops[0]
		bgImage = fmt.Sprintf("radial-gradient(circle at 50%% 0%%, %s)", strings.Join(t.Background.Stops, ", "))
	case "image":
		bgColor = t.Background.Color
		overlay = strconv.FormatFloat(float64(t.Background.Overlay)/100, 'f', 2, 64)
	}

	var b strings.Builder
	fmt.Fprintf(&b, ":root{--bg:%s;--card:%s;--accent:%s;--accent-text:%s;--text:%s;", bgColor, t.CardFill, t.Accent, t.ButtonText, t.Text)
	fmt.Fprintf(&b, "--muted:%s;--soft:%s;--line:%s;", blendHex(t.Text, t.CardFill, 0.68), blendHex(t.Accent, t.CardFill, 0.10), blendHex(t.Text, t.CardFill, 0.15))
	fmt.Fprintf(&b, "--font:%s;--radius:%s;--card-radius:%s;--shadow:%s;--avatar-border:%s;--overlay:%s}", MicrositeFonts[t.Font], radius, cardRadius, shadow, avatarBorder, overlay)
	fmt.Fprintf(&b, "body{background-image:%s;background-attachment:fixed}", bgImage)
	return b.String()
}

// blendHex mixes color a over b with weight w (0..1) of a. Inputs are #rrggbb.
func blendHex(a, b string, w float64) string {
	pa, pb := parseHex(a), parseHex(b)
	var out [3]int
	for i := 0; i < 3; i++ {
		out[i] = int(float64(pa[i])*w + float64(pb[i])*(1-w) + 0.5)
	}
	return fmt.Sprintf("#%02x%02x%02x", out[0], out[1], out[2])
}

func parseHex(h string) [3]int {
	var c [3]int
	if len(h) != 7 {
		return c
	}
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseUint(h[1+2*i:3+2*i], 16, 8)
		if err == nil {
			c[i] = int(v)
		}
	}
	return c
}

// absoluteAssetURL turns an uploaded-asset path into an absolute URL so it
// resolves inside the builder's srcdoc preview and in OG tags.
func absoluteAssetURL(u, base string) string {
	if strings.HasPrefix(u, "/") {
		return base + u
	}
	return u
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func joinNonEmpty(sep string, vals ...string) string {
	parts := make([]string, 0, len(vals))
	for _, v := range vals {
		if v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, sep)
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:n-1])) + "…"
}
