package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/lib/pq"
)

// CampaignService manages promotional campaigns.
type CampaignService struct {
	db            *storage.PostgresStorage
	config        *configs.Config
	emailService  *EmailService
	smsService    *SMSService
	httpClient    *http.Client
	openRouterKey string
}

// NewCampaignService creates a new CampaignService.
func NewCampaignService(db *storage.PostgresStorage, config *configs.Config, emailService *EmailService, smsService *SMSService) *CampaignService {
	return &CampaignService{
		db:            db,
		config:        config,
		emailService:  emailService,
		smsService:    smsService,
		httpClient:    &http.Client{Timeout: 60 * time.Second},
		openRouterKey: config.OpenRouterAPIKey,
	}
}

// ─── Campaign CRUD ─────────────────────────────────────────────────────────────

// CreateCampaign inserts a new campaign row.
func (s *CampaignService) CreateCampaign(req *models.CreateCampaignRequest, createdBy *int64) (*models.Campaign, error) {
	c := &models.Campaign{
		Name:        req.Name,
		Description: req.Description,
		AIContext:   req.AIContext,
		Channels:    req.Channels,
		Status:      "draft",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO campaigns (name, description, ai_context, channels, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	err := s.db.QueryRow(query,
		c.Name, c.Description, c.AIContext, pq.Array(c.Channels),
		c.Status, c.CreatedBy, c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID)
	if err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}
	return c, nil
}

// GetCampaign returns a single campaign by ID.
func (s *CampaignService) GetCampaign(id int64) (*models.Campaign, error) {
	c := &models.Campaign{}
	query := `SELECT id, name, description, ai_context, channels, status, created_by, created_at, updated_at
	          FROM campaigns WHERE id = $1`
	var channels pq.StringArray
	err := s.db.QueryRow(query, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.AIContext, &channels,
		&c.Status, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	c.Channels = []string(channels)
	return c, nil
}

// ListCampaigns returns paginated campaigns.
func (s *CampaignService) ListCampaigns(page, limit int) ([]models.Campaign, int64, error) {
	var total int64
	if err := s.db.QueryRow("SELECT COUNT(*) FROM campaigns").Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	query := `SELECT id, name, description, ai_context, channels, status, created_by, created_at, updated_at
	          FROM campaigns ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		var channels pq.StringArray
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.AIContext, &channels,
			&c.Status, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		c.Channels = []string(channels)
		campaigns = append(campaigns, c)
	}
	return campaigns, total, nil
}

// UpdateCampaign updates mutable campaign fields.
func (s *CampaignService) UpdateCampaign(id int64, req *models.UpdateCampaignRequest) (*models.Campaign, error) {
	c, err := s.GetCampaign(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Description != nil {
		c.Description = *req.Description
	}
	if req.AIContext != nil {
		c.AIContext = *req.AIContext
	}
	if len(req.Channels) > 0 {
		c.Channels = req.Channels
	}
	c.UpdatedAt = time.Now()

	_, err = s.db.Exec(
		`UPDATE campaigns SET name=$1, description=$2, ai_context=$3, channels=$4, updated_at=$5 WHERE id=$6`,
		c.Name, c.Description, c.AIContext, pq.Array(c.Channels), c.UpdatedAt, c.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}
	return c, nil
}

// DeleteCampaign removes a campaign (and cascades to sends/links).
func (s *CampaignService) DeleteCampaign(id int64) error {
	_, err := s.db.Exec("DELETE FROM campaigns WHERE id=$1", id)
	return err
}

// ─── Contact Import ────────────────────────────────────────────────────────────

// ImportContactsJSON parses a JSON array and links contacts to a campaign.
func (s *CampaignService) ImportContactsJSON(campaignID int64, rows []models.ContactImportRow) (int, error) {
	return s.importContacts(campaignID, rows)
}

// ImportContactsCSV parses an io.Reader CSV and links contacts to a campaign.
func (s *CampaignService) ImportContactsCSV(campaignID int64, r io.Reader) (int, error) {
	csvReader := csv.NewReader(r)
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("parse csv: %w", err)
	}

	// Detect header row (name, email, phone in any order)
	if len(records) == 0 {
		return 0, nil
	}
	header := make(map[string]int)
	for i, h := range records[0] {
		header[strings.ToLower(strings.TrimSpace(h))] = i
	}
	nameIdx, hasName := header["name"]
	emailIdx, hasEmail := header["email"]
	phoneIdx, hasPhone := header["phone"]

	var rows []models.ContactImportRow
	for _, rec := range records[1:] {
		row := models.ContactImportRow{}
		if hasName && nameIdx < len(rec) {
			row.Name = strings.TrimSpace(rec[nameIdx])
		}
		if hasEmail && emailIdx < len(rec) {
			row.Email = strings.TrimSpace(rec[emailIdx])
		}
		if hasPhone && phoneIdx < len(rec) {
			row.Phone = strings.TrimSpace(rec[phoneIdx])
		}
		rows = append(rows, row)
	}
	return s.importContacts(campaignID, rows)
}

func (s *CampaignService) importContacts(campaignID int64, rows []models.ContactImportRow) (int, error) {
	count := 0
	for _, row := range rows {
		if row.Email == "" && row.Phone == "" {
			continue
		}
		var emailPtr, phonePtr *string
		if row.Email != "" {
			emailPtr = &row.Email
		}
		if row.Phone != "" {
			phonePtr = &row.Phone
		}

		// Upsert contact. Email-only/email+phone rows dedupe on uq_campaign_contacts_email;
		// phone-only rows (no email) dedupe on the partial index uq_campaign_contacts_phone_only,
		// since SQL UNIQUE constraints never treat two NULLs as conflicting.
		var contactID int64
		var err error
		if emailPtr != nil {
			err = s.db.QueryRow(`
				INSERT INTO campaign_contacts (name, email, phone, created_at)
				VALUES ($1, $2, $3, NOW())
				ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, phone = COALESCE(EXCLUDED.phone, campaign_contacts.phone)
				RETURNING id`,
				row.Name, emailPtr, phonePtr,
			).Scan(&contactID)
		} else {
			err = s.db.QueryRow(`
				INSERT INTO campaign_contacts (name, email, phone, created_at)
				VALUES ($1, $2, $3, NOW())
				ON CONFLICT (phone) WHERE email IS NULL DO UPDATE SET name = EXCLUDED.name
				RETURNING id`,
				row.Name, emailPtr, phonePtr,
			).Scan(&contactID)
		}
		if err != nil {
			log.Printf("[Campaign] upsert contact error: %v", err)
			continue
		}

		// Link to campaign
		_, err = s.db.Exec(`
			INSERT INTO campaign_contact_links (campaign_id, contact_id)
			VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			campaignID, contactID,
		)
		if err != nil {
			log.Printf("[Campaign] link contact error: %v", err)
			continue
		}
		count++
	}
	return count, nil
}

