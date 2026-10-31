package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/URLshorter/url-shortener/internal/models"
)

const xss = `"><script>alert(1)</script><img src=x onerror=alert(1)>`

func rawBlock(id, typ string, data interface{}) models.MicrositeBlock {
	b, _ := json.Marshal(data)
	return models.MicrositeBlock{ID: id, Type: typ, Data: b}
}

func renderString(t *testing.T, c models.MicrositeContent, opts MicrositeRenderOptions) string {
	t.Helper()
	out, err := RenderMicrosite(c, opts)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestRenderMicrositeEscapesUserContent(t *testing.T) {
	c := models.MicrositeContent{
		Settings: models.MicrositeSettings{SEOTitle: "T" + xss[:20], SEODescription: xss},
		Blocks: []models.MicrositeBlock{
			rawBlock("p", models.BlockTypeProfileHeader, models.ProfileHeaderData{
				DisplayName: xss, Title: xss, Company: xss, Bio: xss, Location: xss, Pronouns: "x" + xss[:20],
			}),
			rawBlock("v", models.BlockTypeVCardAction, models.VCardActionData{ButtonText: xss[:40], FirstName: "A"}),
			rawBlock("l", models.BlockTypeLinkList, models.LinkListData{Items: []models.LinkItem{
				{ID: "i1", Label: xss, URL: "https://example.com/?q=" + "%22%3E%3Cscript%3E"},
			}}),
			rawBlock("t", models.BlockTypeText, models.TextData{Text: xss}),
			rawBlock("img", models.BlockTypeImage, models.ImageData{URL: "https://cdn.example.com/a.png", Alt: xss}),
		},
	}
	html := renderString(t, c, MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "sasha", Title: xss})
	if strings.Contains(html, "<script>alert(1)") || strings.Contains(html, "onerror=alert(1)>") {
		t.Fatalf("unescaped payload in output:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Error("expected escaped payload text in output")
	}
}

func TestRenderMicrositeDropsUnsafeLinks(t *testing.T) {
	c := models.MicrositeContent{Blocks: []models.MicrositeBlock{
		// Stored data that bypassed validation must still not render as a link.
		{ID: "l", Type: models.BlockTypeLinkList, Data: json.RawMessage(`{"items":[{"id":"a","label":"bad","url":"javascript:alert(1)"}]}`)},
		rawBlock("c", models.BlockTypeContactButtons, models.ContactButtonsData{Items: []models.ContactButton{
			{ID: "call", Kind: "call", Value: "+1 303 333 4444"},
			{ID: "wa", Kind: "whatsapp", Value: "+1 303 333 4444"},
			{ID: "mail", Kind: "email", Value: "sasha@example.com"},
		}}),
	}}
	html := renderString(t, c, MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "s1x"})
	if strings.Contains(html, "javascript:") {
		t.Error("javascript: URL rendered")
	}
	for _, want := range []string{`href="tel:&#43;13033334444"`, `href="https://wa.me/13033334444"`, `href="mailto:sasha@example.com"`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in output", want)
		}
	}
}

func TestRenderMicrositeMetaAndTracking(t *testing.T) {
	c := BuildMicrositeTemplate(MicrositeTemplateExecutive, "Sasha Fields")
	pub := renderString(t, c, MicrositeRenderOptions{BaseURL: "https://trunc.site/", Slug: "sasha", Title: "Sasha", ShowBadge: true})
	for _, want := range []string{
		`<link rel="canonical" href="https://trunc.site/m/sasha">`,
		`<meta name="robots" content="noindex, nofollow">`,
		`src="/m-static/beacon.js"`,
		`data-endpoint="/api/v1/m/sasha/event"`,
		`href="/m/sasha/vcard?b=blk_vcard"`,
		`Made with Trunc`,
		`data-track="call"`,
	} {
		if !strings.Contains(pub, want) {
			t.Errorf("published page missing %q", want)
		}
	}

	c.Settings.AllowIndexing = true
	preview := renderString(t, c, MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "sasha", Preview: true})
	if strings.Contains(preview, "beacon.js") {
		t.Error("preview must not include the tracking script")
	}
	if !strings.Contains(preview, `<base target="_blank">`) || !strings.Contains(preview, "noindex") {
		t.Error("preview should set base target and noindex")
	}
	if strings.Contains(preview, "Made with Trunc") {
		t.Error("badge rendered although ShowBadge=false")
	}
}

func TestMicrositeThemeCSSOnlyValidatedValues(t *testing.T) {
	th, err := normalizeTheme(models.MicrositeTheme{
		Background: models.MicrositeBackground{Type: "linear", Stops: []string{"#1e3c72", "#2a5298"}, Angle: 135},
		Accent:     "#00d2ff",
	})
	if err != nil {
		t.Fatal(err)
	}
	css := micrositeThemeCSS(th)
	for _, want := range []string{"--accent:#00d2ff", "linear-gradient(135deg, #1e3c72, #2a5298)", "--font:system-ui"} {
		if !strings.Contains(css, want) {
			t.Errorf("css missing %q: %s", want, css)
		}
	}
	// An invalid stored theme falls back to defaults rather than emitting raw values.
	out := renderString(t, models.MicrositeContent{Theme: models.MicrositeTheme{Accent: "red;}</style><script>x()</script>"}},
		MicrositeRenderOptions{BaseURL: "https://trunc.site"})
	if strings.Contains(out, "x()") {
		t.Error("invalid theme value leaked into CSS")
	}
}

func TestBlendHex(t *testing.T) {
	if got := blendHex("#000000", "#ffffff", 0.5); got != "#808080" {
		t.Errorf("blendHex = %s", got)
	}
	if got := blendHex("#ff0000", "#0000ff", 1); got != "#ff0000" {
		t.Errorf("blendHex = %s", got)
	}
}

func TestMicrositeBlockLabels(t *testing.T) {
	labels := MicrositeBlockLabels(BuildMicrositeTemplate(MicrositeTemplateCreator, "X"))
	if labels["blk_links/itm_latest"] != "My latest video" || labels["blk_social/itm_youtube"] != "YouTube" {
		t.Errorf("labels = %v", labels)
	}
}
