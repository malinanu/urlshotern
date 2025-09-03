package services

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// AdvancedAnalyticsService handles enhanced analytics operations
type AdvancedAnalyticsService struct {
	storage           *storage.PostgresStorage
	geoIPService      *FreeGeoIPService
	userAgentService  *UserAgentService
	referrerService   *ReferrerParsingService
}

// NewAdvancedAnalyticsService creates a new advanced analytics service
func NewAdvancedAnalyticsService(storage *storage.PostgresStorage) *AdvancedAnalyticsService {
	return &AdvancedAnalyticsService{
		storage:          storage,
		geoIPService:     NewFreeGeoIPService(),
		userAgentService: NewUserAgentService(),
		referrerService:  NewReferrerParsingService(),
	}
}

// ProcessEnhancedClickEvent processes a click event and updates all analytics tables
func (a *AdvancedAnalyticsService) ProcessEnhancedClickEvent(clickEvent *models.ClickEvent) error {
	// Parse user agent for device information using our free service
	deviceInfo := a.userAgentService.ParseUserAgent(clickEvent.UserAgent)
	
	// Get geographic information using our free GeoIP service
	geoInfo, err := a.geoIPService.LookupIP(clickEvent.IPAddress)
	if err != nil {
		// Use fallback if GeoIP lookup fails
		geoInfo = &GeoIPResult{
			CountryCode: clickEvent.CountryCode,
			CountryName: a.geoIPService.getCountryName(clickEvent.CountryCode),
		}
	}
	
	// Parse referrer for detailed analytics using our enhanced referrer service
	referrerData, err := a.referrerService.ProcessReferrerData(
		clickEvent.ShortCode, 
		clickEvent.Referrer, 
		clickEvent.IPAddress, 
		clickEvent.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("failed to process referrer data: %w", err)
	}
	
	// Extract time-based information
	timeInfo := a.extractTimeInfo(clickEvent.ClickedAt)
	
	// Begin transaction to ensure all analytics are updated atomically
	tx, err := a.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Update geographic analytics
	if err := a.updateGeographicAnalytics(tx, clickEvent.ShortCode, geoInfo); err != nil {
		return fmt.Errorf("failed to update geographic analytics: %w", err)
	}
	
	// Update device analytics
	if err := a.updateDeviceAnalytics(tx, clickEvent.ShortCode, deviceInfo, clickEvent.IPAddress); err != nil {
		return fmt.Errorf("failed to update device analytics: %w", err)
	}
	
	// Update time analytics
	if err := a.updateTimeAnalytics(tx, clickEvent.ShortCode, timeInfo, clickEvent.IPAddress); err != nil {
		return fmt.Errorf("failed to update time analytics: %w", err)
	}
	
	// Update referrer analytics
	if err := a.updateReferrerAnalytics(tx, clickEvent.ShortCode, referrerData); err != nil {
		return fmt.Errorf("failed to update referrer analytics: %w", err)
	}
	
	return tx.Commit()
}

// Note: DeviceInfo and GeographicInfo are now handled by dedicated services
// DeviceInfo is provided by UserAgentService
// GeographicInfo is provided by FreeGeoIPService

// ReferrerInfo represents parsed referrer information
type ReferrerInfo struct {
	Domain          string
	URL             string
	Type            string
	CampaignSource  string
	CampaignMedium  string
	CampaignName    string
	CampaignTerm    string
	CampaignContent string
}

// TimeInfo represents time-based information
type TimeInfo struct {
	Date        time.Time
	HourOfDay   int
	DayOfWeek   int
	DayOfMonth  int
	Month       int
	Year        int
}

// Note: parseUserAgent is now handled by UserAgentService

// Note: getGeographicInfo is now handled by FreeGeoIPService

