package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CheckURLSpam queries the VirusTotal API to check if a URL is malicious.
// It returns true if the URL is flagged as malicious or suspicious.
// If the URL is unknown (404), it returns false (assuming innocent until proven guilty).
func CheckURLSpam(targetUrl string, apiKey string) (bool, error) {
	if apiKey == "" {
		return false, nil // Skip check if no API key is provided
	}

	// The URL identifier must be base64url encoded without padding
	id := base64.RawURLEncoding.EncodeToString([]byte(targetUrl))

	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.virustotal.com/api/v3/urls/%s", id), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create VirusTotal request: %w", err)
	}

	req.Header.Add("x-apikey", apiKey)
	req.Header.Add("accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call VirusTotal API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// URL is unknown to VirusTotal, we allow it
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("VirusTotal API returned unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Attributes struct {
				LastAnalysisStats struct {
					Malicious  int `json:"malicious"`
					Suspicious int `json:"suspicious"`
				} `json:"last_analysis_stats"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to parse VirusTotal response: %w", err)
	}

	stats := result.Data.Attributes.LastAnalysisStats
	if stats.Malicious > 0 || stats.Suspicious > 0 {
		return true, nil // Flagged as spam/malicious
	}

	return false, nil
}
