package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/URLshorter/url-shortener/internal/models"
)

// Microsite content limits
const (
	MaxMicrositeBlocks       = 50
	MaxMicrositeContentBytes = 100 * 1024
	maxLinkItems             = 30
	maxSocialItems           = 20
	maxContactButtons        = 6
	maxURLLength             = 2048
	currentBlockSchema       = 1
)

// MicrositeValidationError reports which field failed validation.
type MicrositeValidationError struct {
	Field   string
	Message string
}

func (e *MicrositeValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

func validationErr(field, format string, args ...interface{}) error {
	return &MicrositeValidationError{Field: field, Message: fmt.Sprintf(format, args...)}
}

var (
	slugPattern       = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,30}[a-z0-9]$`)
	idPattern         = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)
	hexColorPattern   = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	phonePattern      = regexp.MustCompile(`^\+?[0-9 ().-]{4,30}$`)
	localAssetPattern = regexp.MustCompile(`^/uploads/m/[0-9]{1,20}/[a-f0-9]{64}\.(jpg|png)$`)
)

// reservedMicrositeSlugs cannot be claimed as page slugs.
var reservedMicrositeSlugs = map[string]bool{
	"api": true, "admin": true, "app": true, "assets": true, "static": true,
	"preview": true, "vcard": true, "new": true, "edit": true, "create": true,
	"login": true, "register": true, "dashboard": true, "settings": true,
	"billing": true, "pricing": true, "help": true, "support": true,
	"trunc": true, "www": true, "mail": true, "root": true, "null": true,
	"undefined": true, "microsites": true, "pages": true,
}

// MicrositeFonts maps allowed font keys to CSS font stacks. Only system fonts
// are used so pages need no third-party font requests.
var MicrositeFonts = map[string]string{
	"system":    `system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif`,
	"serif":     `Georgia, Cambria, "Times New Roman", Times, serif`,
	"mono":      `ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace`,
	"rounded":   `ui-rounded, "SF Pro Rounded", "Hiragino Maru Gothic ProN", Quicksand, Nunito, system-ui, sans-serif`,
	"humanist":  `Seravek, "Gill Sans Nova", Ubuntu, Calibri, "DejaVu Sans", source-sans-pro, sans-serif`,
	"geometric": `Avenir, Montserrat, Corbel, "URW Gothic", source-sans-pro, sans-serif`,
}

var (
	allowedBackgroundTypes = map[string]bool{"solid": true, "linear": true, "radial": true, "image": true}
	allowedRadius          = map[string]bool{"sharp": true, "rounded": true, "pill": true}
	allowedShadow          = map[string]bool{"none": true, "soft": true, "strong": true}
	allowedAvatarBorder    = map[string]bool{"none": true, "thin": true, "thick": true}
	allowedContactKinds    = map[string]bool{"call": true, "sms": true, "email": true, "whatsapp": true}
	allowedTextAlign       = map[string]bool{"": true, "left": true, "center": true}
	allowedDividerStyles   = map[string]bool{"": true, "line": true, "space": true}
)

// MicrositeSocialPlatforms is the allowlist of social icon platforms.
var MicrositeSocialPlatforms = map[string]string{
	"linkedin":  "LinkedIn",
	"github":    "GitHub",
	"x":         "X",
	"instagram": "Instagram",
	"facebook":  "Facebook",
	"youtube":   "YouTube",
	"tiktok":    "TikTok",
	"threads":   "Threads",
	"website":   "Website",
}

// NormalizeMicrositeSlug lowercases and validates a page slug.
func NormalizeMicrositeSlug(raw string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	if !slugPattern.MatchString(slug) {
		return "", validationErr("slug", "must be 3-32 characters: lowercase letters, numbers and hyphens, not starting or ending with a hyphen")
	}
	if reservedMicrositeSlugs[slug] {
		return "", validationErr("slug", "%q is reserved", slug)
	}
	return slug, nil
}

// DefaultMicrositeTheme returns the theme used when none is supplied.
func DefaultMicrositeTheme() models.MicrositeTheme {
	return models.MicrositeTheme{
		Background:   models.MicrositeBackground{Type: "linear", Stops: []string{"#1e3c72", "#2a5298"}, Angle: 135},
		CardFill:     "#ffffff",
		Accent:       "#2563eb",
		ButtonText:   "#ffffff",
		Text:         "#111827",
		Font:         "system",
		ButtonRadius: "rounded",
		Shadow:       "soft",
		AvatarBorder: "thin",
	}
}

// NormalizeMicrositeContent validates every field of a page document for
// publishing and returns a normalized copy. Unknown block fields are dropped;
// empty theme fields fall back to defaults; any invalid or missing required
// value is rejected.
func NormalizeMicrositeContent(in models.MicrositeContent) (models.MicrositeContent, error) {
	return normalizeContent(in, false)
}

// NormalizeMicrositeDraft is the lenient variant used for saving work in
// progress: required fields may be left empty (e.g. a new link without a URL
// yet), but every non-empty value gets the same validation as on publish, so
// unsafe values (javascript: URLs, raw CSS) are still rejected.
func NormalizeMicrositeDraft(in models.MicrositeContent) (models.MicrositeContent, error) {
	return normalizeContent(in, true)
}

func normalizeContent(in models.MicrositeContent, draft bool) (models.MicrositeContent, error) {
	var out models.MicrositeContent

	theme, err := normalizeTheme(in.Theme, draft)
	if err != nil {
		return out, err
	}
	out.Theme = theme

	settings, err := normalizeSettings(in.Settings)
	if err != nil {
		return out, err
	}
	out.Settings = settings

	if len(in.Blocks) > MaxMicrositeBlocks {
		return out, validationErr("blocks", "a page can have at most %d blocks", MaxMicrositeBlocks)
	}
	seen := make(map[string]bool, len(in.Blocks))
	out.Blocks = make([]models.MicrositeBlock, 0, len(in.Blocks))
	for i, b := range in.Blocks {
		field := fmt.Sprintf("blocks[%d]", i)
		nb, err := normalizeBlock(b, field, draft)
		if err != nil {
			return out, err
		}
		if seen[nb.ID] {
			return out, validationErr(field+".id", "duplicate block id %q", nb.ID)
		}
		seen[nb.ID] = true
		out.Blocks = append(out.Blocks, nb)
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return out, err
	}
	if len(encoded) > MaxMicrositeContentBytes {
		return out, validationErr("content", "page content is too large (max %d KB)", MaxMicrositeContentBytes/1024)
	}
	return out, nil
}

func normalizeBlock(b models.MicrositeBlock, field string, draft bool) (models.MicrositeBlock, error) {
	id := strings.TrimSpace(b.ID)
	if id == "" {
		id = newMicrositeID("blk")
	}
	if !idPattern.MatchString(id) {
		return b, validationErr(field+".id", "invalid block id")
	}
	data, err := DecodeMicrositeBlockData(b.Type, b.Data, field+".data", draft)
	if err != nil {
		return b, err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return b, err
	}
	return models.MicrositeBlock{
		ID:            id,
		Type:          b.Type,
		SchemaVersion: currentBlockSchema,
		Hidden:        b.Hidden,
		Data:          encoded,
	}, nil
}

// DecodeMicrositeBlockData decodes and validates block data for its type,
// returning the normalized typed payload. With draft=true, empty required
// fields are allowed (renderers skip incomplete items); non-empty values are
// always fully validated.
func DecodeMicrositeBlockData(blockType string, raw json.RawMessage, field string, draft bool) (interface{}, error) {
	if len(raw) == 0 || string(raw) == "null" {
		raw = json.RawMessage(`{}`)
	}
	switch blockType {
	case models.BlockTypeProfileHeader:
		var d models.ProfileHeaderData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid profile header data")
		}
		return normalizeProfileHeader(d, field, draft)
	case models.BlockTypeVCardAction:
		var d models.VCardActionData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid contact card data")
		}
		return normalizeVCardAction(d, field, draft)
	case models.BlockTypeLinkList:
		var d models.LinkListData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid link list data")
		}
		return normalizeLinkList(d, field, draft)
	case models.BlockTypeContactButtons:
		var d models.ContactButtonsData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid contact buttons data")
		}
		return normalizeContactButtons(d, field, draft)
	case models.BlockTypeSocialIcons:
		var d models.SocialIconsData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid social icons data")
		}
		return normalizeSocialIcons(d, field, draft)
	case models.BlockTypeText:
		var d models.TextData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid text data")
		}
		return normalizeText(d, field, draft)
	case models.BlockTypeImage:
		var d models.ImageData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid image data")
		}
		return normalizeImage(d, field, draft)
	case models.BlockTypeDivider:
		var d models.DividerData
		if err := json.Unmarshal(raw, &d); err != nil {
			return nil, validationErr(field, "invalid divider data")
		}
		if !allowedDividerStyles[d.Style] {
			return nil, validationErr(field+".style", "must be line or space")
		}
		return d, nil
	default:
		return nil, validationErr(field, "unknown block type %q", blockType)
	}
}

// ---- per-block normalizers -------------------------------------------------

func normalizeProfileHeader(d models.ProfileHeaderData, field string, _ bool) (models.ProfileHeaderData, error) {
	var err error
	if d.DisplayName, err = cleanText(d.DisplayName, 80, false, field+".display_name"); err != nil {
		return d, err
	}
	if d.Pronouns, err = cleanText(d.Pronouns, 30, false, field+".pronouns"); err != nil {
		return d, err
	}
	if d.Title, err = cleanText(d.Title, 100, false, field+".title"); err != nil {
		return d, err
	}
	if d.Company, err = cleanText(d.Company, 100, false, field+".company"); err != nil {
		return d, err
	}
	if d.Bio, err = cleanText(d.Bio, 500, true, field+".bio"); err != nil {
		return d, err
	}
	if d.Location, err = cleanText(d.Location, 100, false, field+".location"); err != nil {
		return d, err
	}
	if d.AvatarURL, err = SanitizeMicrositeImageURL(d.AvatarURL, field+".avatar_url"); err != nil {
		return d, err
	}
	if d.CompanyLogoURL, err = SanitizeMicrositeImageURL(d.CompanyLogoURL, field+".company_logo_url"); err != nil {
		return d, err
	}
	return d, nil
}

func normalizeVCardAction(d models.VCardActionData, field string, draft bool) (models.VCardActionData, error) {
	var err error
	texts := []struct {
		v     *string
		max   int
		multi bool
		name  string
	}{
		{&d.ButtonText, 40, false, "button_text"},
		{&d.FirstName, 60, false, "first_name"},
		{&d.LastName, 60, false, "last_name"},
		{&d.Organization, 100, false, "organization"},
		{&d.Title, 100, false, "title"},
		{&d.Street, 120, false, "street"},
		{&d.City, 60, false, "city"},
		{&d.Region, 60, false, "region"},
		{&d.PostalCode, 20, false, "postal_code"},
		{&d.Country, 60, false, "country"},
		{&d.Note, 500, true, "note"},
	}
	for _, t := range texts {
		if *t.v, err = cleanText(*t.v, t.max, t.multi, field+"."+t.name); err != nil {
			return d, err
		}
	}
	if d.ButtonText == "" {
		d.ButtonText = "Save Contact"
	}
	if !draft && d.FirstName == "" && d.LastName == "" && d.Organization == "" {
		return d, validationErr(field+".first_name", "a name or organization is required for the contact card")
	}
	for _, p := range []struct {
		v    *string
		name string
	}{{&d.PhoneMobile, "phone_mobile"}, {&d.PhoneWork, "phone_work"}} {
		if *p.v, err = cleanPhone(*p.v, field+"."+p.name); err != nil {
			return d, err
		}
	}
	for _, e := range []struct {
		v    *string
		name string
	}{{&d.EmailWork, "email_work"}, {&d.EmailPersonal, "email_personal"}} {
		if *e.v, err = cleanEmail(*e.v, field+"."+e.name); err != nil {
			return d, err
		}
	}
	if d.Website != "" {
		if d.Website, err = sanitizeWebURL(d.Website, field+".website"); err != nil {
			return d, err
		}
	}
	if d.PhotoURL, err = SanitizeMicrositeImageURL(d.PhotoURL, field+".photo_url"); err != nil {
		return d, err
	}
	return d, nil
}

func normalizeLinkList(d models.LinkListData, field string, draft bool) (models.LinkListData, error) {
	if len(d.Items) > maxLinkItems {
		return d, validationErr(field+".items", "at most %d links per block", maxLinkItems)
	}
	seen := map[string]bool{}
	out := make([]models.LinkItem, 0, len(d.Items))
	for i, it := range d.Items {
		f := fmt.Sprintf("%s.items[%d]", field, i)
		var err error
		if it.ID, err = normalizeItemID(it.ID, seen, f); err != nil {
			return d, err
		}
		if it.Label, err = cleanText(it.Label, 80, false, f+".label"); err != nil {
			return d, err
		}
		if it.Label == "" && !draft {
			return d, validationErr(f+".label", "is required")
		}
		if it.URL, err = optionalURL(it.URL, draft, f+".url", SanitizeMicrositeLinkURL); err != nil {
			return d, err
		}
		if it.ThumbnailURL, err = SanitizeMicrositeImageURL(it.ThumbnailURL, f+".thumbnail_url"); err != nil {
			return d, err
		}
		out = append(out, it)
	}
	d.Items = out
	return d, nil
}

func normalizeContactButtons(d models.ContactButtonsData, field string, draft bool) (models.ContactButtonsData, error) {
	if len(d.Items) > maxContactButtons {
		return d, validationErr(field+".items", "at most %d contact buttons per block", maxContactButtons)
	}
	seen := map[string]bool{}
	out := make([]models.ContactButton, 0, len(d.Items))
	for i, it := range d.Items {
		f := fmt.Sprintf("%s.items[%d]", field, i)
		var err error
		if it.ID, err = normalizeItemID(it.ID, seen, f); err != nil {
			return d, err
		}
		if !allowedContactKinds[it.Kind] {
			return d, validationErr(f+".kind", "must be call, sms, email or whatsapp")
		}
		if it.Label, err = cleanText(it.Label, 40, false, f+".label"); err != nil {
			return d, err
		}
		switch it.Kind {
		case "email":
			it.Value, err = cleanEmail(it.Value, f+".value")
		case "whatsapp":
			it.Value, err = cleanWhatsApp(it.Value, f+".value")
		default:
			it.Value, err = cleanPhone(it.Value, f+".value")
		}
		if err != nil {
			return d, err
		}
		if it.Value == "" && !draft {
			return d, validationErr(f+".value", "is required")
		}
		out = append(out, it)
	}
	d.Items = out
	return d, nil
}

func normalizeSocialIcons(d models.SocialIconsData, field string, draft bool) (models.SocialIconsData, error) {
	if len(d.Items) > maxSocialItems {
		return d, validationErr(field+".items", "at most %d social links per block", maxSocialItems)
	}
	seen := map[string]bool{}
	out := make([]models.SocialLink, 0, len(d.Items))
	for i, it := range d.Items {
		f := fmt.Sprintf("%s.items[%d]", field, i)
		var err error
		if it.ID, err = normalizeItemID(it.ID, seen, f); err != nil {
			return d, err
		}
		if _, ok := MicrositeSocialPlatforms[it.Platform]; !ok {
			return d, validationErr(f+".platform", "unsupported platform %q", it.Platform)
		}
		if it.URL, err = optionalURL(it.URL, draft, f+".url", sanitizeWebURL); err != nil {
			return d, err
		}
		out = append(out, it)
	}
	d.Items = out
	return d, nil
}

func normalizeText(d models.TextData, field string, _ bool) (models.TextData, error) {
	var err error
	if d.Text, err = cleanText(d.Text, 2000, true, field+".text"); err != nil {
		return d, err
	}
	if !allowedTextAlign[d.Align] {
		return d, validationErr(field+".align", "must be left or center")
	}
	return d, nil
}

func normalizeImage(d models.ImageData, field string, draft bool) (models.ImageData, error) {
	var err error
	if d.URL, err = SanitizeMicrositeImageURL(d.URL, field+".url"); err != nil {
		return d, err
	}
	if d.URL == "" && !draft {
		return d, validationErr(field+".url", "is required")
	}
	if d.Alt, err = cleanText(d.Alt, 150, false, field+".alt"); err != nil {
		return d, err
	}
	if d.LinkURL != "" {
		if d.LinkURL, err = SanitizeMicrositeLinkURL(d.LinkURL, field+".link_url"); err != nil {
			return d, err
		}
	}
	return d, nil
}

// ---- theme & settings --------------------------------------------------------

func normalizeTheme(t models.MicrositeTheme, draft bool) (models.MicrositeTheme, error) {
	def := DefaultMicrositeTheme()
	var err error

	bg := t.Background
	if bg.Type == "" {
		bg = def.Background
	}
	if !allowedBackgroundTypes[bg.Type] {
		return t, validationErr("theme.background.type", "must be solid, linear, radial or image")
	}
	switch bg.Type {
	case "solid":
		if bg.Color, err = normalizeColor(bg.Color, "#1e3c72", "theme.background.color"); err != nil {
			return t, err
		}
		bg.Stops, bg.Angle, bg.ImageURL, bg.Overlay = nil, 0, "", 0
	case "linear", "radial":
		if len(bg.Stops) < 2 || len(bg.Stops) > 3 {
			return t, validationErr("theme.background.stops", "gradients need 2 or 3 colors")
		}
		stops := make([]string, len(bg.Stops))
		for i, s := range bg.Stops {
			if stops[i], err = normalizeColor(s, "", fmt.Sprintf("theme.background.stops[%d]", i)); err != nil {
				return t, err
			}
		}
		bg.Stops = stops
		if bg.Angle < 0 || bg.Angle > 360 {
			return t, validationErr("theme.background.angle", "must be between 0 and 360")
		}
		if bg.Type == "radial" {
			bg.Angle = 0
		}
		bg.Color, bg.ImageURL, bg.Overlay = "", "", 0
	case "image":
		if bg.ImageURL, err = SanitizeMicrositeImageURL(bg.ImageURL, "theme.background.image_url"); err != nil {
			return t, err
		}
		if bg.ImageURL == "" && !draft {
			return t, validationErr("theme.background.image_url", "is required for image backgrounds")
		}
		if bg.Color, err = normalizeColor(bg.Color, "#111827", "theme.background.color"); err != nil {
			return t, err
		}
		if bg.Overlay < 0 || bg.Overlay > 80 {
			return t, validationErr("theme.background.overlay", "must be between 0 and 80")
		}
		bg.Stops, bg.Angle = nil, 0
	}
	t.Background = bg

	colors := []struct {
		v    *string
		def  string
		name string
	}{
		{&t.CardFill, def.CardFill, "theme.card_fill"},
		{&t.Accent, def.Accent, "theme.accent"},
		{&t.ButtonText, def.ButtonText, "theme.button_text"},
		{&t.Text, def.Text, "theme.text"},
	}
	for _, c := range colors {
		if *c.v, err = normalizeColor(*c.v, c.def, c.name); err != nil {
			return t, err
		}
	}

	enums := []struct {
		v       *string
		def     string
		allowed map[string]bool
		name    string
	}{
		{&t.ButtonRadius, def.ButtonRadius, allowedRadius, "theme.button_radius"},
		{&t.Shadow, def.Shadow, allowedShadow, "theme.shadow"},
		{&t.AvatarBorder, def.AvatarBorder, allowedAvatarBorder, "theme.avatar_border"},
	}
	for _, e := range enums {
		if *e.v == "" {
			*e.v = e.def
		}
		if !e.allowed[*e.v] {
			return t, validationErr(e.name, "unsupported value %q", *e.v)
		}
	}
	if t.Font == "" {
		t.Font = def.Font
	}
	if _, ok := MicrositeFonts[t.Font]; !ok {
		return t, validationErr("theme.font", "unsupported font %q", t.Font)
	}
	return t, nil
}

func normalizeSettings(s models.MicrositeSettings) (models.MicrositeSettings, error) {
	var err error
	if s.SEOTitle, err = cleanText(s.SEOTitle, 70, false, "settings.seo_title"); err != nil {
		return s, err
	}
	if s.SEODescription, err = cleanText(s.SEODescription, 200, false, "settings.seo_description"); err != nil {
		return s, err
	}
	if s.OGImageURL, err = SanitizeMicrositeImageURL(s.OGImageURL, "settings.og_image_url"); err != nil {
		return s, err
	}
	if s.FaviconURL, err = SanitizeMicrositeImageURL(s.FaviconURL, "settings.favicon_url"); err != nil {
		return s, err
	}
	return s, nil
}

// ---- primitive sanitizers ----------------------------------------------------

// cleanText trims, strips control characters (keeping newlines when multiline)
// and enforces a maximum length in characters.
func cleanText(s string, maxRunes int, multiline bool, field string) (string, error) {
	if !utf8.ValidString(s) {
		return "", validationErr(field, "must be valid UTF-8")
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\n' && multiline:
			b.WriteRune(r)
		case r == '\t' || r == '\n' || r == '\r':
			if r != '\r' {
				b.WriteRune(' ')
			}
		case unicode.IsControl(r) || r == ' ' || r == ' ':
			// drop
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if utf8.RuneCountInString(out) > maxRunes {
		return "", validationErr(field, "must be at most %d characters", maxRunes)
	}
	return out, nil
}

func normalizeColor(v, def, field string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		if def == "" {
			return "", validationErr(field, "is required")
		}
		return def, nil
	}
	if !hexColorPattern.MatchString(v) {
		return "", validationErr(field, "must be a hex color like #1a2b3c")
	}
	return strings.ToLower(v), nil
}

func cleanPhone(v, field string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", nil
	}
	if !phonePattern.MatchString(v) {
		return "", validationErr(field, "must be a phone number")
	}
	return v, nil
}

func cleanEmail(v, field string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", nil
	}
	if len(v) > 254 {
		return "", validationErr(field, "is too long")
	}
	addr, err := mail.ParseAddress(v)
	if err != nil || addr.Address != v || addr.Name != "" {
		return "", validationErr(field, "must be an email address")
	}
	return v, nil
}

func cleanWhatsApp(v, field string) (string, error) {
	digits := PhoneDigits(v)
	if digits == "" {
		return "", nil
	}
	if len(digits) < 7 || len(digits) > 15 {
		return "", validationErr(field, "must be a WhatsApp number with country code")
	}
	return "+" + digits, nil
}

// PhoneDigits returns only the digits of a phone number.
func PhoneDigits(v string) string {
	var b strings.Builder
	for _, r := range v {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeItemID(id string, seen map[string]bool, field string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		id = newMicrositeID("itm")
	}
	if !idPattern.MatchString(id) {
		return "", validationErr(field+".id", "invalid item id")
	}
	if seen[id] {
		return "", validationErr(field+".id", "duplicate item id %q", id)
	}
	seen[id] = true
	return id, nil
}

func newMicrositeID(prefix string) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return prefix + "_0"
	}
	return prefix + "_" + hex.EncodeToString(b)
}

// optionalURL validates a URL, allowing it to be empty in drafts.
func optionalURL(raw string, draft bool, field string, sanitize func(string, string) (string, error)) (string, error) {
	if draft && strings.TrimSpace(raw) == "" {
		return "", nil
	}
	return sanitize(raw, field)
}

// sanitizeWebURL accepts only absolute http(s) URLs with a host.
func sanitizeWebURL(raw, field string) (string, error) {
	u, err := parseURLStrict(raw, field)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", validationErr(field, "must start with https://")
	}
	if u.Host == "" || u.User != nil {
		return "", validationErr(field, "must be a valid web address")
	}
	return u.String(), nil
}

// SanitizeMicrositeLinkURL accepts http(s), mailto:, tel: and sms: links and
// rejects everything else (javascript:, data:, vbscript:, relative URLs...).
func SanitizeMicrositeLinkURL(raw, field string) (string, error) {
	u, err := parseURLStrict(raw, field)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http", "https":
		if u.Host == "" || u.User != nil {
			return "", validationErr(field, "must be a valid web address")
		}
		return u.String(), nil
	case "mailto":
		if _, err := cleanEmail(u.Opaque, field); err != nil || u.Opaque == "" {
			return "", validationErr(field, "must be a valid mailto: address")
		}
		return "mailto:" + u.Opaque, nil
	case "tel", "sms":
		if !phonePattern.MatchString(u.Opaque) {
			return "", validationErr(field, "must be a valid phone link")
		}
		return u.Scheme + ":" + u.Opaque, nil
	default:
		return "", validationErr(field, "links must use https://, mailto:, tel: or sms:")
	}
}

// SanitizeMicrositeImageURL accepts an https image URL or a path to an image
// uploaded through the microsite asset endpoint. Empty is allowed.
func SanitizeMicrositeImageURL(raw, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if strings.HasPrefix(raw, "/") {
		if localAssetPattern.MatchString(raw) {
			return raw, nil
		}
		return "", validationErr(field, "must be an https image URL or an uploaded image")
	}
	u, err := parseURLStrict(raw, field)
	if err != nil {
		return "", err
	}
	if u.Scheme != "https" || u.Host == "" || u.User != nil {
		return "", validationErr(field, "images must use an https:// URL")
	}
	return u.String(), nil
}

func parseURLStrict(raw, field string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, validationErr(field, "is required")
	}
	if len(raw) > maxURLLength {
		return nil, validationErr(field, "is too long")
	}
	for _, r := range raw {
		if r <= 0x20 || r == 0x7f || r == '<' || r == '>' || r == '"' || r == '\\' {
			return nil, validationErr(field, "contains invalid characters")
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, validationErr(field, "is not a valid URL")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	return u, nil
}