// parseReferrer extracts referrer information including UTM parameters
func (a *AdvancedAnalyticsService) parseReferrer(referrerURL string) ReferrerInfo {
	info := ReferrerInfo{
		URL: referrerURL,
	}
	
	if referrerURL == "" {
		info.Type = "direct"
		return info
	}
	
	parsedURL, err := url.Parse(referrerURL)
	if err != nil {
		info.Type = "unknown"
		return info
	}
	
	info.Domain = parsedURL.Host
	
	// Extract UTM parameters
	query := parsedURL.Query()
	info.CampaignSource = query.Get("utm_source")
	info.CampaignMedium = query.Get("utm_medium")
	info.CampaignName = query.Get("utm_campaign")
	info.CampaignTerm = query.Get("utm_term")
	info.CampaignContent = query.Get("utm_content")
	
	// Determine referrer type
	domain := strings.ToLower(info.Domain)
	if strings.Contains(domain, "google") || strings.Contains(domain, "bing") || strings.Contains(domain, "yahoo") || strings.Contains(domain, "duckduckgo") {
		info.Type = "search"
	} else if strings.Contains(domain, "facebook") || strings.Contains(domain, "twitter") || strings.Contains(domain, "instagram") || strings.Contains(domain, "linkedin") || strings.Contains(domain, "pinterest") || strings.Contains(domain, "tiktok") {
		info.Type = "social"
	} else if info.CampaignMedium == "email" {
		info.Type = "email"
	} else if info.CampaignMedium != "" {
		info.Type = "campaign"
	} else {
		info.Type = "referral"
	}
	
	return info
}

// extractTimeInfo extracts time-based information from a timestamp
func (a *AdvancedAnalyticsService) extractTimeInfo(clickTime time.Time) TimeInfo {
	return TimeInfo{
		Date:       time.Date(clickTime.Year(), clickTime.Month(), clickTime.Day(), 0, 0, 0, 0, clickTime.Location()),
		HourOfDay:  clickTime.Hour(),
		DayOfWeek:  int(clickTime.Weekday()),
		DayOfMonth: clickTime.Day(),
		Month:      int(clickTime.Month()),
		Year:       clickTime.Year(),
	}
}

// updateGeographicAnalytics updates or inserts geographic analytics data
func (a *AdvancedAnalyticsService) updateGeographicAnalytics(tx *sql.Tx, shortCode string, geoInfo *GeoIPResult) error {
	query := `
		INSERT INTO geographic_analytics (
			short_code, country_code, country_name, region, city, latitude, longitude, clicks, unique_ips, last_click
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 1, 1, NOW())
		ON CONFLICT (short_code, country_code, region, city) 
		DO UPDATE SET 
			clicks = geographic_analytics.clicks + 1,
			unique_ips = geographic_analytics.unique_ips + 1,
			last_click = NOW(),
			updated_at = NOW()
	`
	
	_, err := tx.Exec(query, shortCode, geoInfo.CountryCode, geoInfo.CountryName, 
		geoInfo.Region, geoInfo.City, geoInfo.Latitude, geoInfo.Longitude)
	return err
}

// updateDeviceAnalytics updates or inserts device analytics data
func (a *AdvancedAnalyticsService) updateDeviceAnalytics(tx *sql.Tx, shortCode string, deviceInfo *DeviceInfo, ipAddress string) error {
	query := `
		INSERT INTO device_analytics (
			short_code, device_type, device_brand, device_model, os_name, os_version, 
			browser_name, browser_version, screen_resolution, user_agent_hash, clicks, unique_ips, last_click
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1, 1, NOW())
		ON CONFLICT (short_code, user_agent_hash) 
		DO UPDATE SET 
			clicks = device_analytics.clicks + 1,
			unique_ips = device_analytics.unique_ips + 1,
			last_click = NOW(),
			updated_at = NOW()
	`
	
	_, err := tx.Exec(query, shortCode, deviceInfo.DeviceType, deviceInfo.DeviceBrand,
		deviceInfo.DeviceModel, deviceInfo.OSName, deviceInfo.OSVersion,
		deviceInfo.BrowserName, deviceInfo.BrowserVersion, deviceInfo.ScreenResolution,
		deviceInfo.UserAgentHash)
	return err
}

