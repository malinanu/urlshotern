package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// DatabaseConfig represents a database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	MaxConns        int
	MaxIdleConns    int
	ConnMaxLifetime int // in seconds
	ConnMaxIdleTime int // in seconds
}

// ConnectionString returns the PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" sslmode=disable"
}

type Config struct {
	// Server Configuration
	ServerPort  string
	ServerHost  string
	BaseURL     string
	Environment string

	// Database Configuration (Legacy - for backward compatibility)
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxConns        int
	DBMaxIdleConns    int
	DBConnMaxLifetime int // in seconds
	DBConnMaxIdleTime int // in seconds

	// Multi-User Database Configuration
	AppDB      DatabaseConfig // Regular application user (limited permissions)
	AdminDB    DatabaseConfig // Admin operations user (full permissions)
	ReadOnlyDB DatabaseConfig // Read-only user (for analytics)

	// Redis Configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Snowflake Configuration
	NodeID int64

	// Logging
	LogLevel  string
	LogFormat string

	// Google OAuth Configuration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// JWT Configuration
	JWTSecret string
	JWTIssuer string

	// Polar Configuration
	PolarAccessToken    string
	PolarWebhookSecret  string
	PolarOrganizationID string
	PolarAPIURL         string

	// Email Configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	FrontendURL  string

	// Webhook Alert Configuration
	WebhookAlertEmails string

	// Product Sync Configuration
	ProductSyncIntervalMinutes int // How often to sync products from Polar (in minutes)
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Legacy database config (for backward compatibility)
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "urlshortener_db")

	config := &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Legacy database configuration (for backward compatibility)
		DBHost:            dbHost,
		DBPort:            dbPort,
		DBUser:            getEnv("DB_USER", "urlshortener"),
		DBPassword:        getEnv("DB_PASSWORD", "password"),
		DBName:            dbName,
		DBMaxConns:        getEnvAsInt("DB_MAX_CONNECTIONS", 25),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
		DBConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600),
		DBConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 600),

		// Multi-User Database Configuration
		AppDB: DatabaseConfig{
			Host:            getEnv("DB_APP_HOST", dbHost),
			Port:            getEnv("DB_APP_PORT", dbPort),
			User:            getEnv("DB_APP_USER", "urlshort_app"),
			Password:        getEnv("DB_APP_PASSWORD", ""),
			DBName:          getEnv("DB_APP_NAME", dbName),
			MaxConns:        getEnvAsInt("DB_APP_MAX_CONNECTIONS", 50),
			MaxIdleConns:    getEnvAsInt("DB_APP_MAX_IDLE_CONNECTIONS", 10),
			ConnMaxLifetime: getEnvAsInt("DB_APP_CONN_MAX_LIFETIME", 3600),
			ConnMaxIdleTime: getEnvAsInt("DB_APP_CONN_MAX_IDLE_TIME", 600),
		},

		AdminDB: DatabaseConfig{
			Host:            getEnv("DB_ADMIN_HOST", dbHost),
			Port:            getEnv("DB_ADMIN_PORT", dbPort),
			User:            getEnv("DB_ADMIN_USER", "urlshort_admin"),
			Password:        getEnv("DB_ADMIN_PASSWORD", ""),
			DBName:          getEnv("DB_ADMIN_NAME", dbName),
			MaxConns:        getEnvAsInt("DB_ADMIN_MAX_CONNECTIONS", 10),
			MaxIdleConns:    getEnvAsInt("DB_ADMIN_MAX_IDLE_CONNECTIONS", 2),
			ConnMaxLifetime: getEnvAsInt("DB_ADMIN_CONN_MAX_LIFETIME", 3600),
			ConnMaxIdleTime: getEnvAsInt("DB_ADMIN_CONN_MAX_IDLE_TIME", 600),
		},

		ReadOnlyDB: DatabaseConfig{
			Host:            getEnv("DB_READONLY_HOST", dbHost),
			Port:            getEnv("DB_READONLY_PORT", dbPort),
			User:            getEnv("DB_READONLY_USER", "urlshort_readonly"),
			Password:        getEnv("DB_READONLY_PASSWORD", ""),
			DBName:          getEnv("DB_READONLY_NAME", dbName),
			MaxConns:        getEnvAsInt("DB_READONLY_MAX_CONNECTIONS", 20),
			MaxIdleConns:    getEnvAsInt("DB_READONLY_MAX_IDLE_CONNECTIONS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_READONLY_CONN_MAX_LIFETIME", 3600),
			ConnMaxIdleTime: getEnvAsInt("DB_READONLY_CONN_MAX_IDLE_TIME", 600),
		},

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		NodeID: getEnvAsInt64("NODE_ID", 1),

		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),

		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		JWTIssuer: getEnv("JWT_ISSUER", "urlshortener"),

		PolarAccessToken:    getEnv("POLAR_ACCESS_TOKEN", ""),
		PolarWebhookSecret:  getEnv("POLAR_WEBHOOK_SECRET", ""),
		PolarOrganizationID: getEnv("POLAR_ORGANIZATION_ID", ""),
		PolarAPIURL:         getEnv("POLAR_API_URL", "https://api.polar.sh/v1"),

		// Email Configuration
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@urlshortener.com"),
		FromName:     getEnv("FROM_NAME", "URL Shortener"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),

		// Webhook Alert Configuration
		WebhookAlertEmails: getEnv("WEBHOOK_ALERT_EMAILS", ""),

		// Product Sync Configuration
		ProductSyncIntervalMinutes: getEnvAsInt("PRODUCT_SYNC_INTERVAL_MINUTES", 60), // Default: sync every hour
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}