package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CheckURLSpam queries the VirusTotal API to check if a URL is malicious.
// It returns true if the URL is flagged as malicious or suspicious.
// If the URL is unknown, it submits it for scanning and polls for up to 10 seconds.
func CheckURLSpam(targetUrl string, apiKey string) (bool, error) {
	if apiKey == "" {
		return false, nil // Skip check if no API key is provided
	}

	// 1. Try to get existing URL report
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
			return false, fmt.Errorf("failed to parse VirusTotal response: %w", err)
		}

		stats := result.Data.Attributes.LastAnalysisStats
		if stats.Malicious > 0 || stats.Suspicious > 0 {
			return true, nil // Flagged as spam/malicious
		}
		return false, nil
	}

	// If not 404 and not 200, return error
	if resp.StatusCode != http.StatusNotFound {
		return false, fmt.Errorf("VirusTotal API returned unexpected status: %d", resp.StatusCode)
	}

	// 2. URL is unknown (404), submit it for real-time analysis
	data := url.Values{}
	data.Set("url", targetUrl)
	
	postReq, err := http.NewRequest("POST", "https://www.virustotal.com/api/v3/urls", strings.NewReader(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create VirusTotal POST request: %w", err)
	}
	
	postReq.Header.Add("x-apikey", apiKey)
	postReq.Header.Add("accept", "application/json")
	postReq.Header.Add("content-type", "application/x-www-form-urlencoded")
	
	postResp, err := client.Do(postReq)
	if err != nil {
		return false, fmt.Errorf("failed to submit URL to VirusTotal: %w", err)
	}
	defer postResp.Body.Close()
	
	if postResp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("VirusTotal submission returned status: %d", postResp.StatusCode)
	}
	
	var postResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(postResp.Body).Decode(&postResult); err != nil {
		return false, fmt.Errorf("failed to parse VirusTotal submission response: %w", err)
	}
	
	analysisID := postResult.Data.ID
	if analysisID == "" {
		return false, fmt.Errorf("VirusTotal submission returned empty analysis ID")
	}

	// 3. Poll for analysis results (up to 5 attempts, waiting 2s each = 10s timeout)
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		
		pollReq, err := http.NewRequest("GET", fmt.Sprintf("https://www.virustotal.com/api/v3/analyses/%s", analysisID), nil)
		if err != nil {
			return false, fmt.Errorf("failed to create VirusTotal poll request: %w", err)
		}
		pollReq.Header.Add("x-apikey", apiKey)
		pollReq.Header.Add("accept", "application/json")
		
		pollResp, err := client.Do(pollReq)
		if err != nil {
			return false, fmt.Errorf("failed to poll VirusTotal API: %w", err)
		}
		
		var pollResult struct {
			Data struct {
				Attributes struct {
					Status string `json:"status"`
					Stats  struct {
						Malicious  int `json:"malicious"`
						Suspicious int `json:"suspicious"`
					} `json:"stats"`
				} `json:"attributes"`
			} `json:"data"`
		}
		
		if err := json.NewDecoder(pollResp.Body).Decode(&pollResult); err != nil {
			pollResp.Body.Close()
			return false, fmt.Errorf("failed to parse VirusTotal poll response: %w", err)
		}
		pollResp.Body.Close()
		
		if pollResult.Data.Attributes.Status == "completed" {
			stats := pollResult.Data.Attributes.Stats
			if stats.Malicious > 0 || stats.Suspicious > 0 {
				return true, nil // Flagged as spam/malicious
			}
			return false, nil // Safe
		}
	}
	
	// Timeout reached (10s), assume safe to prevent blocking the user
	return false, nil
}