// updateTimeAnalytics updates or inserts time-based analytics data
func (a *AdvancedAnalyticsService) updateTimeAnalytics(tx *sql.Tx, shortCode string, timeInfo TimeInfo, ipAddress string) error {
	query := `
		INSERT INTO time_analytics (
			short_code, click_date, hour_of_day, day_of_week, day_of_month, month, year, clicks, unique_ips
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 1, 1)
		ON CONFLICT (short_code, click_date, hour_of_day) 
		DO UPDATE SET 
			clicks = time_analytics.clicks + 1,
			unique_ips = time_analytics.unique_ips + 1,
			updated_at = NOW()
	`
	
	_, err := tx.Exec(query, shortCode, timeInfo.Date, timeInfo.HourOfDay,
		timeInfo.DayOfWeek, timeInfo.DayOfMonth, timeInfo.Month, timeInfo.Year)
	return err
}

// updateReferrerAnalytics updates or inserts referrer analytics data
func (a *AdvancedAnalyticsService) updateReferrerAnalytics(tx *sql.Tx, shortCode string, referrerData *EnhancedReferrerData) error {
	parsed := referrerData.ParsedReferrer
	
	// Extract UTM parameters if they exist
	utmSource, utmMedium, utmCampaign, utmTerm, utmContent := "", "", "", "", ""
	if parsed.UTMParams != nil {
		utmSource = parsed.UTMParams.Source
		utmMedium = parsed.UTMParams.Medium
		utmCampaign = parsed.UTMParams.Campaign
		utmTerm = parsed.UTMParams.Term
		utmContent = parsed.UTMParams.Content
	}
	
	query := `
		INSERT INTO referrer_analytics (
			short_code, referrer_domain, referrer_url, referrer_type, 
			campaign_source, campaign_medium, campaign_name, campaign_term, campaign_content, 
			clicks, unique_clicks, last_click
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 1, 1, NOW())
		ON CONFLICT (short_code, referrer_domain, campaign_source, campaign_medium) 
		DO UPDATE SET 
			clicks = referrer_analytics.clicks + 1,
			unique_clicks = referrer_analytics.unique_clicks + 1,
			last_click = NOW(),
			updated_at = NOW()
	`
	
	_, err := tx.Exec(query, shortCode, parsed.Domain, parsed.OriginalReferrer,
		parsed.Category, utmSource, utmMedium, utmCampaign, utmTerm, utmContent)
	return err
}

// GetAdvancedAnalytics retrieves comprehensive analytics for a short code
func (a *AdvancedAnalyticsService) GetAdvancedAnalytics(shortCode string, days int) (*models.AdvancedAnalyticsResponse, error) {
	// Get basic URL info first
	mapping, err := a.storage.GetURLMappingByShortCode(shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL mapping: %w", err)
	}
	
	response := &models.AdvancedAnalyticsResponse{
		ShortCode:   mapping.ShortCode,
		OriginalURL: mapping.OriginalURL,
		TotalClicks: mapping.ClickCount,
		CreatedAt:   mapping.CreatedAt,
		LastUpdated: time.Now(),
	}
	
	// Get geographic analytics
	geographic, err := a.GetGeographicAnalytics(shortCode, days)
	if err == nil {
		response.Geographic = *geographic
	}
	
	// Get time analytics
	timeData, err := a.GetTimeAnalytics(shortCode, days)
	if err == nil {
		response.TimeData = *timeData
	}
	
	// Get device analytics
	deviceData, err := a.GetDeviceAnalytics(shortCode, days)
	if err == nil {
		response.DeviceData = *deviceData
	}
	
	// Get referrer analytics
	referrers, err := a.GetReferrerAnalytics(shortCode, days)
	if err == nil {
		response.Referrers = referrers
	}
	
	return response, nil
}

