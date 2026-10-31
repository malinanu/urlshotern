package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

const micrositeAnalyticsCacheTTL = 5 * time.Minute

// micrositeEventTypes are the event types accepted by RecordEvent.
var micrositeEventTypes = map[string]bool{
	models.MicrositeEventView:          true,
	models.MicrositeEventClick:         true,
	models.MicrositeEventVCardDownload: true,
	models.MicrositeEventCall:          true,
	models.MicrositeEventSMS:           true,
	models.MicrositeEventEmail:         true,
	models.MicrositeEventWhatsApp:      true,
	models.MicrositeEventFormSubmit:    true,
	models.MicrositeEventCalendarAdd:   true,
	models.MicrositeEventVideoPlay:     true,
}

// IsMicrositeEventType reports whether t is a known event type.
func IsMicrositeEventType(t string) bool { return micrositeEventTypes[t] }

// MicrositeAnalyticsService records and reports first-party page telemetry.
// It is independent of the short-link click pipeline.
type MicrositeAnalyticsService struct {
	db         micrositeDB
	cache      micrositeCache
	visitorKey []byte
}

// NewMicrositeAnalyticsService creates the analytics service. secret keys the
// visitor hash so it cannot be reversed by brute-forcing the IPv4 space.
func NewMicrositeAnalyticsService(db *storage.PostgresStorage, redis *storage.RedisStorage, secret string) *MicrositeAnalyticsService {
	sum := sha256.Sum256([]byte("microsite-visitor:" + secret))
	s := &MicrositeAnalyticsService{db: db, visitorKey: sum[:]}
	if redis != nil {
		s.cache = redis
	}
	return s
}

// VisitorHash returns HMAC-SHA256(key, ip|ua|UTC day). It rotates daily, so a
// visitor is "unique" per page per day and raw IPs are never stored.
func (s *MicrositeAnalyticsService) VisitorHash(ip, userAgent string, now time.Time) string {
	mac := hmac.New(sha256.New, s.visitorKey)
	mac.Write([]byte(ip + "|" + userAgent + "|" + now.UTC().Format("2006-01-02")))
	return hex.EncodeToString(mac.Sum(nil))
}

// RecordEvent inserts one telemetry row.
func (s *MicrositeAnalyticsService) RecordEvent(ev models.MicrositeEvent) error {
	if !micrositeEventTypes[ev.EventType] {
		return fmt.Errorf("unknown event type %q", ev.EventType)
	}
	_, err := s.db.Exec(`
		INSERT INTO microsite_events (microsite_id, event_type, block_id, item_id, visitor_hash, is_unique, is_bot,
			country, device_type, os, browser, referrer_domain, referrer_category, utm_source, utm_medium, utm_campaign)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, $7,
			NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, ''),
			NULLIF($14, ''), NULLIF($15, ''), NULLIF($16, ''))`,
		ev.MicrositeID, ev.EventType, clip(ev.BlockID, 40), clip(ev.ItemID, 40), ev.VisitorHash, ev.IsUnique, ev.IsBot,
		normalizeCountry(ev.Country), clip(ev.DeviceType, 20), clip(ev.OS, 50), clip(ev.Browser, 50),
		clip(ev.ReferrerDomain, 255), clip(ev.ReferrerCategory, 30),
		clip(ev.UTMSource, 100), clip(ev.UTMMedium, 100), clip(ev.UTMCampaign, 100))
	return mapMicrositeDBError(err)
}

