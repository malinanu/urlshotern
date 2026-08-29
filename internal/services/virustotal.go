package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SpamStatus int

const (
	SpamStatusSafe SpamStatus = iota
	SpamStatusSpam
	SpamStatusUnknown
)

// CheckURLSpam queries the VirusTotal API to check if a URL is malicious.
func CheckURLSpam(targetUrl string, apiKey string) (SpamStatus, error) {
	if apiKey == "" {
		return SpamStatusSafe, nil // Skip check if no API key is provided
	}

	id := base64.RawURLEncoding.EncodeToString([]byte(targetUrl))
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.virustotal.com/api/v3/urls/%s", id), nil)
	if err != nil {
		return SpamStatusUnknown, fmt.Errorf("failed to create VirusTotal request: %w", err)
	}

	req.Header.Add("x-apikey", apiKey)
	req.Header.Add("accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Warning: Network error calling VirusTotal: %v\n", err)
		return SpamStatusUnknown, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
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
			return SpamStatusUnknown, fmt.Errorf("failed to parse VirusTotal response: %w", err)
		}

		stats := result.Data.Attributes.LastAnalysisStats
		if stats.Malicious > 0 || stats.Suspicious > 0 {
			return SpamStatusSpam, nil 
		}
		return SpamStatusSafe, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return SpamStatusUnknown, nil
	}

	if resp.StatusCode != http.StatusNotFound {
		return SpamStatusUnknown, fmt.Errorf("VirusTotal API returned unexpected status: %d", resp.StatusCode)
	}

	return SpamStatusUnknown, nil
}

