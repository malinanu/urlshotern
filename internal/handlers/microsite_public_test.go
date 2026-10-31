package handlers

import (
	"net/url"
	"strings"
	"testing"

	"github.com/URLshorter/url-shortener/internal/services"
)

func cspDirective(policy, name string) string {
	for _, d := range strings.Split(policy, ";") {
		d = strings.TrimSpace(d)
		if strings.HasPrefix(d, name+" ") {
			return strings.TrimPrefix(d, name+" ")
		}
	}
	return ""
}

func TestMicrositePageCSP(t *testing.T) {
	p := micrositePageCSP("https://trunc.site")
	if s := cspDirective(p, "script-src"); s != "'self' https://trunc.site" {
		t.Errorf("script-src = %q; pages must not allow inline or third-party script", s)
	}
	for name, want := range map[string]string{
		"default-src":     "'none'",
		"frame-ancestors": "'none'",
		"object-src":      "'none'",
		"base-uri":        "'none'",
		"form-action":     "'self' https://trunc.site",
	} {
		if got := cspDirective(p, name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if !strings.Contains(micrositePageCSP("not a url"), "script-src 'self';") {
		t.Errorf("a bad base URL should fall back to 'self' only")
	}
}

// Every embed URL the renderer can produce must be frameable under the page
// policy, and nothing else may be.
func TestMicrositeFrameSrcMatchesEmbeds(t *testing.T) {
	allowed := map[string]bool{}
	for _, o := range strings.Fields(micrositeFrameSrc) {
		u, err := url.Parse(o)
		if err != nil || u.Scheme != "https" {
			t.Fatalf("frame-src entry %q must be an https origin", o)
		}
		allowed[u.Host] = true
	}

	var urls []string
	for _, raw := range []string{"https://youtu.be/dQw4w9WgXcQ", "https://vimeo.com/76979871"} {
		ref, err := services.ParseVideoURL(raw)
		if err != nil {
			t.Fatal(err)
		}
		urls = append(urls, ref.EmbedURL())
	}
	ref, err := services.ParseAudioURL("https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC")
	if err != nil {
		t.Fatal(err)
	}
	urls = append(urls, ref.EmbedURL(), "https://www.openstreetmap.org/export/embed.html?bbox=0")

	used := map[string]bool{}
	for _, raw := range urls {
		u, _ := url.Parse(raw)
		if !allowed[u.Host] {
			t.Errorf("embed host %s is not in frame-src", u.Host)
		}
		used[u.Host] = true
	}
	for h := range allowed {
		if !used[h] {
			t.Errorf("frame-src allows %s, which no embed uses", h)
		}
	}
}
