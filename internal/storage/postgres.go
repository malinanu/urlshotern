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

	return nil
}

// SaveURLMapping saves a URL mapping to the database
func (p *PostgresStorage) SaveURLMapping(mapping *models.URLMapping) error {
	query := `
		INSERT INTO url_mappings (id, short_code, original_url, created_at, expires_at, created_by_ip)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := p.db.Exec(query, mapping.ID, mapping.ShortCode, mapping.OriginalURL,
		mapping.CreatedAt, mapping.ExpiresAt, mapping.CreatedByIP)
	
	if err != nil {
		return fmt.Errorf("failed to save URL mapping: %w", err)
	}
	return nil
}

// GetURLMappingByShortCode retrieves a URL mapping by its short code
func (p *PostgresStorage) GetURLMappingByShortCode(shortCode string) (*models.URLMapping, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count, is_active, created_by_ip
		FROM url_mappings
		WHERE short_code = $1 AND is_active = TRUE
	`
	
	mapping := &models.URLMapping{}
	var expiresAt sql.NullTime
	var createdByIP sql.NullString
	
	err := p.db.QueryRow(query, shortCode).Scan(
		&mapping.ID, &mapping.ShortCode, &mapping.OriginalURL,
		&mapping.CreatedAt, &expiresAt, &mapping.ClickCount,
		&mapping.IsActive, &createdByIP,
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

// Custom errors
var (
	ErrURLNotFound = &StorageError{Message: "URL not found"}
	ErrURLExpired  = &StorageError{Message: "URL has expired"}
)

type StorageError struct {
	Message string
}

func (e *StorageError) Error() string {
	return e.Message
}