// GetReport builds the Track-tab report for the last `days` days. labels maps
// "blockID/itemID" to a human label for the per-link breakdown.
func (s *MicrositeAnalyticsService) GetReport(micrositeID int64, days int, labels map[string]string) (*models.MicrositeAnalytics, error) {
	if days < 1 {
		days = 1
	}
	if days > 365 {
		days = 365
	}
	cacheKey := fmt.Sprintf("microsite:analytics:%d:%d", micrositeID, days)
	if s.cache != nil {
		if cached, err := s.cache.Get(cacheKey); err == nil {
			var r models.MicrositeAnalytics
			if json.Unmarshal([]byte(cached), &r) == nil {
				applyBlockLabels(&r, labels)
				return &r, nil
			}
		}
	}

	r := &models.MicrositeAnalytics{Days: days, ContactActions: map[string]int64{}, Blocks: []models.MicrositeBlockClicks{}}
	const window = `microsite_id = $1 AND is_bot = FALSE AND occurred_at >= NOW() - make_interval(days => $2)`

	// Totals
	var calls, sms, emails, whatsapp int64
	err := s.db.QueryRow(`
		SELECT COUNT(*) FILTER (WHERE event_type = 'view'),
		       COUNT(*) FILTER (WHERE event_type = 'view' AND is_unique),
		       COUNT(*) FILTER (WHERE event_type = 'click'),
		       COUNT(*) FILTER (WHERE event_type = 'vcard_download'),
		       COUNT(*) FILTER (WHERE event_type = 'call'),
		       COUNT(*) FILTER (WHERE event_type = 'sms'),
		       COUNT(*) FILTER (WHERE event_type = 'email'),
		       COUNT(*) FILTER (WHERE event_type = 'whatsapp'),
		       COUNT(*) FILTER (WHERE event_type = 'form_submit'),
		       COUNT(*) FILTER (WHERE event_type = 'calendar_add'),
		       COUNT(*) FILTER (WHERE event_type = 'video_play')
		  FROM microsite_events WHERE `+window, micrositeID, days).
		Scan(&r.Views, &r.UniqueVisitors, &r.LinkClicks, &r.VCardDownloads, &calls, &sms, &emails, &whatsapp,
			&r.FormSubmits, &r.CalendarAdds, &r.VideoPlays)
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	r.ContactActions = map[string]int64{"call": calls, "sms": sms, "email": emails, "whatsapp": whatsapp}
	if r.Views > 0 {
		r.CTR = round2(float64(r.LinkClicks+calls+sms+emails+whatsapp) / float64(r.Views) * 100)
	}
	if r.UniqueVisitors > 0 {
		r.ContactSaveRate = round2(float64(r.VCardDownloads) / float64(r.UniqueVisitors) * 100)
	}

	// Daily series (filled with zero days)
	daily := map[string]*models.MicrositeDailyPoint{}
	rows, err := s.db.Query(`
		SELECT to_char(date_trunc('day', occurred_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS d,
		       COUNT(*) FILTER (WHERE event_type = 'view'),
		       COUNT(*) FILTER (WHERE event_type = 'view' AND is_unique),
		       COUNT(*) FILTER (WHERE event_type IN ('click', 'call', 'sms', 'email', 'whatsapp')),
		       COUNT(*) FILTER (WHERE event_type = 'vcard_download')
		  FROM microsite_events WHERE `+window+`
		 GROUP BY d`, micrositeID, days)
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	for rows.Next() {
		var p models.MicrositeDailyPoint
		if err := rows.Scan(&p.Date, &p.Views, &p.UniqueVisitors, &p.Clicks, &p.VCardDownloads); err != nil {
			_ = rows.Close()
			return nil, err
		}
		daily[p.Date] = &p
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	start := time.Now().UTC().AddDate(0, 0, -(days - 1))
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i).Format("2006-01-02")
		if p, ok := daily[d]; ok {
			r.Daily = append(r.Daily, *p)
		} else {
			r.Daily = append(r.Daily, models.MicrositeDailyPoint{Date: d})
		}
	}

	viewsOnly := window + ` AND event_type = 'view'`
	if r.TopReferrers, err = s.countBy(`COALESCE(referrer_domain, 'Direct')`, viewsOnly, micrositeID, days, 10); err != nil {
		return nil, err
	}
	if r.ReferrerSources, err = s.countBy(`COALESCE(referrer_category, 'direct')`, viewsOnly, micrositeID, days, 10); err != nil {
		return nil, err
	}
	if r.Devices, err = s.countBy(`COALESCE(device_type, 'unknown')`, viewsOnly, micrositeID, days, 10); err != nil {
		return nil, err
	}
	if r.OperatingSystems, err = s.countBy(`COALESCE(os, 'Unknown')`, viewsOnly, micrositeID, days, 10); err != nil {
		return nil, err
	}
	if r.Countries, err = s.countBy(`COALESCE(country, '--')`, viewsOnly, micrositeID, days, 10); err != nil {
		return nil, err
	}

	// Per-block / per-item distribution of taps
	rows, err = s.db.Query(`
		SELECT COALESCE(block_id, ''), COALESCE(item_id, ''), event_type, COUNT(*)
		  FROM microsite_events
		 WHERE `+window+` AND event_type <> 'view'
		 GROUP BY 1, 2, 3
		 ORDER BY 4 DESC
		 LIMIT 100`, micrositeID, days)
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var bc models.MicrositeBlockClicks
		if err := rows.Scan(&bc.BlockID, &bc.ItemID, &bc.EventType, &bc.Count); err != nil {
			return nil, err
		}
		r.Blocks = append(r.Blocks, bc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if s.cache != nil {
		if encoded, err := json.Marshal(r); err == nil {
			_ = s.cache.Set(cacheKey, encoded, micrositeAnalyticsCacheTTL)
		}
	}
	applyBlockLabels(r, labels)
	return r, nil
}

