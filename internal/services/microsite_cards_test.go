package services

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/URLshorter/url-shortener/internal/models"
)

func cardIDs(t *testing.T) []string {
	t.Helper()
	ids := make([]string, 0, len(cardCatalog.Templates))
	for _, tpl := range cardCatalog.Templates {
		ids = append(ids, tpl.ID)
	}
	if len(ids) != 12 {
		t.Fatalf("catalog has %d templates, want 12", len(ids))
	}
	return ids
}

func TestCardCatalogHasFourCollectionsOfThree(t *testing.T) {
	cardIDs(t)
	if len(cardCatalog.Collections) != 4 {
		t.Fatalf("%d collections, want 4", len(cardCatalog.Collections))
	}
	for _, c := range cardCatalog.Collections {
		n := 0
		for _, tpl := range cardCatalog.Templates {
			if tpl.Collection == c.ID {
				n++
			}
		}
		if n != 3 {
			t.Errorf("collection %s has %d templates, want 3", c.ID, n)
		}
	}
	for _, tpl := range cardCatalog.Templates {
		if len(tpl.Palettes) < 3 || tpl.CTA == "" || len(tpl.Fonts) == 0 {
			t.Errorf("%s: incomplete catalog entry", tpl.ID)
		}
		for _, key := range tpl.Fonts {
			f, ok := cardFontFaces[key]
			if !ok {
				t.Errorf("%s: unknown font %s", tpl.ID, key)
				continue
			}
			if _, ok := MicrositeFontFile(f.File); !ok {
				t.Errorf("%s: font file %s is not embedded", tpl.ID, f.File)
			}
		}
		if _, err := cardCSSFS.ReadFile("microsite_web/cards/" + tpl.ID + ".css"); err != nil {
			t.Errorf("%s: stylesheet missing", tpl.ID)
		}
	}
}

// The Go copies of the shared card assets must match the frontend source.
func TestCardAssetsMatchFrontendSource(t *testing.T) {
	const src = "../../frontend/src/components/micropages/cards/"
	if _, err := os.Stat(src + "catalog.json"); err != nil {
		t.Skip("frontend sources not present")
	}
	check := func(frontend, embedded string, got []byte) {
		want, err := os.ReadFile(frontend)
		if err != nil {
			t.Errorf("read %s: %v", frontend, err)
			return
		}
		if !bytes.Equal(want, got) {
			t.Errorf("%s differs from the embedded copy; run scripts/sync-card-assets.sh", embedded)
		}
	}
	check(src+"catalog.json", "cards/catalog.json", cardCatalogJSON)
	for _, name := range append([]string{"base"}, cardIDs(t)...) {
		got, err := cardCSSFS.ReadFile("microsite_web/cards/" + name + ".css")
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		check(src+"css/"+name+".css", "cards/"+name+".css", got)
	}
}

func publishedCard(t *testing.T, id string) (models.MicrositeContent, string) {
	t.Helper()
	content := BuildMicrositeTemplate(CardTemplateID(id), "Alexandra Reyes", "UTC")
	norm, err := NormalizeMicrositeContent(content)
	if err != nil {
		t.Fatalf("%s: sample page is not publishable: %v", id, err)
	}
	return norm, id
}

func TestEveryCardTemplateRendersAPublishablePage(t *testing.T) {
	for _, id := range cardIDs(t) {
		content, _ := publishedCard(t, id)
		html := renderString(t, content, MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "alex", PageID: 42, Title: "Alex"})
		for _, want := range []string{
			`class="vc vc-` + id + `"`, `data-vc="` + id + `"`, `<body class="s-card"`, `DIGITAL ID · TC-000042`,
			`class="vc-qr"`, `<path d="M`, `class="vc-cta"`, `href="/m/alex/vcard?b=blk_vcard"`, `class="vc-name"`,
			`src="/m-static/card-export.js"`, `class="vc-actions"`,
		} {
			if !strings.Contains(html, want) {
				t.Errorf("%s: output missing %q", id, want)
			}
		}
		// Only our own scripts, and no inline script.
		for _, m := range regexp.MustCompile(`<script[^>]*>`).FindAllString(html, -1) {
			if !strings.Contains(m, `src="/m-static/`) {
				t.Errorf("%s: unexpected script tag %s", id, m)
			}
		}
		if strings.Contains(html, "javascript:") {
			t.Errorf("%s: javascript: URL in output", id)
		}
		// The shared stylesheet and this template's stylesheet are inlined.
		if !strings.Contains(html, ".vc-"+id+" {") {
			t.Errorf("%s: template stylesheet missing", id)
		}
	}
}

func TestCardPreviewHasNoScriptsOrDownloads(t *testing.T) {
	content, _ := publishedCard(t, "zenith")
	html := renderString(t, content, MicrositeRenderOptions{BaseURL: "https://trunc.site", Preview: true})
	if strings.Contains(html, "<script") {
		t.Error("preview includes a script")
	}
	if strings.Contains(html, `class="vc-actions"`) {
		t.Error("preview shows the download actions")
	}
	if !strings.Contains(html, `class="vc-qr"`) || !strings.Contains(html, `<path d="M`) {
		t.Error("preview still shows a real QR code")
	}
}

