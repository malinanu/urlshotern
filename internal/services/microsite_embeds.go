package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// EmbedRef is a parsed, validated third-party media reference. Embed URLs
// are always built from the provider and ID, never taken from user input.
type EmbedRef struct {
	Provider string // youtube | vimeo | spotify
	Kind     string // spotify: track | album | playlist | episode | show
	ID       string
}

var (
	youTubeIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	vimeoIDPattern   = regexp.MustCompile(`^[0-9]{6,12}$`)
	spotifyIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{22}$`)
	spotifyKinds     = map[string]bool{"track": true, "album": true, "playlist": true, "episode": true, "show": true}
)

// ParseVideoURL accepts YouTube and Vimeo page or share URLs.
func ParseVideoURL(raw string) (*EmbedRef, error) {
	u, err := parseEmbedURL(raw)
	if err != nil {
		return nil, err
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	segs := pathSegments(u.Path)
	switch host {
	case "youtube.com", "m.youtube.com", "music.youtube.com", "youtube-nocookie.com":
		id := ""
		switch {
		case len(segs) == 1 && segs[0] == "watch":
			id = u.Query().Get("v")
		case len(segs) == 2 && (segs[0] == "shorts" || segs[0] == "embed" || segs[0] == "live"):
			id = segs[1]
		}
		if youTubeIDPattern.MatchString(id) {
			return &EmbedRef{Provider: "youtube", ID: id}, nil
		}
	case "youtu.be":
		if len(segs) == 1 && youTubeIDPattern.MatchString(segs[0]) {
			return &EmbedRef{Provider: "youtube", ID: segs[0]}, nil
		}
	case "vimeo.com":
		if len(segs) >= 1 && vimeoIDPattern.MatchString(segs[len(segs)-1]) {
			return &EmbedRef{Provider: "vimeo", ID: segs[len(segs)-1]}, nil
		}
	case "player.vimeo.com":
		if len(segs) == 2 && segs[0] == "video" && vimeoIDPattern.MatchString(segs[1]) {
			return &EmbedRef{Provider: "vimeo", ID: segs[1]}, nil
		}
	}
	return nil, fmt.Errorf("must be a YouTube or Vimeo video link")
}

// ParseAudioURL accepts Spotify track, album, playlist, episode and show links.
func ParseAudioURL(raw string) (*EmbedRef, error) {
	u, err := parseEmbedURL(raw)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(u.Hostname()) != "open.spotify.com" {
		return nil, fmt.Errorf("must be an open.spotify.com link")
	}
	segs := pathSegments(u.Path)
	if len(segs) > 0 && strings.HasPrefix(segs[0], "intl-") {
		segs = segs[1:]
	}
	if len(segs) > 0 && segs[0] == "embed" {
		segs = segs[1:]
	}
	if len(segs) == 2 && spotifyKinds[segs[0]] && spotifyIDPattern.MatchString(segs[1]) {
		return &EmbedRef{Provider: "spotify", Kind: segs[0], ID: segs[1]}, nil
	}
	return nil, fmt.Errorf("must be a Spotify track, album, playlist, episode or show link")
}

// EmbedURL returns the player URL for the reference (autoplay on load,
// since the player is only inserted after the visitor taps play).
func (e *EmbedRef) EmbedURL() string {
	switch e.Provider {
	case "youtube":
		return "https://www.youtube-nocookie.com/embed/" + e.ID + "?autoplay=1&rel=0&modestbranding=1"
	case "vimeo":
		return "https://player.vimeo.com/video/" + e.ID + "?autoplay=1&dnt=1"
	case "spotify":
		return "https://open.spotify.com/embed/" + e.Kind + "/" + e.ID
	}
	return ""
}

// WatchURL returns the canonical page on the provider (the no-JS fallback).
func (e *EmbedRef) WatchURL() string {
	switch e.Provider {
	case "youtube":
		return "https://www.youtube.com/watch?v=" + e.ID
	case "vimeo":
		return "https://vimeo.com/" + e.ID
	case "spotify":
		return "https://open.spotify.com/" + e.Kind + "/" + e.ID
	}
	return ""
}

// ProviderName is the human name of the provider.
func (e *EmbedRef) ProviderName() string {
	return map[string]string{"youtube": "YouTube", "vimeo": "Vimeo", "spotify": "Spotify"}[e.Provider]
}

// EmbedHeight is the player height in px for audio embeds (0 for video, which uses 16:9).
func (e *EmbedRef) EmbedHeight() int {
	if e.Provider != "spotify" {
		return 0
	}
	if e.Kind == "track" || e.Kind == "episode" {
		return 152
	}
	return 352
}

func parseEmbedURL(raw string) (*url.URL, error) {
	u, err := parseURLStrict(raw, "url")
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("must be a web link")
	}
	if u.User != nil || u.Port() != "" {
		return nil, fmt.Errorf("must be a plain web link")
	}
	return u, nil
}

func pathSegments(p string) []string {
	var out []string
	for _, s := range strings.Split(p, "/") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
