package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/storage"
)

// SpamAlertEmailer defines the interface for sending spam alerts
type SpamAlertEmailer interface {
	SendSpamAlert(toEmail, originalURL, shortCode string) error
}

type VirusTotalWorker struct {
	redis        *storage.RedisStorage
	db           *storage.PostgresStorage
	emailService SpamAlertEmailer
	apiKey       string
}

type VTTask struct {
	OriginalURL string `json:"original_url"`
	UserID      *int64 `json:"user_id"`
}

func NewVirusTotalWorker(redis *storage.RedisStorage, db *storage.PostgresStorage, emailService SpamAlertEmailer, apiKey string) *VirusTotalWorker {
	return &VirusTotalWorker{
		redis:        redis,
		db:           db,
		emailService: emailService,
		apiKey:       apiKey,
	}
}

func (w *VirusTotalWorker) Start(ctx context.Context) {
	if w.redis == nil || w.apiKey == "" {
		log.Println("VirusTotal worker skipping start: redis or API key not configured")
		return
	}

	log.Println("Starting VirusTotal background worker...")

	ticker := time.NewTicker(30 * time.Second) // Strict rate limit control
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("VirusTotal background worker shutting down")
			return
		case <-ticker.C:
			w.processNextTask()
		}
	}
}

func (w *VirusTotalWorker) processNextTask() {
	// Pop task from queue
	res, err := w.redis.RPop("vt_scan_queue")
	if err != nil {
		// Queue empty or redis error
		return
	}

	var task VTTask
	if err := json.Unmarshal([]byte(res), &task); err != nil {
		log.Printf("Error unmarshaling VT task: %v", err)
		return
	}

	log.Printf("Processing background VT scan for %s", task.OriginalURL)

	// Check VT
	isSpam, err := w.backgroundSubmitAndCheck(task.OriginalURL)
	if err != nil {
		log.Printf("Background VT check error for %s: %v", task.OriginalURL, err)
		// Requeue if we hit rate limit or network error
		w.redis.LPush("vt_scan_queue", res)
		return
	}

	if isSpam {
		log.Printf("ALERT: URL %s flagged as spam in background check!", task.OriginalURL)
		
		// Find URLs in DB and disable them
		rows, err := w.db.Query("SELECT short_code, user_id FROM url_mappings WHERE original_url = $1", task.OriginalURL)
		if err != nil {
			log.Printf("Error fetching URL mappings for %s: %v", task.OriginalURL, err)
			return
		}
		
		type targetURL struct {
			ShortCode string
			UserID    *int64
		}
		var targets []targetURL
		
		for rows.Next() {
			var t targetURL
			if err := rows.Scan(&t.ShortCode, &t.UserID); err != nil {
				log.Printf("Row scan error: %v", err)
				continue
			}
			targets = append(targets, t)
		}
		rows.Close()

		for _, t := range targets {
			// Disable URL
			_, err := w.db.Exec("UPDATE url_mappings SET is_active = false WHERE short_code = $1", t.ShortCode)
			if err != nil {
				log.Printf("Failed to disable URL %s: %v", t.ShortCode, err)
			} else {
				// Clear from cache
				w.redis.DeleteURLMapping(t.ShortCode)

				// Send email alert to user if they own it
				if t.UserID != nil {
					user, err := w.db.GetUserByID(*t.UserID)
					if err == nil && user != nil && user.Email != "" {
						err = w.emailService.SendSpamAlert(user.Email, task.OriginalURL, t.ShortCode)
						if err != nil {
							log.Printf("Failed to send spam alert email to %s: %v", user.Email, err)
						} else {
							log.Printf("Sent spam alert email to %s", user.Email)
						}
					}
				}
			}
		}
	}
}

func (w *VirusTotalWorker) backgroundSubmitAndCheck(targetUrl string) (bool, error) {
	data := url.Values{}
	data.Set("url", targetUrl)
	
	postReq, err := http.NewRequest("POST", "https://www.virustotal.com/api/v3/urls", strings.NewReader(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("error creating VT POST request: %w", err)
	}
	
	postReq.Header.Add("x-apikey", w.apiKey)
	postReq.Header.Add("accept", "application/json")
	postReq.Header.Add("content-type", "application/x-www-form-urlencoded")
	
	client := &http.Client{Timeout: 10 * time.Second}
	postResp, err := client.Do(postReq)
	if err != nil {
		return false, fmt.Errorf("error submitting URL to VT: %w", err)
	}
	defer postResp.Body.Close()
	
	if postResp.StatusCode == http.StatusTooManyRequests {
		return false, fmt.Errorf("rate limit hit during POST")
	}
	
	if postResp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("VT submission returned status %d", postResp.StatusCode)
	}

	var postResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(postResp.Body).Decode(&postResult); err != nil {
		return false, fmt.Errorf("failed to parse VT submission response: %w", err)
	}
	
	analysisID := postResult.Data.ID
	if analysisID == "" {
		return false, fmt.Errorf("VT submission returned empty analysis ID")
	}

	// Wait 20 seconds to respect rate limit (since we just used 1 request, and limit is 4/min)
	time.Sleep(20 * time.Second)

	pollReq, err := http.NewRequest("GET", fmt.Sprintf("https://www.virustotal.com/api/v3/analyses/%s", analysisID), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create VT poll request: %w", err)
	}
	pollReq.Header.Add("x-apikey", w.apiKey)
	pollReq.Header.Add("accept", "application/json")
	
	pollResp, err := client.Do(pollReq)
	if err != nil {
		return false, fmt.Errorf("failed to poll VT API: %w", err)
	}
	defer pollResp.Body.Close()

	if pollResp.StatusCode == http.StatusTooManyRequests {
		return false, fmt.Errorf("rate limit hit during polling")
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
		return false, fmt.Errorf("failed to parse VT poll response: %w", err)
	}
	
	if pollResult.Data.Attributes.Status == "completed" {
		stats := pollResult.Data.Attributes.Stats
		if stats.Malicious > 0 || stats.Suspicious > 0 {
			return true, nil
		}
		return false, nil
	}

	return false, fmt.Errorf("analysis not completed yet")
}