// countBy runs a grouped count. expr and where are constant SQL fragments.
func (s *MicrositeAnalyticsService) countBy(expr, where string, id int64, days, limit int) ([]models.MicrositeCount, error) {
	rows, err := s.db.Query(`SELECT `+expr+` AS label, COUNT(*) AS n FROM microsite_events WHERE `+where+
		` GROUP BY label ORDER BY n DESC LIMIT `+fmt.Sprint(limit), id, days) // #nosec G202 -- constant fragments only
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	defer rows.Close()
	out := []models.MicrositeCount{}
	for rows.Next() {
		var c models.MicrositeCount
		if err := rows.Scan(&c.Label, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// MicrositeBlockLabels maps "blockID/itemID" (and "blockID/") to labels taken
// from page content, for the per-link breakdown.
func MicrositeBlockLabels(content models.MicrositeContent) map[string]string {
	labels := map[string]string{}
	for _, b := range content.Blocks {
		data, err := DecodeMicrositeBlockData(b.Type, b.Data, "block", true)
		if err != nil {
			continue
		}
		switch d := data.(type) {
		case models.LinkListData:
			for _, it := range d.Items {
				labels[b.ID+"/"+it.ID] = it.Label
			}
		case models.ContactButtonsData:
			for _, it := range d.Items {
				labels[b.ID+"/"+it.ID] = firstNonEmpty(it.Label, contactLabels[it.Kind])
			}
		case models.SocialIconsData:
			for _, it := range d.Items {
				labels[b.ID+"/"+it.ID] = MicrositeSocialPlatforms[it.Platform]
			}
		case models.VCardActionData:
			labels[b.ID+"/"] = d.ButtonText
		case models.ImageData:
			labels[b.ID+"/"] = firstNonEmpty(d.Alt, "Image")
		}
	}
	return labels
}

func applyBlockLabels(r *models.MicrositeAnalytics, labels map[string]string) {
	for i := range r.Blocks {
		key := r.Blocks[i].BlockID + "/" + r.Blocks[i].ItemID
		if l, ok := labels[key]; ok {
			r.Blocks[i].Label = l
		} else if r.Blocks[i].Label == "" {
			r.Blocks[i].Label = "Removed item"
		}
	}
}

func normalizeCountry(c string) string {
	c = strings.ToUpper(strings.TrimSpace(c))
	if len(c) != 2 || c == "XX" || c == "T1" {
		return ""
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return ""
		}
	}
	return c
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	// trim to n bytes on a rune boundary
	for n > 0 && n < len(s) && (s[n]&0xC0) == 0x80 {
		n--
	}
	return s[:n]
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}
