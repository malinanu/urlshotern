package services

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/URLshorter/url-shortener/internal/models"
)

func block(t *testing.T, id, typ string, data interface{}) models.MicrositeBlock {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return models.MicrositeBlock{ID: id, Type: typ, Data: raw}
}

func TestNormalizeMicrositeSlug(t *testing.T) {
	ok := map[string]string{
		"sasha":         "sasha",
		"  Sales-Club ": "sales-club",
		"a1b":           "a1b",
		"abc-123-xyz":   "abc-123-xyz",
	}
	for in, want := range ok {
		got, err := NormalizeMicrositeSlug(in)
		if err != nil || got != want {
			t.Errorf("NormalizeMicrositeSlug(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	bad := []string{"", "ab", "-abc", "abc-", "a.b.c", "abc/def", "abc_def", "über", strings.Repeat("a", 33),
		"admin", "api", "vcard", "preview", "../x", "a b c"}
	for _, in := range bad {
		if _, err := NormalizeMicrositeSlug(in); err == nil {
			t.Errorf("NormalizeMicrositeSlug(%q) accepted, want error", in)
		}
	}
}

func TestSanitizeMicrositeLinkURL(t *testing.T) {
	ok := map[string]string{
		"https://example.com/path?q=1": "https://example.com/path?q=1",
		"HTTP://Example.com":           "http://Example.com",
		"mailto:sasha@example.com":     "mailto:sasha@example.com",
		"tel:+13033334444":             "tel:+13033334444",
		"sms:+1-303-333-4444":          "sms:+1-303-333-4444",
	}
	for in, want := range ok {
		got, err := SanitizeMicrositeLinkURL(in, "url")
		if err != nil || got != want {
			t.Errorf("SanitizeMicrositeLinkURL(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	bad := []string{
		"javascript:alert(1)", "JaVaScRiPt:alert(1)", " javascript:alert(1)", "java\tscript:alert(1)", "sms:+1 (303) 333-4444",
		"data:text/html,<script>alert(1)</script>", "vbscript:msgbox(1)", "//evil.example", "/relative",
		"https://", "https://user:pass@example.com", "mailto:not-an-email", "tel:abc", "ftp://example.com",
		"https://example.com/\"onmouseover=\"x", "https://example.com/<script>", "",
	}
	for _, in := range bad {
		if got, err := SanitizeMicrositeLinkURL(in, "url"); err == nil {
			t.Errorf("SanitizeMicrositeLinkURL(%q) = %q, want error", in, got)
		}
	}
}

func TestSanitizeMicrositeImageURL(t *testing.T) {
	asset := "/uploads/m/42/" + strings.Repeat("ab", 32) + ".jpg"
	for _, in := range []string{"", "https://cdn.example.com/a.jpg", asset} {
		if _, err := SanitizeMicrositeImageURL(in, "img"); err != nil {
			t.Errorf("SanitizeMicrositeImageURL(%q) rejected: %v", in, err)
		}
	}
	for _, in := range []string{
		"http://cdn.example.com/a.jpg", "javascript:alert(1)", "data:image/svg+xml,<svg onload=alert(1)>",
		"/uploads/m/42/../../etc/passwd", "/uploads/m/42/" + strings.Repeat("ab", 32) + ".svg", "/static/x.png",
	} {
		if _, err := SanitizeMicrositeImageURL(in, "img"); err == nil {
			t.Errorf("SanitizeMicrositeImageURL(%q) accepted, want error", in)
		}
	}
}

func TestNormalizeThemeRejectsCSSInjection(t *testing.T) {
	bad := []models.MicrositeTheme{
		{Accent: "#fff;background:url(https://evil)"},
		{Accent: "red"},
		{Text: "#12345"},
		{Font: "Comic Sans'; } body { display:none"},
		{ButtonRadius: "999px"},
		{Background: models.MicrositeBackground{Type: "linear", Stops: []string{"#000000", "url(x)"}}},
		{Background: models.MicrositeBackground{Type: "linear", Stops: []string{"#000000"}}},
		{Background: models.MicrositeBackground{Type: "image", ImageURL: "javascript:alert(1)"}},
		{Background: models.MicrositeBackground{Type: "video"}},
		{Background: models.MicrositeBackground{Type: "linear", Stops: []string{"#000000", "#ffffff"}, Angle: 720}},
	}
	for i, th := range bad {
		if _, err := normalizeTheme(th, true); err == nil {
			t.Errorf("case %d: theme %+v accepted, want error", i, th)
		}
	}

	got, err := normalizeTheme(models.MicrositeTheme{Accent: "#00D2FF"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Accent != "#00d2ff" || got.Font != "system" || got.Background.Type != "linear" {
		t.Errorf("defaults/normalization not applied: %+v", got)
	}
}

func TestNormalizeMicrositeContent(t *testing.T) {
	in := models.MicrositeContent{
		Blocks: []models.MicrositeBlock{
			block(t, "blk_1", models.BlockTypeProfileHeader, map[string]interface{}{
				"display_name": "  Sasha\u0000 Fields ", "unknown_field": "dropped",
			}),
			block(t, "", models.BlockTypeLinkList, models.LinkListData{Items: []models.LinkItem{{Label: "Site", URL: "https://example.com"}}}),
		},
	}
	out, err := NormalizeMicrositeContent(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Blocks) != 2 {
		t.Fatalf("got %d blocks", len(out.Blocks))
	}
	if strings.Contains(string(out.Blocks[0].Data), "unknown_field") {
		t.Error("unknown fields should be dropped")
	}
	var ph models.ProfileHeaderData
	_ = json.Unmarshal(out.Blocks[0].Data, &ph)
	if ph.DisplayName != "Sasha Fields" {
		t.Errorf("display name = %q, want control chars stripped and trimmed", ph.DisplayName)
	}
	if out.Blocks[1].ID == "" || out.Blocks[1].SchemaVersion != currentBlockSchema {
		t.Errorf("block id/schema not assigned: %+v", out.Blocks[1])
	}
	var ll models.LinkListData
	_ = json.Unmarshal(out.Blocks[1].Data, &ll)
	if ll.Items[0].ID == "" {
		t.Error("item id not assigned")
	}
}

func TestNormalizeMicrositeContentErrors(t *testing.T) {
	cases := map[string]models.MicrositeContent{
		"unknown type": {Blocks: []models.MicrositeBlock{block(t, "b1", "raw_html", map[string]string{"html": "<script>"})}},
		"duplicate ids": {Blocks: []models.MicrositeBlock{
			block(t, "b1", models.BlockTypeDivider, models.DividerData{}),
			block(t, "b1", models.BlockTypeDivider, models.DividerData{}),
		}},
		"bad link": {Blocks: []models.MicrositeBlock{block(t, "b1", models.BlockTypeLinkList,
			models.LinkListData{Items: []models.LinkItem{{Label: "x", URL: "javascript:alert(1)"}}})}},
		"bad contact kind": {Blocks: []models.MicrositeBlock{block(t, "b1", models.BlockTypeContactButtons,
			models.ContactButtonsData{Items: []models.ContactButton{{Kind: "fax", Value: "+1234567"}}})}},
		"bad social platform": {Blocks: []models.MicrositeBlock{block(t, "b1", models.BlockTypeSocialIcons,
			models.SocialIconsData{Items: []models.SocialLink{{Platform: "myspace", URL: "https://myspace.com"}}})}},
		"vcard without name": {Blocks: []models.MicrositeBlock{block(t, "b1", models.BlockTypeVCardAction, models.VCardActionData{})}},
		"bad email": {Blocks: []models.MicrositeBlock{block(t, "b1", models.BlockTypeVCardAction,
			models.VCardActionData{FirstName: "A", EmailWork: "a@b@c"})}},
		"text too long": {Blocks: []models.MicrositeBlock{block(t, "b1", models.BlockTypeText,
			models.TextData{Text: strings.Repeat("x", 2001)})}},
		"bad block id": {Blocks: []models.MicrositeBlock{block(t, "b 1<", models.BlockTypeDivider, models.DividerData{})}},
	}
	for name, c := range cases {
		_, err := NormalizeMicrositeContent(c)
		var verr *MicrositeValidationError
		if !errors.As(err, &verr) {
			t.Errorf("%s: err = %v, want MicrositeValidationError", name, err)
		}
	}

	many := models.MicrositeContent{}
	for i := 0; i <= MaxMicrositeBlocks; i++ {
		many.Blocks = append(many.Blocks, block(t, "", models.BlockTypeDivider, models.DividerData{}))
	}
	if _, err := NormalizeMicrositeContent(many); err == nil {
		t.Error("expected error for too many blocks")
	}
}

func TestContactButtonsNormalizeWhatsApp(t *testing.T) {
	out, err := DecodeMicrositeBlockData(models.BlockTypeContactButtons,
		json.RawMessage(`{"items":[{"id":"w","kind":"whatsapp","value":"+1 (303) 333-4444"}]}`), "b", false)
	if err != nil {
		t.Fatal(err)
	}
	d := out.(models.ContactButtonsData)
	if d.Items[0].Value != "+13033334444" {
		t.Errorf("whatsapp value = %q", d.Items[0].Value)
	}
}

func TestTemplatesAreValid(t *testing.T) {
	for _, tmpl := range MicrositeTemplates {
		if _, err := NormalizeMicrositeContent(BuildMicrositeTemplate(tmpl.ID, "Sasha Fields")); err != nil {
			t.Errorf("template %s does not validate: %v", tmpl.ID, err)
		}
	}
}

func TestDraftAllowsIncompleteButNotUnsafe(t *testing.T) {
	incomplete := models.MicrositeContent{Blocks: []models.MicrositeBlock{
		block(t, "l", models.BlockTypeLinkList, models.LinkListData{Items: []models.LinkItem{{ID: "a", Label: "", URL: ""}}}),
		block(t, "c", models.BlockTypeContactButtons, models.ContactButtonsData{Items: []models.ContactButton{{ID: "b", Kind: "call"}}}),
		block(t, "s", models.BlockTypeSocialIcons, models.SocialIconsData{Items: []models.SocialLink{{ID: "s1", Platform: "github"}}}),
		block(t, "i", models.BlockTypeImage, models.ImageData{}),
		block(t, "v", models.BlockTypeVCardAction, models.VCardActionData{}),
	}}
	if _, err := NormalizeMicrositeDraft(incomplete); err != nil {
		t.Fatalf("draft with empty required fields rejected: %v", err)
	}
	if _, err := NormalizeMicrositeContent(incomplete); err == nil {
		t.Fatal("publish validation must reject incomplete content")
	}

	unsafe := models.MicrositeContent{Blocks: []models.MicrositeBlock{
		block(t, "l", models.BlockTypeLinkList, models.LinkListData{Items: []models.LinkItem{{ID: "a", Label: "x", URL: "javascript:alert(1)"}}}),
	}}
	if _, err := NormalizeMicrositeDraft(unsafe); err == nil {
		t.Fatal("draft must still reject unsafe URLs")
	}
	if _, err := NormalizeMicrositeDraft(models.MicrositeContent{Theme: models.MicrositeTheme{Accent: "red"}}); err == nil {
		t.Fatal("draft must still reject invalid theme values")
	}
}
