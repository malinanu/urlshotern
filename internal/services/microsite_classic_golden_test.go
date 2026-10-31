package services

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/URLshorter/url-shortener/internal/models"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite testdata/classic golden files")

// classicFixtures are v1 pages. Their rendering must never change: existing
// published pages keep the classic style byte for byte.
var classicFixtures = map[string]string{
	"card": `{"theme":{"background":{"type":"linear","stops":["#1e3c72","#2a5298"],"angle":135},"card_fill":"#ffffff","accent":"#2563eb","button_text":"#ffffff","text":"#111827","font":"system","button_radius":"rounded","shadow":"soft","avatar_border":"thin"},
	  "settings":{"hide_badge":false,"allow_indexing":false},
	  "blocks":[
	    {"id":"blk_profile","type":"profile_header","data":{"display_name":"Sasha Fields","title":"Brand Director","company":"Betty Baker","bio":"Line one\nLine two","location":"Denver, CO","pronouns":"she/her","verified":true}},
	    {"id":"blk_vcard","type":"vcard_action","data":{"button_text":"Save Contact","first_name":"Sasha","last_name":"Fields"}},
	    {"id":"blk_contact","type":"contact_buttons","data":{"items":[{"id":"c1","kind":"call","value":"+1 303 333 4444"},{"id":"c2","kind":"whatsapp","value":"+13033334444","label":"WhatsApp"},{"id":"c3","kind":"email","value":"s@example.com"},{"id":"c4","kind":"sms","value":"+13033334444"}]}},
	    {"id":"blk_social","type":"social_icons","data":{"items":[{"id":"s1","platform":"linkedin","url":"https://www.linkedin.com/in/x"},{"id":"s2","platform":"website","url":"https://example.com"}]}}
	  ]}`,
	"links": `{"theme":{"background":{"type":"radial","stops":["#134e5e","#71b280"]},"card_fill":"#f8faf5","accent":"#166534","button_text":"#ffffff","text":"#14251b","font":"serif","button_radius":"pill","shadow":"strong","avatar_border":"thick"},
	  "settings":{"hide_badge":true,"allow_indexing":true,"seo_title":"Creator","seo_description":"All my links","og_image_url":"https://cdn.example.com/og.jpg","favicon_url":"https://cdn.example.com/f.png"},
	  "blocks":[
	    {"id":"blk_profile","type":"profile_header","data":{"display_name":"Ari","avatar_url":"https://cdn.example.com/a.jpg","bio":"Creator"}},
	    {"id":"blk_links","type":"link_list","data":{"items":[{"id":"l1","label":"Latest","url":"https://youtube.com/@x","highlight":true,"thumbnail_url":"https://cdn.example.com/t.jpg"},{"id":"l2","label":"Mail me","url":"mailto:a@example.com"}]}},
	    {"id":"blk_text","type":"text","data":{"text":"Hello <there> & welcome","align":"center"}},
	    {"id":"blk_img","type":"image","data":{"url":"https://cdn.example.com/i.jpg","alt":"Photo","link_url":"https://example.com"}},
	    {"id":"blk_div","type":"divider","data":{"style":"space"}},
	    {"id":"blk_div2","type":"divider","data":{"style":"line"}},
	    {"id":"blk_hidden","type":"text","hidden":true,"data":{"text":"secret"}}
	  ]}`,
	"image-bg": `{"theme":{"background":{"type":"image","image_url":"https://cdn.example.com/bg.jpg","color":"#111827","overlay":40},"font":"mono","button_radius":"sharp","shadow":"none","avatar_border":"none"},
	  "settings":{},"blocks":[{"id":"p","type":"profile_header","data":{"display_name":"Bo"}}]}`,
	"empty": `{"theme":{"background":{"type":"solid","color":"#0b1020"}},"settings":{},"blocks":[]}`,
}

func TestClassicRenderingIsUnchanged(t *testing.T) {
	for name, raw := range classicFixtures {
		var c models.MicrositeContent
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, mode := range []string{"public", "preview"} {
			opts := MicrositeRenderOptions{BaseURL: "https://trunc.site", Slug: "sasha", Title: "Sasha", ShowBadge: true, Preview: mode == "preview"}
			out, err := RenderMicrosite(c, opts)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join("testdata", "classic", name+"-"+mode+".html")
			if *updateGolden {
				_ = os.MkdirAll(filepath.Dir(path), 0o750)
				if err := os.WriteFile(path, out, 0o600); err != nil {
					t.Fatal(err)
				}
				continue
			}
			want, err := os.ReadFile(path) // #nosec G304 -- fixed test path
			if err != nil {
				t.Fatalf("missing golden %s (run with -update-golden on v1 code): %v", path, err)
			}
			if string(out) != string(want) {
				t.Errorf("%s/%s: classic output changed (%d bytes vs %d golden)", name, mode, len(out), len(want))
			}
		}
	}
}
