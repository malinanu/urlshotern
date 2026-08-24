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
// To conserve rate limits, if the URL is unknown it submits it asynchronously and allows it.
// If the rate limit (429) is hit, it allows the URL.
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
		// Log network error but allow the URL to pass
		fmt.Printf("Warning: Network error calling VirusTotal: %v\n", err)
		return false, nil
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

	if resp.StatusCode == http.StatusTooManyRequests {
		// Rate limit hit. Allow the URL to prevent blocking users
		fmt.Println("Warning: VirusTotal rate limit (429) hit. Allowing URL.")
		return false, nil
	}

	if resp.StatusCode != http.StatusNotFound {
		// Some other error (401, 403, 500)
		return false, fmt.Errorf("VirusTotal API returned unexpected status: %d", resp.StatusCode)
	}

	// 2. URL is unknown (404), submit it asynchronously for future checks
	go submitURLToVirusTotal(targetUrl, apiKey)

	// Assume safe for now so we don't block the user (and we can't poll due to rate limits)
	return false, nil
}

// submitURLToVirusTotal submits a URL to be scanned in the background
func submitURLToVirusTotal(targetUrl, apiKey string) {
	data := url.Values{}
	data.Set("url", targetUrl)
	
	postReq, err := http.NewRequest("POST", "https://www.virustotal.com/api/v3/urls", strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Error creating VT POST request: %v\n", err)
		return
	}
	
	postReq.Header.Add("x-apikey", apiKey)
	postReq.Header.Add("accept", "application/json")
	postReq.Header.Add("content-type", "application/x-www-form-urlencoded")
	
	client := &http.Client{Timeout: 10 * time.Second}
	postResp, err := client.Do(postReq)
	if err != nil {
		fmt.Printf("Error submitting URL to VT: %v\n", err)
		return
	}
	defer postResp.Body.Close()
	
	if postResp.StatusCode == http.StatusTooManyRequests {
		fmt.Println("Warning: VirusTotal rate limit hit during background submission.")
		return
	}
	
	if postResp.StatusCode != http.StatusOK {
		fmt.Printf("Warning: VirusTotal submission returned status %d\n", postResp.StatusCode)
	}
}