// GetEnhancedReferrerAnalytics retrieves detailed referrer analytics with UTM parameters
func (a *AdvancedAnalyticsService) GetEnhancedReferrerAnalytics(shortCode string, days int) (*ReferrerAnalytics, error) {
	query := `
		SELECT 
			referrer_domain,
			referrer_url,
			referrer_type,
			campaign_source,
			campaign_medium,
			campaign_name,
			campaign_term,
			campaign_content,
			clicks,
			unique_clicks,
			last_click
		FROM referrer_analytics 
		WHERE short_code = $1 
		  AND created_at >= NOW() - INTERVAL '%d days'
		ORDER BY clicks DESC
		LIMIT 100
	`
	
	rows, err := a.storage.Query(fmt.Sprintf(query, days), shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to query referrer analytics: %w", err)
	}
	defer rows.Close()
	
	var referrers []models.ReferrerStat
	var totalClicks int64
	categoryMap := make(map[string]int64)
	campaignMap := make(map[string]*CampaignStat)
	domainMap := make(map[string]*DomainStat)
	utmSources := make(map[string]int64)
	utmMediums := make(map[string]int64)
	utmCampaigns := make(map[string]int64)
	organicClicks := int64(0)
	
	for rows.Next() {
		var domain, url, refType, source, medium, campaign, term, content sql.NullString
		var clicks, uniqueClicks int64
		var lastClick time.Time
		
		err := rows.Scan(&domain, &url, &refType, &source, &medium, &campaign, &term, &content, &clicks, &uniqueClicks, &lastClick)
		if err != nil {
			return nil, fmt.Errorf("failed to scan referrer row: %w", err)
		}
		
		referrer := models.ReferrerStat{
			Referrer: url.String,
			Clicks:   clicks,
		}
		referrers = append(referrers, referrer)
		totalClicks += clicks
		
		// Categorize referrers
		category := refType.String
		if category == "" {
			category = "other"
		}
		categoryMap[category] += clicks
		
		// Track campaigns
		if source.Valid && source.String != "" {
			key := fmt.Sprintf("%s|%s|%s", campaign.String, source.String, medium.String)
			if _, exists := campaignMap[key]; !exists {
				campaignMap[key] = &CampaignStat{
					Campaign: campaign.String,
					Source:   source.String,
					Medium:   medium.String,
				}
			}
			campaignMap[key].Clicks += clicks
			
			// UTM breakdown
			utmSources[source.String] += clicks
			if medium.Valid {
				utmMediums[medium.String] += clicks
			}
			if campaign.Valid {
				utmCampaigns[campaign.String] += clicks
			}
		} else {
			organicClicks += clicks
		}
		
		// Track domains
		if domain.Valid {
			parsed := a.referrerService.ParseReferrer(url.String)
			domainKey := domain.String
			if _, exists := domainMap[domainKey]; !exists {
				domainMap[domainKey] = &DomainStat{
					Domain:   domain.String,
					Platform: parsed.Platform,
				}
			}
			domainMap[domainKey].Clicks += clicks
		}
	}
	
	// Convert maps to slices and calculate percentages
	categories := make([]ReferrerCategoryStat, 0, len(categoryMap))
	for category, clicks := range categoryMap {
		percentage := float64(clicks) / float64(totalClicks) * 100
		categories = append(categories, ReferrerCategoryStat{
			Category:   category,
			Clicks:     clicks,
			Percentage: percentage,
		})
	}
	
	campaigns := make([]CampaignStat, 0, len(campaignMap))
	for _, campaign := range campaignMap {
		campaign.Percentage = float64(campaign.Clicks) / float64(totalClicks) * 100
		campaigns = append(campaigns, *campaign)
	}
	
	domains := make([]DomainStat, 0, len(domainMap))
	for _, domain := range domainMap {
		domain.Percentage = float64(domain.Clicks) / float64(totalClicks) * 100
		domains = append(domains, *domain)
	}
	
	// Create UTM breakdown
	sources := make([]UTMParamStat, 0, len(utmSources))
	for source, clicks := range utmSources {
		sources = append(sources, UTMParamStat{
			Value:      source,
			Clicks:     clicks,
			Percentage: float64(clicks) / float64(totalClicks) * 100,
		})
	}
	
	mediums := make([]UTMParamStat, 0, len(utmMediums))
	for medium, clicks := range utmMediums {
		mediums = append(mediums, UTMParamStat{
			Value:      medium,
			Clicks:     clicks,
			Percentage: float64(clicks) / float64(totalClicks) * 100,
		})
	}
	
	campaignsList := make([]UTMParamStat, 0, len(utmCampaigns))
	for campaign, clicks := range utmCampaigns {
		campaignsList = append(campaignsList, UTMParamStat{
			Value:      campaign,
			Clicks:     clicks,
			Percentage: float64(clicks) / float64(totalClicks) * 100,
		})
	}
	
	paidClicks := totalClicks - organicClicks
	
	return &ReferrerAnalytics{
		ShortCode:   shortCode,
		TotalClicks: totalClicks,
		Referrers:   referrers,
		Categories:  categories,
		Campaigns:   campaigns,
		TopDomains:  domains,
		UTMBreakdown: UTMBreakdown{
			Sources:   sources,
			Mediums:   mediums,
			Campaigns: campaignsList,
		},
		OrganicVsPaid: OrganicVsPaidStat{
			OrganicClicks:  organicClicks,
			PaidClicks:     paidClicks,
			OrganicPercent: float64(organicClicks) / float64(totalClicks) * 100,
			PaidPercent:    float64(paidClicks) / float64(totalClicks) * 100,
		},
	}, nil
}

