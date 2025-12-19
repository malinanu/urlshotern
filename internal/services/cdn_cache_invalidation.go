package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CDNProvider represents different CDN providers
type CDNProvider string

const (
	CDNProviderCloudFlare CDNProvider = "cloudflare"
	CDNProviderCloudFront CDNProvider = "cloudfront"
	CDNProviderFastly     CDNProvider = "fastly"
)

// CDNInvalidator handles cache invalidation across different CDN providers
type CDNInvalidator struct {
	provider           CDNProvider
	cloudflareAPIToken string
	cloudflareZoneID   string
	httpClient         *http.Client
}

// NewCDNInvalidator creates a new CDN cache invalidator
func NewCDNInvalidator(provider CDNProvider, apiToken, zoneID string) *CDNInvalidator {
	return &CDNInvalidator{
		provider:           provider,
		cloudflareAPIToken: apiToken,
		cloudflareZoneID:   zoneID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// InvalidateShortURL invalidates cache for a specific short URL
func (c *CDNInvalidator) InvalidateShortURL(ctx context.Context, shortCode string, baseURL string) error {
	switch c.provider {
	case CDNProviderCloudFlare:
		return c.invalidateCloudFlare(ctx, []string{
			fmt.Sprintf("%s/%s", baseURL, shortCode),
		})
	case CDNProviderCloudFront:
		return c.invalidateCloudFront(ctx, []string{fmt.Sprintf("/%s", shortCode)})
	case CDNProviderFastly:
		return c.invalidateFastly(ctx, shortCode)
	default:
		return fmt.Errorf("unsupported CDN provider: %s", c.provider)
	}
}

// InvalidateMultiple invalidates cache for multiple URLs
func (c *CDNInvalidator) InvalidateMultiple(ctx context.Context, urls []string) error {
	switch c.provider {
	case CDNProviderCloudFlare:
		return c.invalidateCloudFlare(ctx, urls)
	case CDNProviderCloudFront:
		return c.invalidateCloudFront(ctx, urls)
	default:
		return fmt.Errorf("batch invalidation not implemented for provider: %s", c.provider)
	}
}

// InvalidateByTags invalidates cache using cache tags (CloudFlare specific)
func (c *CDNInvalidator) InvalidateByTags(ctx context.Context, tags []string) error {
	if c.provider != CDNProviderCloudFlare {
		return fmt.Errorf("tag-based invalidation only supported for CloudFlare")
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.cloudflareZoneID)

	payload := map[string]interface{}{
		"tags": tags,
	}

	return c.sendCloudFlareRequest(ctx, url, payload)
}

// InvalidatePrefix invalidates all URLs with a given prefix (e.g., /api/*)
func (c *CDNInvalidator) InvalidatePrefix(ctx context.Context, prefix string) error {
	if c.provider != CDNProviderCloudFlare {
		return fmt.Errorf("prefix invalidation not supported for provider: %s", c.provider)
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.cloudflareZoneID)

	payload := map[string]interface{}{
		"prefixes": []string{prefix},
	}

	return c.sendCloudFlareRequest(ctx, url, payload)
}

// InvalidateEverything purges the entire cache (use with caution!)
func (c *CDNInvalidator) InvalidateEverything(ctx context.Context) error {
	switch c.provider {
	case CDNProviderCloudFlare:
		url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.cloudflareZoneID)
		payload := map[string]interface{}{
			"purge_everything": true,
		}
		return c.sendCloudFlareRequest(ctx, url, payload)
	default:
		return fmt.Errorf("purge everything not implemented for provider: %s", c.provider)
	}
}

// CloudFlare-specific invalidation
func (c *CDNInvalidator) invalidateCloudFlare(ctx context.Context, files []string) error {
	if len(files) == 0 {
		return nil
	}

	// CloudFlare allows up to 30 URLs per request
	batchSize := 30
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.cloudflareZoneID)

		payload := map[string]interface{}{
			"files": batch,
		}

		if err := c.sendCloudFlareRequest(ctx, url, payload); err != nil {
			return fmt.Errorf("failed to invalidate batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

func (c *CDNInvalidator) sendCloudFlareRequest(ctx context.Context, url string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cloudflareAPIToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		Success bool     `json:"success"`
		Errors  []string `json:"errors"`
		Result  struct {
			ID string `json:"id"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("cloudflare API error: %v", result.Errors)
	}

	return nil
}

// CloudFront-specific invalidation
func (c *CDNInvalidator) invalidateCloudFront(ctx context.Context, paths []string) error {
	// TODO: Implement AWS CloudFront invalidation using AWS SDK
	// This would require:
	// 1. AWS credentials configuration
	// 2. CloudFront distribution ID
	// 3. AWS SDK for Go integration
	return fmt.Errorf("cloudfront invalidation not yet implemented")
}

// Fastly-specific invalidation (uses surrogate keys)
func (c *CDNInvalidator) invalidateFastly(ctx context.Context, surrogateKey string) error {
	// TODO: Implement Fastly cache invalidation
	// Fastly uses surrogate keys for efficient cache invalidation
	return fmt.Errorf("fastly invalidation not yet implemented")
}

// CacheWarmingService preloads cache for popular URLs
type CacheWarmingService struct {
	invalidator *CDNInvalidator
	httpClient  *http.Client
}

// NewCacheWarmingService creates a new cache warming service
func NewCacheWarmingService(invalidator *CDNInvalidator) *CacheWarmingService {
	return &CacheWarmingService{
		invalidator: invalidator,
		httpClient: &http.Client{
			Timeout:       5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
	}
}

// WarmCache preloads cache by making requests to URLs
func (c *CacheWarmingService) WarmCache(ctx context.Context, urls []string) error {
	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		if err != nil {
			continue
		}

		// Make request to trigger cache population
		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		// Small delay to avoid overwhelming the origin
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

// WarmPopularURLs preloads cache for most frequently accessed URLs
func (c *CacheWarmingService) WarmPopularURLs(ctx context.Context, shortCodes []string, baseURL string) error {
	urls := make([]string, len(shortCodes))
	for i, code := range shortCodes {
		urls[i] = fmt.Sprintf("%s/%s", baseURL, code)
	}

	return c.WarmCache(ctx, urls)
}
