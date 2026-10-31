package services

import (
	"strings"
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

func ptr(f float64) *float64 { return &f }

// v2Content has every v2 block, with an XSS payload in every free-text field.
func v2Content(style string) models.MicrositeContent {
	theme := DefaultMicrositeTheme()
	theme.Style = style
	if style != MicrositeStyleClassic {
		theme.Font = "hanken"
	}
	product := models.ProductItem{
		ID: "pi", Name: xss, Price: "$1" + xss[:10], ComparePrice: xss[:20], Badge: xss[:20], Description: xss,
		BuyURL: "https://shop.example.com/?q=%22%3E", ButtonText: xss[:30], ImageURL: "https://cdn.example.com/p.png",
	}
	second := product
	second.ID = "pj"
	return models.MicrositeContent{
		Theme: theme,
		Blocks: []models.MicrositeBlock{
			rawBlock("hdr", models.BlockTypeProfileHeader, models.ProfileHeaderData{DisplayName: xss}),
			rawBlock("lnk", models.BlockTypeLinkList, models.LinkListData{Items: []models.LinkItem{
				{ID: "l1", Label: xss, URL: "https://example.com/a", Highlight: true},
			}}),
			rawBlock("pc", models.BlockTypeProductCard, product),
			rawBlock("car", models.BlockTypeProductCarousel, models.ProductCarouselData{Title: xss, Items: []models.ProductItem{product, second}}),
			rawBlock("ev", models.BlockTypeEventDetails, models.EventDetailsData{
				Title: xss, Start: "2030-05-01T18:00", End: "2030-05-01T21:00", Timezone: "Asia/Colombo",
				Venue: xss, Address: xss, TicketURL: "https://tickets.example.com/x", TicketText: xss[:30], ShowAddToCalendar: true,
			}),
			rawBlock("cd", models.BlockTypeCountdown, models.CountdownData{Target: "2030-05-01T18:00", Timezone: "Asia/Colombo", Label: xss[:40], DoneText: xss}),
			rawBlock("ag", models.BlockTypeAgenda, models.AgendaData{Title: xss, Items: []models.AgendaItem{
				{ID: "a1", Time: "09:00", Title: xss, Speaker: xss, Track: xss[:30], Detail: xss},
			}}),
			rawBlock("mp", models.BlockTypeMap, models.MapData{Label: xss, Address: xss, Lat: ptr(6.9271), Lng: ptr(79.8612)}),
			rawBlock("vd", models.BlockTypeVideo, models.VideoData{URL: "https://youtu.be/dQw4w9WgXcQ", Title: xss}),
			rawBlock("au", models.BlockTypeAudio, models.AudioData{URL: "https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC", Title: xss}),
			rawBlock("fm", models.BlockTypeForm, models.FormData{
				Kind: "contact", Title: xss, Description: xss, ButtonText: xss[:30], SuccessText: xss, ConsentText: xss, Notify: true,
			}),
		},
	}
}

func TestRenderV2EscapesUserContentInEveryStyle(t *testing.T) {
	for style := range MicrositeStyles {
		html := renderString(t, v2Content(style), MicrositeRenderOptions{
			BaseURL: "https://trunc.site", Slug: "x1", FormsEnabled: true,
			FormToken: func(string) string { return "tok" }, Now: time.Date(2030, 4, 1, 0, 0, 0, 0, time.UTC),
		})
		if strings.Contains(html, "<script>alert(1)") || strings.Contains(html, "onerror=alert(1)>") {
			t.Errorf("%s: unescaped payload in output", style)
		}
		if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
			t.Errorf("%s: expected escaped payload text", style)
		}
		// Every v2 block made it to the page.
		for _, want := range []string{`class="product"`, `class="carousel"`, `class="event"`, `class="countdown"`, `class="agenda"`,
			`class="map"`, `class="media media-youtube"`, `class="media media-spotify"`, `class="mform"`} {
			if !strings.Contains(html, want) {
				t.Errorf("%s: block %s missing", style, want)
			}
		}
	}
}

