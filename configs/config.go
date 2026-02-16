package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server Configuration
	ServerPort string
	ServerHost string
	BaseURL    string
	Environment string

	// Database Configuration
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	DBMaxConns          int
	DBMaxIdleConns      int
	DBConnMaxLifetime   int // in seconds
	DBConnMaxIdleTime   int // in seconds

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

	// Webhook Alert Configuration
	WebhookAlertEmails string

	// Product Sync Configuration
	ProductSyncIntervalMinutes int // How often to sync products from Polar (in minutes)

	// Admin Panel Control
	EnableAdminRoutes bool // Enable admin routes at /api/v1/admin/* (default: false for security)
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		ServerHost:  getEnv("SERVER_HOST", "localhost"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "urlshortener"),
		DBPassword:        getEnv("DB_PASSWORD", "password"),
		DBName:            getEnv("DB_NAME", "urlshortener_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxConns:        getEnvAsInt("DB_MAX_CONNECTIONS", 25),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
		DBConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600),
		DBConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 600),

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
		FromEmail:    getEnv("FROM_EMAIL", "noreply@trunc.site"),
		FromName:     getEnv("FROM_NAME", "trunc.site"),

		// Webhook Alert Configuration
		WebhookAlertEmails: getEnv("WEBHOOK_ALERT_EMAILS", ""),

		// Product Sync Configuration
		ProductSyncIntervalMinutes: getEnvAsInt("PRODUCT_SYNC_INTERVAL_MINUTES", 60), // Default: sync every hour

		// Admin Panel Control (default: false for security)
		EnableAdminRoutes: getEnvAsBool("ENABLE_ADMIN_ROUTES", false),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}