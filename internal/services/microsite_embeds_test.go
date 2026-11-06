package services

import "testing"

func TestParseVideoURL(t *testing.T) {
	ok := map[string]string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ":        "youtube:dQw4w9WgXcQ",
		"https://youtube.com/watch?v=dQw4w9WgXcQ&t=42s":      "youtube:dQw4w9WgXcQ",
		"https://m.youtube.com/watch?v=dQw4w9WgXcQ":          "youtube:dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ?si=abc":                "youtube:dQw4w9WgXcQ",
		"https://www.youtube.com/shorts/dQw4w9WgXcQ":         "youtube:dQw4w9WgXcQ",
		"https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ": "youtube:dQw4w9WgXcQ",
		"https://vimeo.com/76979871":                         "vimeo:76979871",
		"https://vimeo.com/channels/staffpicks/76979871":     "vimeo:76979871",
		"https://player.vimeo.com/video/76979871":            "vimeo:76979871",
	}
	for in, want := range ok {
		ref, err := ParseVideoURL(in)
		if err != nil || ref.Provider+":"+ref.ID != want {
			t.Errorf("ParseVideoURL(%q) = %+v, %v; want %s", in, ref, err, want)
		}
	}
	bad := []string{
		"https://evil.com/watch?v=dQw4w9WgXcQ",
		"https://youtube.com.evil.com/watch?v=dQw4w9WgXcQ",
		"https://evilyoutube.com/watch?v=dQw4w9WgXcQ",
		"https://www.youtube.com/watch?v=short",
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ\"><script>",
		"javascript:alert(1)//youtube.com/watch?v=dQw4w9WgXcQ",
		"https://user@youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtube.com:8443/watch?v=dQw4w9WgXcQ",
		"https://vimeo.com/abc",
		"",
	}
	for _, in := range bad {
		if ref, err := ParseVideoURL(in); err == nil {
			t.Errorf("ParseVideoURL(%q) accepted: %+v", in, ref)
		}
	}
}

func TestParseAudioURL(t *testing.T) {
	ok := map[string]string{
		"https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC":              "track:4uLU6hMCjMI75M1A2tKUQC",
		"https://open.spotify.com/intl-de/album/4uLU6hMCjMI75M1A2tKUQC?si=x": "album:4uLU6hMCjMI75M1A2tKUQC",
		"https://open.spotify.com/embed/playlist/37i9dQZF1DXcBWIGoYBM5M":     "playlist:37i9dQZF1DXcBWIGoYBM5M",
		"https://open.spotify.com/episode/4uLU6hMCjMI75M1A2tKUQC":            "episode:4uLU6hMCjMI75M1A2tKUQC",
	}
	for in, want := range ok {
		ref, err := ParseAudioURL(in)
		if err != nil || ref.Kind+":"+ref.ID != want {
			t.Errorf("ParseAudioURL(%q) = %+v, %v; want %s", in, ref, err, want)
		}
	}
	for _, in := range []string{
		"https://spotify.com/track/4uLU6hMCjMI75M1A2tKUQC",
		"https://open.spotify.com/user/4uLU6hMCjMI75M1A2tKUQC",
		"https://open.spotify.com/track/short",
		"https://open.spotify.com.evil.com/track/4uLU6hMCjMI75M1A2tKUQC",
	} {
		if ref, err := ParseAudioURL(in); err == nil {
			t.Errorf("ParseAudioURL(%q) accepted: %+v", in, ref)
		}
	}
}

func TestEmbedURLsUseAllowedHostsOnly(t *testing.T) {
	refs := []EmbedRef{{Provider: "youtube", ID: "dQw4w9WgXcQ"}, {Provider: "vimeo", ID: "76979871"}, {Provider: "spotify", Kind: "track", ID: "4uLU6hMCjMI75M1A2tKUQC"}}
	want := []string{"https://www.youtube-nocookie.com/embed/", "https://player.vimeo.com/video/", "https://open.spotify.com/embed/track/"}
	for i, r := range refs {
		if got := r.EmbedURL(); len(got) < len(want[i]) || got[:len(want[i])] != want[i] {
			t.Errorf("EmbedURL = %q, want prefix %q", got, want[i])
		}
	}
}