// GetUTMCampaignAnalytics retrieves UTM campaign performance analytics
func (a *AdvancedAnalyticsService) GetUTMCampaignAnalytics(shortCode string, days int) ([]CampaignStat, error) {
	query := `
		SELECT 
			campaign_name,
			campaign_source,
			campaign_medium,
			SUM(clicks) as total_clicks
		FROM referrer_analytics 
		WHERE short_code = $1 
		  AND created_at >= NOW() - INTERVAL '%d days'
		  AND campaign_source IS NOT NULL 
		  AND campaign_source != ''
		GROUP BY campaign_name, campaign_source, campaign_medium
		ORDER BY total_clicks DESC
		LIMIT 50
	`
	
	rows, err := a.storage.Query(fmt.Sprintf(query, days), shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to query UTM campaigns: %w", err)
	}
	defer rows.Close()
	
	var campaigns []CampaignStat
	totalClicks := int64(0)
	
	// First pass to calculate total
	for rows.Next() {
		var campaign, source, medium sql.NullString
		var clicks int64
		
		err := rows.Scan(&campaign, &source, &medium, &clicks)
		if err != nil {
			return nil, fmt.Errorf("failed to scan UTM campaign row: %w", err)
		}
		
		campaigns = append(campaigns, CampaignStat{
			Campaign: campaign.String,
			Source:   source.String,
			Medium:   medium.String,
			Clicks:   clicks,
		})
		totalClicks += clicks
	}
	
	// Second pass to calculate percentages
	for i := range campaigns {
		campaigns[i].Percentage = float64(campaigns[i].Clicks) / float64(totalClicks) * 100
	}
	
	return campaigns, nil
}