// ─── Stats ─────────────────────────────────────────────────────────────────────

// GetCampaignStats returns aggregate delivery stats for a campaign.
func (s *CampaignService) GetCampaignStats(campaignID int64) (*models.CampaignStats, error) {
	stats := &models.CampaignStats{CampaignID: campaignID}
	err := s.db.QueryRow(`
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'sent') AS sent,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed,
			COUNT(*) FILTER (WHERE joined_at IS NOT NULL) AS joined
		FROM campaign_sends
		WHERE campaign_id = $1`, campaignID).
		Scan(&stats.Total, &stats.Sent, &stats.Failed, &stats.Joined)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// ─── Join Tracking ─────────────────────────────────────────────────────────────

// MarkContactJoined is called after successful registration.
// It marks any pending campaign_sends for the given email/phone as joined.
func (s *CampaignService) MarkContactJoined(email, phone string) {
	var contactID int64
	var err error
	if email != "" {
		err = s.db.QueryRow("SELECT id FROM campaign_contacts WHERE email = $1", email).Scan(&contactID)
	} else if phone != "" {
		err = s.db.QueryRow("SELECT id FROM campaign_contacts WHERE phone = $1", phone).Scan(&contactID)
	} else {
		return
	}
	if err != nil {
		return // Not a campaign contact — silent
	}
	_, _ = s.db.Exec(
		"UPDATE campaign_sends SET joined_at = NOW() WHERE contact_id = $1 AND joined_at IS NULL",
		contactID,
	)
}

// ─── Send Campaign ─────────────────────────────────────────────────────────────

// SendCampaign kicks off a goroutine that sends messages to all contacts.
// It returns immediately (async fire-and-forget with logging).
func (s *CampaignService) SendCampaign(campaignID int64) error {
	campaign, err := s.GetCampaign(campaignID)
	if err != nil {
		return fmt.Errorf("campaign not found: %w", err)
	}
	if campaign.Status == "sending" {
		return fmt.Errorf("campaign is already sending")
	}

	// Mark campaign as sending
	_, _ = s.db.Exec("UPDATE campaigns SET status='sending', updated_at=NOW() WHERE id=$1", campaignID)

	// Fetch contacts linked to this campaign
	contacts, err := s.fetchCampaignContacts(campaignID)
	if err != nil {
		return fmt.Errorf("fetch contacts: %w", err)
	}

	go s.runSend(campaign, contacts)
	return nil
}

func (s *CampaignService) fetchCampaignContacts(campaignID int64) ([]models.CampaignContact, error) {
	rows, err := s.db.Query(`
		SELECT cc.id, cc.name, cc.email, cc.phone, cc.created_at
		FROM campaign_contacts cc
		JOIN campaign_contact_links l ON l.contact_id = cc.id
		WHERE l.campaign_id = $1`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.CampaignContact
	for rows.Next() {
		var c models.CampaignContact
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.CreatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}

func (s *CampaignService) runSend(campaign *models.Campaign, contacts []models.CampaignContact) {
	log.Printf("[Campaign] Starting send for campaign %d (%s), %d contacts", campaign.ID, campaign.Name, len(contacts))

	for _, contact := range contacts {
		for _, channel := range campaign.Channels {
			s.sendToContact(campaign, &contact, channel)
		}
	}

	_, _ = s.db.Exec("UPDATE campaigns SET status='done', updated_at=NOW() WHERE id=$1", campaign.ID)
	log.Printf("[Campaign] Completed send for campaign %d", campaign.ID)
}

func (s *CampaignService) sendToContact(campaign *models.Campaign, contact *models.CampaignContact, channel string) {
	// Insert a pending send record (ignore if already exists)
	var sendID int64
	err := s.db.QueryRow(`
		INSERT INTO campaign_sends (campaign_id, contact_id, channel, status, created_at)
		VALUES ($1, $2, $3, 'pending', NOW())
		ON CONFLICT (campaign_id, contact_id, channel) DO UPDATE SET status='pending'
		RETURNING id`, campaign.ID, contact.ID, channel).Scan(&sendID)
	if err != nil {
		log.Printf("[Campaign] insert send row error: %v", err)
		return
	}

	var sendErr error
	switch channel {
	case "email":
		sendErr = s.sendEmail(campaign, contact)
	case "sms":
		sendErr = s.sendSMSMsg(campaign, contact)
	case "whatsapp":
		sendErr = s.sendWhatsApp(campaign, contact)
	default:
		sendErr = fmt.Errorf("unknown channel: %s", channel)
	}

	now := time.Now()
	if errors.Is(sendErr, ErrEmailSuppressed) {
		// CAN-SPAM: the contact opted out of marketing email, so nothing was sent
		_, _ = s.db.Exec(
			"UPDATE campaign_sends SET status='skipped', error=$1 WHERE id=$2",
			"recipient unsubscribed from marketing emails", sendID)
	} else if sendErr != nil {
		errMsg := sendErr.Error()
		_, _ = s.db.Exec(
			"UPDATE campaign_sends SET status='failed', error=$1, sent_at=$2 WHERE id=$3",
			errMsg, now, sendID)
		log.Printf("[Campaign] %s send failed for contact %d: %v", channel, contact.ID, sendErr)
	} else {
		_, _ = s.db.Exec(
			"UPDATE campaign_sends SET status='sent', sent_at=$1 WHERE id=$2",
			now, sendID)
	}
}

// ─── Email (AI-generated) ─────────────────────────────────────────────────────

func (s *CampaignService) sendEmail(campaign *models.Campaign, contact *models.CampaignContact) error {
	if contact.Email == nil || *contact.Email == "" {
		return fmt.Errorf("no email address for contact %d", contact.ID)
	}

	// Honour unsubscribes before spending an AI call on the message
	suppressed, err := s.emailService.IsEmailSuppressed(*contact.Email)
	if err != nil {
		return fmt.Errorf("check suppression list: %w", err)
	}
	if suppressed {
		return ErrEmailSuppressed
	}

	subject, htmlBody, err := s.generateEmailContent(campaign, contact)
	if err != nil {
		return fmt.Errorf("AI generate email: %w", err)
	}

	// Plain text fallback: strip HTML tags naively
	textBody := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "</p>", "\n\n", "<li>", "• ").Replace(htmlBody)
	textBody = stripHTMLTagsSimple(textBody)

	// Adds the promotional footer: postal address, privacy policy, unsubscribe link
	return s.emailService.SendPromotionalEmail(*contact.Email, subject, htmlBody, textBody)
}

func (s *CampaignService) generateEmailContent(campaign *models.Campaign, contact *models.CampaignContact) (subject, htmlBody string, err error) {
	contactName := contact.Name
	if contactName == "" {
		contactName = "there"
	}

	prompt := fmt.Sprintf(`You are writing a promotional email for trunc.site (a URL shortener & link analytics platform).

Campaign context: %s

Write a warm, professional promotional email to %s.
The subject line must accurately describe what the email is about (no misleading or "Re:"/"Fwd:" subjects).
Do not add a footer, postal address, unsubscribe link or sign-off disclaimer; those are appended automatically.
Return ONLY valid JSON (no markdown):
{
  "subject": "email subject here",
  "html_body": "full HTML email body here (use <p>, <ul>, <li>, <strong> tags, inline CSS for bold/color)"
}`, campaign.AIContext, contactName)

	raw, err := s.callAI(prompt)
	if err != nil {
		return "", "", err
	}

	// Parse JSON response
	raw = stripCodeFences(raw)
	var result struct {
		Subject  string `json:"subject"`
		HTMLBody string `json:"html_body"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Fallback: use raw as body
		return fmt.Sprintf("Exclusive offer from trunc.site for %s", contactName), raw, nil
	}
	return result.Subject, result.HTMLBody, nil
}

// ─── SMS ──────────────────────────────────────────────────────────────────────

func (s *CampaignService) sendSMSMsg(campaign *models.Campaign, contact *models.CampaignContact) error {
	if contact.Phone == nil || *contact.Phone == "" {
		return fmt.Errorf("no phone for contact %d", contact.ID)
	}

	// Short plain-text SMS — use AI for a one-liner, fall back to canned text
	text, err := s.generateSMSContent(campaign, contact)
	if err != nil {
		text = fmt.Sprintf("Hi %s! Check out trunc.site for powerful link shortening. Sign up free: https://trunc.site", contact.Name)
	}

	return s.smsService.SendRawSMS(*contact.Phone, text)
}

func (s *CampaignService) generateSMSContent(campaign *models.Campaign, contact *models.CampaignContact) (string, error) {
	contactName := contact.Name
	if contactName == "" {
		contactName = "there"
	}
	prompt := fmt.Sprintf(`Write a single SMS (max 160 chars) promoting trunc.site URL shortener to %s. Context: %s. Return ONLY the SMS text, no quotes.`,
		contactName, campaign.AIContext)
	return s.callAI(prompt)
}

// ─── WhatsApp ─────────────────────────────────────────────────────────────────

func (s *CampaignService) sendWhatsApp(campaign *models.Campaign, contact *models.CampaignContact) error {
	if contact.Phone == nil || *contact.Phone == "" {
		return fmt.Errorf("no phone for contact %d (whatsapp)", contact.ID)
	}
	if s.config.TwilioAccountSID == "" || s.config.TwilioAuthToken == "" {
		return fmt.Errorf("Twilio credentials not configured")
	}
	fromWA := s.config.TwilioWhatsAppFrom
	if fromWA == "" {
		return fmt.Errorf("TWILIO_WHATSAPP_FROM not configured")
	}

	text, err := s.generateSMSContent(campaign, contact)
	if err != nil {
		text = fmt.Sprintf("Hi %s! Check out trunc.site. Sign up free: https://trunc.site", contact.Name)
	}

	return s.sendViaTwilioWhatsApp(*contact.Phone, text, fromWA)
}

func (s *CampaignService) sendViaTwilioWhatsApp(toPhone, message, fromWA string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.config.TwilioAccountSID)

	form := url.Values{}
	form.Set("To", "whatsapp:"+toPhone)
	form.Set("From", fromWA)
	form.Set("Body", message)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build twilio whatsapp request: %w", err)
	}
	req.SetBasicAuth(s.config.TwilioAccountSID, s.config.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("twilio whatsapp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio whatsapp error (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// ─── OpenRouter AI helper ─────────────────────────────────────────────────────

func (s *CampaignService) callAI(prompt string) (string, error) {
	if s.openRouterKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not configured")
	}

	type chatMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type orRequest struct {
		Model     string    `json:"model"`
		Messages  []chatMsg `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
	}
	type orResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	payload := orRequest{
		Model:     openRouterModel, // reuse const from ai_insights.go
		MaxTokens: 1024,
		Messages:  []chatMsg{{Role: "user", Content: prompt}},
	}
	body, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.openRouterKey)
	req.Header.Set("HTTP-Referer", "https://trunc.site")
	req.Header.Set("X-Title", "trunc.site Campaign")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result orResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse AI response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("openrouter: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// stripHTMLTagsSimple removes HTML angle-bracket tags from a string.
func stripHTMLTagsSimple(s string) string {
	var buf strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
