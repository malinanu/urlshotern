package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(config *configs.Config) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.DBMaxConns)
	db.SetMaxIdleConns(config.DBMaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}
	
	// Initialize tables if they don't exist
	if err := storage.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

// Database access methods for authentication services

// QueryRow executes a query that returns at most one row
func (p *PostgresStorage) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

// Exec executes a query that doesn't return rows
func (p *PostgresStorage) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

// Query executes a query that returns multiple rows
func (p *PostgresStorage) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

// Begin starts a database transaction
func (p *PostgresStorage) Begin() (*sql.Tx, error) {
	return p.db.Begin()
}

// initTables creates the necessary tables if they don't exist
func (p *PostgresStorage) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS url_mappings (
			id BIGINT PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE,
			click_count BIGINT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_by_ip INET
		)`,
		`CREATE INDEX IF NOT EXISTS idx_short_code ON url_mappings(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_created_at ON url_mappings(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_active ON url_mappings(is_active)`,
		`CREATE TABLE IF NOT EXISTS click_events (
			id BIGINT PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			ip_address INET,
			user_agent TEXT,
			referrer TEXT,
			country_code CHAR(2)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_short_code_time ON click_events(short_code, clicked_at)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	// Initialize enhanced analytics tables
	if err := p.initAdvancedAnalyticsTables(); err != nil {
		return fmt.Errorf("failed to initialize advanced analytics tables: %w", err)
	}

	return nil
}

// initAdvancedAnalyticsTables creates enhanced analytics tables
func (p *PostgresStorage) initAdvancedAnalyticsTables() error {
	queries := []string{
		// Enhanced geographic analytics table
		`CREATE TABLE IF NOT EXISTS geographic_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			country_code CHAR(2),
			country_name VARCHAR(100),
			region VARCHAR(100),
			city VARCHAR(100),
			latitude DECIMAL(10,8),
			longitude DECIMAL(11,8),
			clicks INTEGER DEFAULT 1,
			unique_ips INTEGER DEFAULT 1,
			last_click TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(short_code, country_code, region, city),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_geographic_short_code ON geographic_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_geographic_country ON geographic_analytics(country_code)`,
		`CREATE INDEX IF NOT EXISTS idx_geographic_location ON geographic_analytics(country_code, region, city)`,
		
		// Enhanced device analytics table
		`CREATE TABLE IF NOT EXISTS device_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			device_type VARCHAR(50), -- mobile, desktop, tablet
			device_brand VARCHAR(50), -- Apple, Samsung, etc.
			device_model VARCHAR(100),
			os_name VARCHAR(50),
			os_version VARCHAR(50),
			browser_name VARCHAR(50),
			browser_version VARCHAR(50),
			screen_resolution VARCHAR(20),
			user_agent_hash VARCHAR(64), -- SHA-256 hash for grouping
			clicks INTEGER DEFAULT 1,
			unique_ips INTEGER DEFAULT 1,
			last_click TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(short_code, user_agent_hash),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_device_short_code ON device_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_device_type ON device_analytics(device_type)`,
		`CREATE INDEX IF NOT EXISTS idx_device_brand ON device_analytics(device_brand, device_model)`,
		`CREATE INDEX IF NOT EXISTS idx_device_os ON device_analytics(os_name, os_version)`,
		`CREATE INDEX IF NOT EXISTS idx_device_browser ON device_analytics(browser_name, browser_version)`,
		
		// Time-based analytics table
		`CREATE TABLE IF NOT EXISTS time_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			click_date DATE NOT NULL,
			hour_of_day INTEGER CHECK (hour_of_day >= 0 AND hour_of_day <= 23),
			day_of_week INTEGER CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday
			day_of_month INTEGER CHECK (day_of_month >= 1 AND day_of_month <= 31),
			month INTEGER CHECK (month >= 1 AND month <= 12),
			year INTEGER,
			clicks INTEGER DEFAULT 1,
			unique_ips INTEGER DEFAULT 1,
			conversions INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			UNIQUE(short_code, click_date, hour_of_day)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_time_short_code ON time_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_time_date ON time_analytics(click_date)`,
		`CREATE INDEX IF NOT EXISTS idx_time_hour ON time_analytics(hour_of_day)`,
		`CREATE INDEX IF NOT EXISTS idx_time_day_week ON time_analytics(day_of_week)`,
		`CREATE INDEX IF NOT EXISTS idx_time_heatmap ON time_analytics(short_code, click_date, hour_of_day)`,
		
		// Enhanced referrer analytics table
		`CREATE TABLE IF NOT EXISTS referrer_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			referrer_domain VARCHAR(255),
			referrer_url TEXT,
			referrer_type VARCHAR(50), -- social, search, direct, email, etc.
			campaign_source VARCHAR(100), -- UTM tracking
			campaign_medium VARCHAR(100),
			campaign_name VARCHAR(100),
			campaign_term VARCHAR(100),
			campaign_content VARCHAR(100),
			clicks INTEGER DEFAULT 1,
			unique_clicks INTEGER DEFAULT 1,
			conversions INTEGER DEFAULT 0,
			last_click TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(short_code, referrer_domain, campaign_source, campaign_medium),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_short_code ON referrer_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_domain ON referrer_analytics(referrer_domain)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_type ON referrer_analytics(referrer_type)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_campaign ON referrer_analytics(campaign_source, campaign_medium)`,
		
		// Conversion tracking tables
		`CREATE TABLE IF NOT EXISTS conversion_goals (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			goal_name VARCHAR(100) NOT NULL,
			goal_type VARCHAR(50) NOT NULL, -- url_visit, custom_event, form_submit, purchase
			target_url TEXT,
			custom_event_name VARCHAR(100),
			goal_value DECIMAL(10,2) DEFAULT 0, -- monetary value if applicable
			attribution_window INTEGER DEFAULT 30, -- days
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conversion_goals_user ON conversion_goals(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversion_goals_active ON conversion_goals(is_active)`,
		
		// Conversions tracking table
		`CREATE TABLE IF NOT EXISTS conversions (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			goal_id BIGINT NOT NULL,
			conversion_id VARCHAR(100) UNIQUE NOT NULL, -- unique identifier for deduplication
			conversion_type VARCHAR(50) NOT NULL,
			conversion_value DECIMAL(10,2) DEFAULT 0,
			user_ip INET,
			user_agent TEXT,
			referrer TEXT,
			session_id VARCHAR(100), -- for session tracking
			click_id BIGINT, -- link to original click event
			conversion_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			attribution_model VARCHAR(50) DEFAULT 'last_click', -- first_click, last_click, linear
			time_to_conversion INTEGER, -- minutes from click to conversion
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			FOREIGN KEY (goal_id) REFERENCES conversion_goals(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_short_code ON conversions(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_goal ON conversions(goal_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_time ON conversions(conversion_time)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_session ON conversions(session_id)`,
		
		// A/B testing tables
		`CREATE TABLE IF NOT EXISTS ab_tests (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			test_name VARCHAR(100) NOT NULL,
			description TEXT,
			short_code_a VARCHAR(10) NOT NULL, -- Control variant
			short_code_b VARCHAR(10) NOT NULL, -- Test variant
			traffic_split INTEGER DEFAULT 50, -- percentage to variant B (0-100)
			start_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			end_date TIMESTAMP WITH TIME ZONE,
			status VARCHAR(20) DEFAULT 'active', -- draft, active, paused, completed
			winner VARCHAR(1), -- 'A', 'B', or NULL if inconclusive
			confidence_level DECIMAL(5,2), -- statistical confidence percentage
			min_sample_size INTEGER DEFAULT 100,
			conversion_goal_id BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (short_code_a) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			FOREIGN KEY (short_code_b) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			FOREIGN KEY (conversion_goal_id) REFERENCES conversion_goals(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_tests_user ON ab_tests(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_tests_status ON ab_tests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_tests_dates ON ab_tests(start_date, end_date)`,
		
		// A/B test results tracking
		`CREATE TABLE IF NOT EXISTS ab_test_results (
			id BIGSERIAL PRIMARY KEY,
			test_id BIGINT NOT NULL,
			variant VARCHAR(1) NOT NULL, -- 'A' or 'B'
			date DATE DEFAULT CURRENT_DATE,
			clicks INTEGER DEFAULT 0,
			conversions INTEGER DEFAULT 0,
			conversion_rate DECIMAL(6,4) DEFAULT 0, -- stored as percentage (0-100.00)
			revenue DECIMAL(10,2) DEFAULT 0,
			unique_visitors INTEGER DEFAULT 0,
			bounce_rate DECIMAL(5,2) DEFAULT 0,
			avg_time_on_site INTEGER DEFAULT 0, -- seconds
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (test_id) REFERENCES ab_tests(id) ON DELETE CASCADE,
			UNIQUE(test_id, variant, date)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_test_results_test ON ab_test_results(test_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_test_results_date ON ab_test_results(date)`,
		
		// Attribution tracking for better conversion analysis
		`CREATE TABLE IF NOT EXISTS attribution_touchpoints (
			id BIGSERIAL PRIMARY KEY,
			session_id VARCHAR(100) NOT NULL,
			short_code VARCHAR(10) NOT NULL,
			user_ip INET,
			user_agent TEXT,
			referrer TEXT,
			campaign_source VARCHAR(100),
			campaign_medium VARCHAR(100),
			campaign_name VARCHAR(100),
			touchpoint_order INTEGER NOT NULL, -- 1st touch, 2nd touch, etc.
			touchpoint_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			conversion_id VARCHAR(100), -- link to final conversion if exists
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_attribution_session ON attribution_touchpoints(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_attribution_conversion ON attribution_touchpoints(conversion_id)`,
		`CREATE INDEX IF NOT EXISTS idx_attribution_time ON attribution_touchpoints(touchpoint_time)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// SaveURLMapping saves a URL mapping to the database
func (p *PostgresStorage) SaveURLMapping(mapping *models.URLMapping) error {
	query := `
		INSERT INTO url_mappings (id, short_code, original_url, created_at, expires_at, created_by_ip, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.db.Exec(query, mapping.ID, mapping.ShortCode, mapping.OriginalURL,
		mapping.CreatedAt, mapping.ExpiresAt, mapping.CreatedByIP, mapping.UserID)
	
	if err != nil {
		return fmt.Errorf("failed to save URL mapping: %w", err)
	}
	return nil
}

// GetURLMappingByShortCode retrieves a URL mapping by its short code
func (p *PostgresStorage) GetURLMappingByShortCode(shortCode string) (*models.URLMapping, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count, is_active, created_by_ip, user_id
		FROM url_mappings
		WHERE short_code = $1 AND is_active = TRUE
	`
	
	mapping := &models.URLMapping{}
	var expiresAt sql.NullTime
	var createdByIP sql.NullString
	var userID sql.NullInt64
	
	err := p.db.QueryRow(query, shortCode).Scan(
		&mapping.ID, &mapping.ShortCode, &mapping.OriginalURL,
		&mapping.CreatedAt, &expiresAt, &mapping.ClickCount,
		&mapping.IsActive, &createdByIP, &userID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL mapping: %w", err)
	}

	if expiresAt.Valid {
		mapping.ExpiresAt = &expiresAt.Time
	}
	if createdByIP.Valid {
		mapping.CreatedByIP = createdByIP.String
	}
	if userID.Valid {
		mapping.UserID = &userID.Int64
	}

	// Check if URL has expired
	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now()) {
		return nil, ErrURLExpired
	}

	return mapping, nil
}

// IncrementClickCount increments the click count for a URL mapping
func (p *PostgresStorage) IncrementClickCount(shortCode string) error {
	query := `UPDATE url_mappings SET click_count = click_count + 1 WHERE short_code = $1`
	_, err := p.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}
	return nil
}

// SaveClickEvent saves a click event to the database
func (p *PostgresStorage) SaveClickEvent(event *models.ClickEvent) error {
	query := `
		INSERT INTO click_events (id, short_code, clicked_at, ip_address, user_agent, referrer, country_code)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.db.Exec(query, event.ID, event.ShortCode, event.ClickedAt,
		event.IPAddress, event.UserAgent, event.Referrer, event.CountryCode)
	
	if err != nil {
		return fmt.Errorf("failed to save click event: %w", err)
	}
	return nil
}

// GetAnalytics retrieves analytics data for a short code
func (p *PostgresStorage) GetAnalytics(shortCode string, days int) (*models.AnalyticsResponse, error) {
	// Get basic URL info
	mapping, err := p.GetURLMappingByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	analytics := &models.AnalyticsResponse{
		ShortCode:   mapping.ShortCode,
		OriginalURL: mapping.OriginalURL,
		TotalClicks: mapping.ClickCount,
		CreatedAt:   mapping.CreatedAt,
	}

	// Get last click time
	var lastClickAt sql.NullTime
	err = p.db.QueryRow(
		`SELECT MAX(clicked_at) FROM click_events WHERE short_code = $1`,
		shortCode,
	).Scan(&lastClickAt)
	if err == nil && lastClickAt.Valid {
		analytics.LastClickAt = &lastClickAt.Time
	}

	// Get daily clicks for the specified number of days
	dailyClicksQuery := `
		SELECT DATE(clicked_at) as date, COUNT(*) as clicks
		FROM click_events
		WHERE short_code = $1 AND clicked_at >= NOW() - INTERVAL '%d days'
		GROUP BY DATE(clicked_at)
		ORDER BY date DESC
	`
	rows, err := p.db.Query(fmt.Sprintf(dailyClicksQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dailyClick models.DailyClick
			if err := rows.Scan(&dailyClick.Date, &dailyClick.Clicks); err == nil {
				analytics.DailyClicks = append(analytics.DailyClicks, dailyClick)
			}
		}
	}

	// Get country stats
	countryStatsQuery := `
		SELECT country_code, COUNT(*) as clicks
		FROM click_events
		WHERE short_code = $1 AND country_code IS NOT NULL
		GROUP BY country_code
		ORDER BY clicks DESC
		LIMIT 10
	`
	rows, err = p.db.Query(countryStatsQuery, shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var countryStat models.CountryStat
			if err := rows.Scan(&countryStat.CountryCode, &countryStat.Clicks); err == nil {
				analytics.CountryStats = append(analytics.CountryStats, countryStat)
			}
		}
	}

	return analytics, nil
}

// ShortCodeExists checks if a short code already exists
func (p *PostgresStorage) ShortCodeExists(shortCode string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM url_mappings WHERE short_code = $1)`
	err := p.db.QueryRow(query, shortCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if short code exists: %w", err)
	}
	return exists, nil
}

// User management methods

// CreateUser creates a new user in the database
func (p *PostgresStorage) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, name, email, phone, password_hash, email_verified, phone_verified, 
			              provider, provider_id, avatar_url, account_type, is_active, is_admin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := p.Exec(query, user.ID, user.Name, user.Email, user.Phone, user.PasswordHash,
		user.EmailVerified, user.PhoneVerified, user.Provider, user.ProviderID, user.AvatarURL,
		user.AccountType, user.IsActive, user.IsAdmin, user.CreatedAt, user.UpdatedAt)
	return err
}

// GetUserByID retrieves a user by ID
func (p *PostgresStorage) GetUserByID(id int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, email, phone, password_hash, email_verified, phone_verified,
		       provider, provider_id, avatar_url, account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`
	err := p.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash,
		&user.EmailVerified, &user.PhoneVerified, &user.Provider, &user.ProviderID,
		&user.AvatarURL, &user.AccountType, &user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (p *PostgresStorage) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, email, phone, password_hash, email_verified, phone_verified,
		       provider, provider_id, avatar_url, account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`
	err := p.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash,
		&user.EmailVerified, &user.PhoneVerified, &user.Provider, &user.ProviderID,
		&user.AvatarURL, &user.AccountType, &user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByPhone retrieves a user by phone
func (p *PostgresStorage) GetUserByPhone(phone string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, email, phone, password_hash, email_verified, phone_verified,
		       provider, provider_id, avatar_url, account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users WHERE phone = $1
	`
	err := p.QueryRow(query, phone).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash,
		&user.EmailVerified, &user.PhoneVerified, &user.Provider, &user.ProviderID,
		&user.AvatarURL, &user.AccountType, &user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserLastLogin updates user's last login time
func (p *PostgresStorage) UpdateUserLastLogin(userID int64, loginTime time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, loginTime, time.Now(), userID)
	return err
}

// CreateUserPreferences creates user preferences
func (p *PostgresStorage) CreateUserPreferences(prefs *models.UserPreferences) error {
	query := `
		INSERT INTO user_preferences (user_id, default_expiration, analytics_public, 
		                            email_notifications, marketing_emails, timezone, theme)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.Exec(query, prefs.UserID, prefs.DefaultExpiration, prefs.AnalyticsPublic,
		prefs.EmailNotifications, prefs.MarketingEmails, prefs.Timezone, prefs.Theme)
	return err
}

// CreateSession creates a new user session
func (p *PostgresStorage) CreateSession(session *models.UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, device_info, ip_address, user_agent,
		                         refresh_token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := p.Exec(query, session.ID, session.UserID, session.DeviceInfo, session.IPAddress, 
		session.UserAgent, session.RefreshTokenHash, session.ExpiresAt, session.CreatedAt)
	return err
}

// GetSessionByID retrieves a session by ID
func (p *PostgresStorage) GetSessionByID(id string) (*models.UserSession, error) {
	session := &models.UserSession{}
	query := `
		SELECT id, user_id, device_info, ip_address, user_agent, 
		       refresh_token_hash, expires_at, created_at
		FROM user_sessions WHERE id = $1
	`
	err := p.QueryRow(query, id).Scan(
		&session.ID, &session.UserID, &session.DeviceInfo, &session.IPAddress, 
		&session.UserAgent, &session.RefreshTokenHash, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Email verification methods

// CreateEmailVerification creates an email verification record
func (p *PostgresStorage) CreateEmailVerification(verification *models.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := p.Exec(query, verification.ID, verification.UserID, verification.Token,
		verification.ExpiresAt, verification.CreatedAt)
	return err
}

// GetEmailVerificationByToken retrieves email verification by token
func (p *PostgresStorage) GetEmailVerificationByToken(token string) (*models.EmailVerification, error) {
	verification := &models.EmailVerification{}
	query := `
		SELECT id, user_id, token, expires_at, verified_at, created_at
		FROM email_verifications WHERE token = $1
	`
	err := p.QueryRow(query, token).Scan(
		&verification.ID, &verification.UserID, &verification.Token,
		&verification.ExpiresAt, &verification.VerifiedAt, &verification.CreatedAt)
	if err != nil {
		return nil, err
	}
	return verification, nil
}

// UpdateEmailVerification updates email verification record
func (p *PostgresStorage) UpdateEmailVerification(id string, verifiedAt time.Time) error {
	query := `UPDATE email_verifications SET verified_at = $1 WHERE id = $2`
	_, err := p.Exec(query, verifiedAt, id)
	return err
}

// UpdateUserEmailVerified updates user email verification status
func (p *PostgresStorage) UpdateUserEmailVerified(userID int64, verified bool) error {
	query := `UPDATE users SET email_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, verified, time.Now(), userID)
	return err
}

// Phone verification methods

// CreatePhoneVerification creates a phone verification record
func (p *PostgresStorage) CreatePhoneVerification(verification *models.PhoneVerification) error {
	query := `
		INSERT INTO phone_verifications (id, user_id, phone, otp_code, attempts, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.Exec(query, verification.ID, verification.UserID, verification.Phone,
		verification.OTPCode, verification.Attempts, verification.ExpiresAt, verification.CreatedAt)
	return err
}

// DeletePhoneVerificationByUserID deletes phone verification by user ID
func (p *PostgresStorage) DeletePhoneVerificationByUserID(userID int64) error {
	query := `DELETE FROM phone_verifications WHERE user_id = $1`
	_, err := p.Exec(query, userID)
	return err
}

// GetPhoneVerificationByUserID retrieves phone verification by user ID
func (p *PostgresStorage) GetPhoneVerificationByUserID(userID int64) (*models.PhoneVerification, error) {
	verification := &models.PhoneVerification{}
	query := `
		SELECT id, user_id, phone, otp_code, attempts, expires_at, verified_at, created_at
		FROM phone_verifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1
	`
	err := p.QueryRow(query, userID).Scan(
		&verification.ID, &verification.UserID, &verification.Phone,
		&verification.OTPCode, &verification.Attempts, &verification.ExpiresAt, &verification.VerifiedAt, &verification.CreatedAt)
	if err != nil {
		return nil, err
	}
	return verification, nil
}

// UpdatePhoneVerificationAttempts updates phone verification attempts
func (p *PostgresStorage) UpdatePhoneVerificationAttempts(userID int64, attempts int) error {
	query := `UPDATE phone_verifications SET attempts = $1 WHERE user_id = $2`
	_, err := p.Exec(query, attempts, userID)
	return err
}

// UpdatePhoneVerification updates phone verification record
func (p *PostgresStorage) UpdatePhoneVerification(id string, verifiedAt time.Time) error {
	query := `UPDATE phone_verifications SET verified_at = $1 WHERE id = $2`
	_, err := p.Exec(query, verifiedAt, id)
	return err
}

// UpdateUserPhoneVerified updates user phone verification status
func (p *PostgresStorage) UpdateUserPhoneVerified(userID int64, verified bool) error {
	query := `UPDATE users SET phone_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, verified, time.Now(), userID)
	return err
}