func TestRenderV2SuccessTextEscapedWhenSent(t *testing.T) {
	html := renderString(t, v2Content(MicrositeStyleProgramme), MicrositeRenderOptions{
		BaseURL: "https://trunc.site", Slug: "x1", FormsEnabled: true, SentBlock: "fm",
		FormToken: func(string) string { return "tok" },
	})
	if strings.Contains(html, "<script>alert(1)") {
		t.Fatal("success text rendered unescaped")
	}
	if strings.Contains(html, `name="email"`) {
		t.Error("form fields still shown after a successful submission")
	}
}

func TestRenderV2LoadsOnlyWhatThePageUses(t *testing.T) {
	opts := MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "x1"}

	// A classic v1 page pulls in no fonts, scripts beyond the beacon, or v2 CSS.
	v1 := models.MicrositeContent{Theme: DefaultMicrositeTheme(), Blocks: []models.MicrositeBlock{
		rawBlock("t", models.BlockTypeText, models.TextData{Text: "hi"}),
	}}
	html := renderString(t, v1, opts)
	for _, banned := range []string{"@font-face", "/m-static/embed.js", "/m-static/countdown.js", `rel="preload"`, "class=\"s-"} {
		if strings.Contains(html, banned) {
			t.Errorf("v1 page includes %s", banned)
		}
	}

	// A styled page preloads exactly its own face and only the scripts its blocks need.
	c := v2Content(MicrositeStyleProgramme)
	html = renderString(t, c, opts)
	if !strings.Contains(html, `<body class="s-programme"`) {
		t.Error("style body class missing")
	}
	if n := strings.Count(html, `rel="preload"`); n != len(MicrositeFontFaces["hanken"]) {
		t.Errorf("preloads = %d, want %d", n, len(MicrositeFontFaces["hanken"]))
	}
	if !strings.Contains(html, "https://trunc.site/m-static/fonts/") {
		t.Error("font URL should be absolute to the site")
	}
	for _, want := range []string{MicrositeEmbedJSPath, MicrositeCountdownJSPath} {
		if !strings.Contains(html, want) {
			t.Errorf("missing script %s", want)
		}
	}
	// Players are never framed on first load: click-to-load facades only.
	if strings.Contains(html, "<iframe") {
		t.Error("page renders an iframe before the visitor asks for it")
	}

	// The builder preview runs no scripts at all.
	html = renderString(t, c, MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "x1", Preview: true})
	if strings.Contains(html, "<script") {
		t.Error("preview includes a script")
	}
}

func TestRenderV2FormsFollowThePlan(t *testing.T) {
	c := v2Content(MicrositeStyleClassic)
	off := renderString(t, c, MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "x1", FormsEnabled: false})
	if strings.Contains(off, `id="fm"`) || strings.Contains(off, "/submit") {
		t.Error("form rendered although the owner's plan lacks forms")
	}

	on := renderString(t, c, MicrositeRenderOptions{
		BaseURL: "https://trunc.site", Slug: "x1", FormsEnabled: true,
		FormToken: func(id string) string { return "signed-" + id },
	})
	for _, want := range []string{`action="/m/x1/submit"`, `name="b" value="fm"`, `name="t" value="signed-fm"`, `name="website"`} {
		if !strings.Contains(on, want) {
			t.Errorf("form missing %s", want)
		}
	}

	// Error codes from the query string map to fixed messages; anything else
	// becomes the generic one and is never echoed.
	bad := renderString(t, c, MicrositeRenderOptions{
		BaseURL: "https://trunc.site", Slug: "x1", FormsEnabled: true, ErrorBlock: "fm", ErrorCode: "<b>pwned</b>",
		FormToken: func(string) string { return "t" },
	})
	if strings.Contains(bad, "pwned") {
		t.Error("error code reflected into the page")
	}
	if !strings.Contains(bad, FormErrorMessages["invalid"]) {
		t.Error("unknown error code should show the generic message")
	}
	known := renderString(t, c, MicrositeRenderOptions{
		BaseURL: "https://trunc.site", Slug: "x1", FormsEnabled: true, ErrorBlock: "fm", ErrorCode: "invalid_email",
		FormToken: func(string) string { return "t" },
	})
	if !strings.Contains(known, `role="alert"`) || !strings.Contains(known, "valid email address") {
		t.Error("known error code should render its message as an alert")
	}
}
