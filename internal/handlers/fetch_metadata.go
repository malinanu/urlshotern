package handlers

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

type ogMetadata struct {
	OGTitle       string `json:"og_title"`
	OGDescription string `json:"og_description"`
	OGImage       string `json:"og_image"`
	Title         string `json:"title"`
}

// FetchURLMetadata fetches OG metadata from a URL and returns it as JSON.
// Used by the frontend to auto-fill OG fields when a destination URL is entered.
// Returns an empty object on any error so the user flow is never blocked.
func (h *Handler) FetchURLMetadata(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusOK, ogMetadata{})
		return
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusOK, ogMetadata{})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		c.JSON(http.StatusOK, ogMetadata{})
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TruncBot/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusOK, ogMetadata{})
		return
	}
	defer resp.Body.Close()

	// Limit body read to 512 KB to avoid reading huge pages
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		c.JSON(http.StatusOK, ogMetadata{})
		return
	}

	meta := parseOGMetadata(body)
	c.JSON(http.StatusOK, meta)
}

func parseOGMetadata(body []byte) ogMetadata {
	var meta ogMetadata
	tokenizer := html.NewTokenizer(strings.NewReader(string(body)))

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		token := tokenizer.Token()

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			switch token.Data {
			case "meta":
				var property, name, content string
				for _, attr := range token.Attr {
					switch attr.Key {
					case "property":
						property = attr.Val
					case "name":
						name = attr.Val
					case "content":
						content = attr.Val
					}
				}
				switch property {
				case "og:title":
					if meta.OGTitle == "" {
						meta.OGTitle = content
					}
				case "og:description":
					if meta.OGDescription == "" {
						meta.OGDescription = content
					}
				case "og:image":
					if meta.OGImage == "" {
						meta.OGImage = content
					}
				}
				// twitter:title / twitter:description as fallbacks
				if property == "twitter:title" && meta.OGTitle == "" {
					meta.OGTitle = content
				}
				if name == "description" && meta.OGDescription == "" {
					meta.OGDescription = content
				}

			case "title":
				if meta.Title == "" {
					// Read the text content of the <title> tag
					tokenizer.Next()
					t := tokenizer.Token()
					if t.Type == html.TextToken {
						meta.Title = strings.TrimSpace(t.Data)
					}
				}

			case "head":
				// skip — just keep parsing

			case "body":
				// Stop once we hit the body — all OG tags are in <head>
				return meta
			}
		}
	}
	return meta
}