// Password reset methods

// CreatePasswordReset creates a password reset record
func (p *PostgresStorage) CreatePasswordReset(reset *models.PasswordReset) error {
	query := `
		INSERT INTO password_resets (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := p.Exec(query, reset.ID, reset.UserID, reset.Token,
		reset.ExpiresAt, reset.CreatedAt)
	return err
}

// GetPasswordResetByToken retrieves password reset by token
func (p *PostgresStorage) GetPasswordResetByToken(token string) (*models.PasswordReset, error) {
	reset := &models.PasswordReset{}
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_resets WHERE token = $1
	`
	err := p.QueryRow(query, token).Scan(
		&reset.ID, &reset.UserID, &reset.Token,
		&reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt)
	if err != nil {
		return nil, err
	}
	return reset, nil
}

// UpdatePasswordReset updates password reset record
func (p *PostgresStorage) UpdatePasswordReset(id string, usedAt time.Time) error {
	query := `UPDATE password_resets SET used_at = $1 WHERE id = $2`
	_, err := p.Exec(query, usedAt, id)
	return err
}

// UpdateUserPassword updates user password
func (p *PostgresStorage) UpdateUserPassword(userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, passwordHash, time.Now(), userID)
	return err
}

// User management methods

// UpdateUser updates user information
func (p *PostgresStorage) UpdateUser(user *models.User) error {
	query := `
		UPDATE users SET name = $1, email = $2, phone = $3, provider = $4, 
		                provider_id = $5, avatar_url = $6, account_type = $7, 
		                updated_at = $8 
		WHERE id = $9
	`
	_, err := p.Exec(query, user.Name, user.Email, user.Phone, user.Provider,
		user.ProviderID, user.AvatarURL, user.AccountType, time.Now(), user.ID)
	return err
}

