package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// indexNowEndpoint is the shared IndexNow submission API. Submitting here fans the
// request out to all participating engines (Bing, Yandex, Naver, Seznam, …).
const indexNowEndpoint = "https://api.indexnow.org/indexnow"

// IndexNowService notifies search engines that URLs were published or updated so
// they recrawl immediately instead of waiting for a scheduled crawl.
//
// It is intentionally best-effort: callers should invoke it from a background
// goroutine (see utils.Go) and treat a returned error as non-fatal.
type IndexNowService struct {
	key     string
	host    string // bare host, e.g. "trunc.site"
	siteURL string // origin with scheme, no trailing slash, e.g. "https://trunc.site"
	client  *http.Client
}

// NewIndexNowService builds a service from the public site URL and IndexNow key.
// If either is empty the service is disabled and all submissions become no-ops.
func NewIndexNowService(siteURL, key string) *IndexNowService {
	siteURL = strings.TrimRight(strings.TrimSpace(siteURL), "/")
	host := siteURL
	if i := strings.Index(host, "://"); i != -1 {
		host = host[i+3:]
	}

	return &IndexNowService{
		key:     strings.TrimSpace(key),
		host:    host,
		siteURL: siteURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Enabled reports whether the service is configured to submit.
func (s *IndexNowService) Enabled() bool {
	return s != nil && s.key != "" && s.host != "" && s.siteURL != ""
}

// SubmitPaths submits site-relative paths (e.g. "/pages/my-slug") or absolute URLs.
// Relative paths are resolved against the configured public site URL.
func (s *IndexNowService) SubmitPaths(paths []string) error {
	if !s.Enabled() {
		return nil
	}

	urls := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
			urls = append(urls, p)
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		urls = append(urls, s.siteURL+p)
	}

	if len(urls) == 0 {
		return nil
	}

	body, err := json.Marshal(map[string]interface{}{
		"host":        s.host,
		"key":         s.key,
		"keyLocation": fmt.Sprintf("%s/%s.txt", s.siteURL, s.key),
		"urlList":     urls,
	})
	if err != nil {
		return fmt.Errorf("indexnow: marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, indexNowEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("indexnow: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("indexnow: submit: %w", err)
	}
	defer resp.Body.Close()

	// 200 OK and 202 Accepted both mean the submission was received.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("indexnow: unexpected status %d", resp.StatusCode)
	}

	log.Printf("INFO [INDEXNOW] Submitted %d URL(s) (HTTP %d)", len(urls), resp.StatusCode)
	return nil
}
