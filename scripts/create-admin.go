package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/URLshorter/url-shortener/internal/utils"
)

func main() {
	// Command line flags
	name := flag.String("name", "", "Admin name")
	email := flag.String("email", "", "Admin email")
	password := flag.String("password", "", "Admin password")
	accountType := flag.String("type", "super_admin", "Account type (super_admin, system_admin, support_admin, billing_admin, content_admin)")

	flag.Parse()

	// Validate inputs
	if *name == "" || *email == "" || *password == "" {
		fmt.Println("Usage: go run scripts/create-admin.go -name=\"Admin Name\" -email=\"admin@example.com\" -password=\"securepassword\" [-type=\"super_admin\"]")
		fmt.Println("\nAccount types:")
		fmt.Println("  - super_admin: Full system access (default)")
		fmt.Println("  - system_admin: System management")
		fmt.Println("  - support_admin: User support")
		fmt.Println("  - billing_admin: Billing management")
		fmt.Println("  - content_admin: Content management")
		os.Exit(1)
	}

	// Validate account type
	validTypes := map[string]bool{
		"super_admin":   true,
		"system_admin":  true,
		"support_admin": true,
		"billing_admin": true,
		"content_admin": true,
	}
	if !validTypes[*accountType] {
		log.Fatalf("Invalid account type: %s", *accountType)
	}

	// Get database connection string from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	// Use admin user for creating admin accounts
	dbUser := getEnv("DB_ADMIN_USER", "urlshort_admin")
	dbPassword := getEnv("DB_ADMIN_PASSWORD", "admin_dev_password")
	dbName := getEnv("DB_NAME", "urlshortener_db")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize Snowflake ID generator (using node ID 1 for admin creation)
	if err := utils.InitializeSnowflake(1); err != nil {
		log.Fatalf("Failed to initialize Snowflake: %v", err)
	}

	// Generate unique ID for user
	userID, err := utils.GenerateID()
	if err != nil {
		log.Fatalf("Failed to generate user ID: %v", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Insert admin user
	query := `
		INSERT INTO users (
			id, name, email, password_hash, account_type,
			is_admin, email_verified, is_active,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (email) DO UPDATE SET
			account_type = EXCLUDED.account_type,
			is_admin = EXCLUDED.is_admin,
			password_hash = EXCLUDED.password_hash,
			updated_at = NOW()
		RETURNING id
	`

	err = db.QueryRow(query, userID, *name, *email, string(hashedPassword), *accountType, true, true, true).Scan(&userID)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Println("✅ Admin user created successfully!")
	fmt.Printf("\nUser ID: %d\n", userID)
	fmt.Printf("Name: %s\n", *name)
	fmt.Printf("Email: %s\n", *email)
	fmt.Printf("Account Type: %s\n", *accountType)
	fmt.Printf("\nYou can now login at: http://localhost:3000/admin/login\n")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