// GetReferrerInsights retrieves intelligent insights about referrer patterns
func (a *AdvancedAnalyticsService) GetReferrerInsights(shortCode string, days int) (map[string]interface{}, error) {
	// Get basic referrer stats first
	query := `
		SELECT referrer, COUNT(*) as count
		FROM clicks 
		WHERE short_code = $1 
		AND clicked_at > NOW() - INTERVAL '%d days'
		AND referrer IS NOT NULL AND referrer != ''
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT 50
	`
	
	rows, err := a.storage.Query(fmt.Sprintf(query, days), shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get referrer stats: %w", err)
	}
	defer rows.Close()

	var referrers []models.ReferrerStat
	for rows.Next() {
		var referrer string
		var count int64
		if err := rows.Scan(&referrer, &count); err != nil {
			continue
		}
		referrers = append(referrers, models.ReferrerStat{
			Referrer: referrer,
			Clicks:   count,
		})
	}
	
	// Use our referrer service to generate insights
	insights := a.referrerService.GetReferrerInsights(referrers)
	
	// Add additional insights specific to this URL
	additionalQuery := `
		SELECT 
			COUNT(DISTINCT referrer_domain) as unique_domains,
			COUNT(*) as total_referrer_records,
			SUM(CASE WHEN campaign_source IS NOT NULL AND campaign_source != '' THEN clicks ELSE 0 END) as utm_clicks,
			SUM(clicks) as total_clicks
		FROM referrer_analytics 
		WHERE short_code = $1 
		  AND created_at >= NOW() - INTERVAL '%d days'
	`
	
	var uniqueDomains, totalRecords, utmClicks, totalClicks int64
	err = a.storage.QueryRow(fmt.Sprintf(additionalQuery, days), shortCode).Scan(&uniqueDomains, &totalRecords, &utmClicks, &totalClicks)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query referrer insights: %w", err)
	}
	
	insights["unique_referring_domains"] = uniqueDomains
	insights["total_referrer_records"] = totalRecords
	if totalClicks > 0 {
		insights["utm_tracking_percentage"] = float64(utmClicks) / float64(totalClicks) * 100
	}
	
	return insights, nil
}

// getGeographicAnalytics retrieves geographic analytics data
func (a *AdvancedAnalyticsService) GetGeographicAnalytics(shortCode string, days int) (*models.GeographicAnalytics, error) {
	geo := &models.GeographicAnalytics{
		ShortCode: shortCode,
	}
	
	// Get total clicks
	err := a.storage.QueryRow(
		`SELECT COALESCE(SUM(clicks), 0) FROM geographic_analytics WHERE short_code = $1`,
		shortCode,
	).Scan(&geo.TotalClicks)
	if err != nil {
		return nil, fmt.Errorf("failed to get total geographic clicks: %w", err)
	}
	
	// Get country details
	countryQuery := `
		SELECT country_code, country_name, SUM(clicks) as clicks, COUNT(DISTINCT id) as unique_locations, MAX(last_click) as last_click
		FROM geographic_analytics 
		WHERE short_code = $1 AND last_click >= NOW() - INTERVAL '%d days'
		GROUP BY country_code, country_name
		ORDER BY clicks DESC
		LIMIT 20
	`
	
	rows, err := a.storage.Query(fmt.Sprintf(countryQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var country models.CountryDetail
			var lastClick sql.NullTime
			err := rows.Scan(&country.CountryCode, &country.CountryName, &country.Clicks, &country.UniqueIPs, &lastClick)
			if err == nil {
				if geo.TotalClicks > 0 {
					country.Percentage = float64(country.Clicks) / float64(geo.TotalClicks) * 100
				}
				if lastClick.Valid {
					country.LastClick = &lastClick.Time
				}
				geo.Countries = append(geo.Countries, country)
			}
		}
	}
	
	// Get city details with coordinates for mapping
	cityQuery := `
		SELECT country_code, region, city, latitude, longitude, SUM(clicks) as clicks
		FROM geographic_analytics 
		WHERE short_code = $1 AND latitude IS NOT NULL AND longitude IS NOT NULL
		AND last_click >= NOW() - INTERVAL '%d days'
		GROUP BY country_code, region, city, latitude, longitude
		ORDER BY clicks DESC
		LIMIT 50
	`
	
	rows, err = a.storage.Query(fmt.Sprintf(cityQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var city models.CityDetail
			var mapPoint models.MapPoint
			err := rows.Scan(&city.CountryCode, &city.Region, &city.City, &city.Latitude, &city.Longitude, &city.Clicks)
			if err == nil {
				if geo.TotalClicks > 0 {
					city.Percentage = float64(city.Clicks) / float64(geo.TotalClicks) * 100
				}
				geo.Cities = append(geo.Cities, city)
				
				// Add to map data if coordinates are available
				if city.Latitude != nil && city.Longitude != nil {
					mapPoint.Latitude = *city.Latitude
					mapPoint.Longitude = *city.Longitude
					mapPoint.Clicks = city.Clicks
					mapPoint.CountryCode = city.CountryCode
					mapPoint.Location = fmt.Sprintf("%s, %s", city.City, city.CountryCode)
					geo.MapData = append(geo.MapData, mapPoint)
				}
			}
		}
	}
	
	return geo, nil
}