// UpdateUserStatus updates user status (active/inactive)
func (p *PostgresStorage) UpdateUserStatus(userID int64, isActive bool) error {
	query := `UPDATE users SET is_active = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, isActive, time.Now(), userID)
	return err
}

// GetUserPreferences retrieves user preferences
func (p *PostgresStorage) GetUserPreferences(userID int64) (*models.UserPreferences, error) {
	prefs := &models.UserPreferences{}
	query := `
		SELECT user_id, default_expiration, analytics_public, email_notifications,
		       marketing_emails, timezone, theme
		FROM user_preferences WHERE user_id = $1
	`
	err := p.QueryRow(query, userID).Scan(
		&prefs.UserID, &prefs.DefaultExpiration, &prefs.AnalyticsPublic,
		&prefs.EmailNotifications, &prefs.MarketingEmails, &prefs.Timezone, &prefs.Theme)
	if err != nil {
		return nil, err
	}
	return prefs, nil
}

// UpdateUserPreferences updates user preferences
func (p *PostgresStorage) UpdateUserPreferences(prefs *models.UserPreferences) error {
	query := `
		UPDATE user_preferences 
		SET default_expiration = $1, analytics_public = $2, email_notifications = $3,
		    marketing_emails = $4, timezone = $5, theme = $6
		WHERE user_id = $7
	`
	_, err := p.Exec(query, prefs.DefaultExpiration, prefs.AnalyticsPublic,
		prefs.EmailNotifications, prefs.MarketingEmails, prefs.Timezone, prefs.Theme, prefs.UserID)
	return err
}

// Custom errors
var (
	ErrURLNotFound   = &StorageError{Message: "URL not found"}
	ErrURLExpired    = &StorageError{Message: "URL has expired"}
	ErrUnauthorized  = &StorageError{Message: "Unauthorized access"}
)

type StorageError struct {
	Message string
}

func (e *StorageError) Error() string {
	return e.Message
}