func TestCardRenderEscapesUserContent(t *testing.T) {
	content, _ := publishedCard(t, "creo")
	for i, b := range content.Blocks {
		if b.Type != models.BlockTypeProfileHeader {
			continue
		}
		var d models.ProfileHeaderData
		_ = json.Unmarshal(b.Data, &d)
		d.DisplayName, d.Title, d.Company, d.Bio = xss, xss, xss, xss
		d.Tags = []string{xss[:20], xss[:20]}
		raw, _ := json.Marshal(d)
		content.Blocks[i].Data = raw
	}
	html := renderString(t, content, MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "alex"})
	if strings.Contains(html, "<script>alert(1)") || strings.Contains(html, "onerror=alert(1)>") {
		t.Fatal("unescaped payload in the card")
	}
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Error("expected escaped payload text")
	}
}

func TestCardQRShapeIsAValidMatrix(t *testing.T) {
	vb, path := cardQRShape("https://trunc.site/m/alexandra-reyes")
	if !strings.HasPrefix(vb, "-4 -4 ") || path == "" {
		t.Fatalf("viewBox %q path %d bytes", vb, len(path))
	}
	// Size 33 (version 4) + 8 for the quiet zone: the URL fits version 3 or 4 at level M.
	if vb != "-4 -4 37 37" && vb != "-4 -4 41 41" {
		t.Errorf("unexpected QR size %q", vb)
	}
	if !strings.HasPrefix(path, "M0 0h7v1h-7z") {
		t.Errorf("top-left finder pattern missing: %.40s", path)
	}
	if v, p := cardQRShape(""); v != "" || p != "" {
		t.Error("empty text must give no QR code")
	}
}

func TestCardPaletteAndAccent(t *testing.T) {
	theme := BuildMicrositeTemplate(CardTemplateID("zenith"), "A", "UTC").Theme
	if _, err := normalizeTheme(theme, false); err != nil {
		t.Fatal(err)
	}
	bad := theme
	bad.Palette = "nope"
	if _, err := normalizeTheme(bad, false); err == nil {
		t.Error("unknown palette accepted")
	}
	empty := theme
	empty.Palette = ""
	got, err := normalizeTheme(empty, false)
	if err != nil || got.Palette != "champagne" {
		t.Errorf("palette default = %q, %v", got.Palette, err)
	}
	classic := DefaultMicrositeTheme()
	classic.Palette = "champagne"
	if got, _ := normalizeTheme(classic, false); got.Palette != "" {
		t.Error("palette must be cleared for non-card styles")
	}

	tpl, _ := CardTemplateByID("zenith")
	rose := theme
	rose.Palette = "rose-gold"
	rose.Accent = tpl.paletteAccent("rose-gold")
	if s := string(cardStyleAttr(tpl, rose)); !strings.Contains(s, "--vc-accent:#d8a08c") || strings.Count(s, "--vc-accent:") != 1 {
		t.Errorf("palette style = %q", s)
	}
	custom := rose
	custom.Accent = "#19C3A6"
	if s := string(cardStyleAttr(tpl, custom)); !strings.Contains(s, "--vc-accent:#19c3a6;--vc-on-accent:#14110c;") {
		t.Errorf("custom accent style = %q", s)
	}
}

func TestQRCodeBlockValidation(t *testing.T) {
	ok := func(url string) bool {
		_, err := normalizeQRCode(models.QRCodeData{URL: url}, "block")
		return err == nil
	}
	for _, u := range []string{"", "https://example.com/menu", " http://example.com "} {
		if !ok(u) {
			t.Errorf("%q should be accepted", u)
		}
	}
	for _, u := range []string{"javascript:alert(1)", "mailto:a@b.co", "tel:+100", "ftp://example.com", "https://example.com/" + strings.Repeat("a", 420)} {
		if ok(u) {
			t.Errorf("%q should be rejected", u)
		}
	}
}

func TestProfileTagsAndLogoPlateValidation(t *testing.T) {
	d := models.ProfileHeaderData{DisplayName: "A", Tags: []string{"a", "b", "c", "d", "e"}}
	if _, err := normalizeProfileHeader(d, "block", false); err == nil {
		t.Error("5 tags accepted")
	}
	d = models.ProfileHeaderData{DisplayName: "A", Tags: []string{" Branding ", ""}, LogoPlate: "dark"}
	got, err := normalizeProfileHeader(d, "block", false)
	if err != nil || len(got.Tags) != 1 || got.Tags[0] != "Branding" {
		t.Errorf("tags = %v, %v", got.Tags, err)
	}
	d.LogoPlate = "purple"
	if _, err := normalizeProfileHeader(d, "block", false); err == nil {
		t.Error("bad logo plate accepted")
	}
}

func TestCardTemplatesAreListedForTheBuilder(t *testing.T) {
	found := 0
	for _, info := range MicrositeTemplates {
		if info.Style != "" {
			found++
			if !IsMicrositeTemplate(info.ID) || info.Collection == "" || info.Category == "" {
				t.Errorf("incomplete template info %+v", info)
			}
		}
	}
	if found != 12 {
		t.Errorf("%d card templates listed, want 12", found)
	}
	// The original five remain, first.
	if MicrositeTemplates[0].ID != MicrositeTemplateExecutive {
		t.Error("original templates must keep their place")
	}
}