// getTimeAnalytics retrieves time-based analytics data
func (a *AdvancedAnalyticsService) GetTimeAnalytics(shortCode string, days int) (*models.TimeAnalytics, error) {
	timeAnalytics := &models.TimeAnalytics{
		ShortCode: shortCode,
	}
	
	// Get hourly pattern
	hourlyQuery := `
		SELECT hour_of_day, SUM(clicks) as clicks
		FROM time_analytics 
		WHERE short_code = $1 AND click_date >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY hour_of_day
		ORDER BY hour_of_day
	`
	
	rows, err := a.storage.Query(fmt.Sprintf(hourlyQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var hourly models.HourlyClick
			if err := rows.Scan(&hourly.Hour, &hourly.Clicks); err == nil {
				timeAnalytics.HourlyPattern = append(timeAnalytics.HourlyPattern, hourly)
			}
		}
	}
	
	// Get weekly pattern
	weeklyQuery := `
		SELECT day_of_week, SUM(clicks) as clicks
		FROM time_analytics 
		WHERE short_code = $1 AND click_date >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY day_of_week
		ORDER BY day_of_week
	`
	
	weekdays := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	rows, err = a.storage.Query(fmt.Sprintf(weeklyQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var weekly models.WeekdayClick
			if err := rows.Scan(&weekly.Weekday, &weekly.Clicks); err == nil {
				if weekly.Weekday >= 0 && weekly.Weekday < len(weekdays) {
					weekly.Day = weekdays[weekly.Weekday]
				}
				timeAnalytics.WeeklyPattern = append(timeAnalytics.WeeklyPattern, weekly)
			}
		}
	}
	
	// Get heatmap data (last 30 days)
	heatmapQuery := `
		SELECT click_date, hour_of_day, SUM(clicks) as clicks
		FROM time_analytics 
		WHERE short_code = $1 AND click_date >= CURRENT_DATE - INTERVAL '30 days'
		ORDER BY click_date, hour_of_day
	`
	
	rows, err = a.storage.Query(heatmapQuery, shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var heatmap models.HeatmapPoint
			var date time.Time
			if err := rows.Scan(&date, &heatmap.Hour, &heatmap.Clicks); err == nil {
				heatmap.Date = date.Format("2006-01-02")
				timeAnalytics.HeatmapData = append(timeAnalytics.HeatmapData, heatmap)
			}
		}
	}
	
	// Get peak times
	peakQuery := `
		SELECT hour_of_day, day_of_week, MAX(clicks) as max_clicks
		FROM time_analytics 
		WHERE short_code = $1 AND click_date >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY hour_of_day, day_of_week
		ORDER BY max_clicks DESC
		LIMIT 1
	`
	
	err = a.storage.QueryRow(fmt.Sprintf(peakQuery, days), shortCode).Scan(
		&timeAnalytics.PeakTimes.PeakHour,
		&timeAnalytics.PeakTimes.PeakWeekday,
		&timeAnalytics.PeakTimes.MaxClicks,
	)
	if err == nil && timeAnalytics.PeakTimes.PeakWeekday >= 0 && timeAnalytics.PeakTimes.PeakWeekday < len(weekdays) {
		timeAnalytics.PeakTimes.PeakDay = weekdays[timeAnalytics.PeakTimes.PeakWeekday]
	}
	
	return timeAnalytics, nil
}

// getDeviceAnalytics retrieves device analytics data
func (a *AdvancedAnalyticsService) GetDeviceAnalytics(shortCode string, days int) (*models.DeviceAnalytics, error) {
	deviceAnalytics := &models.DeviceAnalytics{
		ShortCode: shortCode,
	}
	
	// Get total clicks for percentage calculations
	var totalClicks int64
	err := a.storage.QueryRow(
		`SELECT COALESCE(SUM(clicks), 0) FROM device_analytics WHERE short_code = $1`,
		shortCode,
	).Scan(&totalClicks)
	if err != nil {
		return nil, fmt.Errorf("failed to get total device clicks: %w", err)
	}
	
	// Get device type statistics
	deviceTypeQuery := `
		SELECT device_type, SUM(clicks) as clicks
		FROM device_analytics 
		WHERE short_code = $1 AND last_click >= NOW() - INTERVAL '%d days'
		GROUP BY device_type
		ORDER BY clicks DESC
	`
	
	rows, err := a.storage.Query(fmt.Sprintf(deviceTypeQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var deviceType models.DeviceTypeStat
			if err := rows.Scan(&deviceType.DeviceType, &deviceType.Clicks); err == nil {
				if totalClicks > 0 {
					deviceType.Percentage = float64(deviceType.Clicks) / float64(totalClicks) * 100
				}
				deviceAnalytics.DeviceTypes = append(deviceAnalytics.DeviceTypes, deviceType)
			}
		}
	}
	
	// Get browser statistics
	browserQuery := `
		SELECT browser_name, browser_version, SUM(clicks) as clicks
		FROM device_analytics 
		WHERE short_code = $1 AND last_click >= NOW() - INTERVAL '%d days'
		AND browser_name IS NOT NULL AND browser_name != ''
		GROUP BY browser_name, browser_version
		ORDER BY clicks DESC
		LIMIT 20
	`
	
	rows, err = a.storage.Query(fmt.Sprintf(browserQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var browser models.BrowserDetailStat
			if err := rows.Scan(&browser.BrowserName, &browser.BrowserVersion, &browser.Clicks); err == nil {
				if totalClicks > 0 {
					browser.Percentage = float64(browser.Clicks) / float64(totalClicks) * 100
				}
				deviceAnalytics.BrowserStats = append(deviceAnalytics.BrowserStats, browser)
			}
		}
	}
	
	// Get OS statistics
	osQuery := `
		SELECT os_name, os_version, SUM(clicks) as clicks
		FROM device_analytics 
		WHERE short_code = $1 AND last_click >= NOW() - INTERVAL '%d days'
		AND os_name IS NOT NULL AND os_name != ''
		GROUP BY os_name, os_version
		ORDER BY clicks DESC
		LIMIT 20
	`
	
	rows, err = a.storage.Query(fmt.Sprintf(osQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var os models.OSDetailStat
			if err := rows.Scan(&os.OSName, &os.OSVersion, &os.Clicks); err == nil {
				if totalClicks > 0 {
					os.Percentage = float64(os.Clicks) / float64(totalClicks) * 100
				}
				deviceAnalytics.OSStats = append(deviceAnalytics.OSStats, os)
			}
		}
	}
	
	return deviceAnalytics, nil
}

// getReferrerAnalytics retrieves referrer analytics data
func (a *AdvancedAnalyticsService) GetReferrerAnalytics(shortCode string, days int) ([]models.ReferrerStat, error) {
	var referrers []models.ReferrerStat
	
	query := `
		SELECT referrer_domain, SUM(clicks) as clicks
		FROM referrer_analytics 
		WHERE short_code = $1 AND last_click >= NOW() - INTERVAL '%d days'
		AND referrer_domain IS NOT NULL AND referrer_domain != ''
		GROUP BY referrer_domain
		ORDER BY clicks DESC
		LIMIT 20
	`
	
	rows, err := a.storage.Query(fmt.Sprintf(query, days), shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get referrer analytics: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var referrer models.ReferrerStat
		if err := rows.Scan(&referrer.Referrer, &referrer.Clicks); err == nil {
			referrers = append(referrers, referrer)
		}
	}
	
	return referrers, nil
}