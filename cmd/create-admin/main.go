package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  URL Shortener - Create Super Admin")
	fmt.Println("========================================")
	fmt.Println()

	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database manager
	dbManager, err := storage.NewDatabaseManager(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbManager.Close()

	// Initialize Snowflake ID generator
	snowflake, err := utils.NewSnowflakeGenerator(config.NodeID)
	if err != nil {
		log.Fatalf("Failed to initialize Snowflake: %v", err)
	}

	// Initialize audit service
	auditService := services.NewAdminAuditService(dbManager, snowflake)

	// Prompt for admin details
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter super admin details:")
	fmt.Println()

	// Get email
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if email == "" {
		log.Fatal("Email is required")
	}

	// Check if user already exists
	var existingUserID int64
	err = dbManager.AppDB.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&existingUserID)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("Failed to check existing user: %v", err)
	}

	if err != sql.ErrNoRows {
		fmt.Printf("\n⚠️  User with email %s already exists (ID: %d)\n", email, existingUserID)
		fmt.Print("Do you want to promote this user to super admin? (yes/no): ")
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm != "yes" && confirm != "y" {
			fmt.Println("Aborted.")
			return
		}

		// Promote existing user
		if err := promoteToSuperAdmin(dbManager, auditService, existingUserID, email); err != nil {
			log.Fatalf("Failed to promote user: %v", err)
		}

		fmt.Println("\n✅ User promoted to super admin successfully!")
		return
	}

	// Get name
	fmt.Print("Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		name = "Super Administrator"
	}

	// Get password (hidden input)
	fmt.Print("Password (min 8 characters): ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println()

	if len(password) < 8 {
		log.Fatal("Password must be at least 8 characters long")
	}

	// Confirm password
	fmt.Print("Confirm password: ")
	confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password confirmation: %v", err)
	}
	confirmPassword := string(confirmPasswordBytes)
	fmt.Println()

	if password != confirmPassword {
		log.Fatal("Passwords do not match")
	}

	// Generate password hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Generate Snowflake ID
	userID, err := snowflake.Generate()
	if err != nil {
		log.Fatalf("Failed to generate user ID: %v", err)
	}

	// Create super admin user
	fmt.Println("\n Creating super admin user...")

	query := `
		INSERT INTO users (
			id, email, name, password_hash, email_verified,
			account_type, is_active, is_admin, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = dbManager.AdminDB.Exec(
		query,
		userID,
		email,
		name,
		string(passwordHash),
		true,                // email_verified
		"super_admin",       // account_type
		true,                // is_active
		true,                // is_admin
		time.Now(),
		time.Now(),
	)

	if err != nil {
		log.Fatalf("Failed to create super admin user: %v", err)
	}

	fmt.Printf("✓ Created user (ID: %d)\n", userID)

	// Assign super_admin role
	fmt.Println("Assigning super_admin role...")

	var roleID int64
	err = dbManager.AdminDB.QueryRow("SELECT id FROM roles WHERE name = 'super_admin'").Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("super_admin role not found. Please run migrations first.")
		}
		log.Fatalf("Failed to find super_admin role: %v", err)
	}

	_, err = dbManager.AdminDB.Exec(
		"INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at) VALUES ($1, $2, $3, $4)",
		userID,
		roleID,
		userID, // self-assigned
		time.Now(),
	)

	if err != nil {
		log.Fatalf("Failed to assign super_admin role: %v", err)
	}

	fmt.Println("✓ Assigned super_admin role")

	// Log the creation in audit logs
	fmt.Println("Logging admin creation...")

	auditReq := models.CreateAuditLogRequest{
		UserID:       &userID,
		Username:     email,
		Action:       models.AuditActionUserCreate,
		ResourceType: models.ResourceTypeUser,
		ResourceID:   fmt.Sprintf("%d", userID),
		Details: map[string]interface{}{
			"method":       "cli_script",
			"type":         "super_admin_creation",
			"name":         name,
			"email":        email,
			"account_type": "super_admin",
		},
		IPAddress: "127.0.0.1",
		UserAgent: "create-admin-cli",
		Status:    models.AuditStatusSuccess,
		Severity:  models.AuditSeverityCritical,
	}

	if err := auditService.LogAction(context.Background(), auditReq); err != nil {
		log.Printf("Warning: Failed to log admin creation in audit log: %v", err)
	} else {
		fmt.Println("✓ Logged in audit trail")
	}

	// Display success message
	fmt.Println("\n========================================")
	fmt.Println("✅ Super Admin Created Successfully!")
	fmt.Println("========================================")
	fmt.Printf("\nCredentials:\n")
	fmt.Printf("  Email:    %s\n", email)
	fmt.Printf("  Password: <hidden>\n")
	fmt.Printf("  User ID:  %d\n", userID)
	fmt.Printf("  Role:     super_admin\n")
	fmt.Println("\n⚠️  IMPORTANT SECURITY NOTES:")
	fmt.Println("  1. Keep these credentials secure")
	fmt.Println("  2. Enable MFA after first login (if available)")
	fmt.Println("  3. Review audit logs regularly")
	fmt.Println("  4. Never share admin credentials")
	fmt.Println()
}

// promoteToSuperAdmin promotes an existing user to super admin
func promoteToSuperAdmin(
	dbManager *storage.DatabaseManager,
	auditService *services.AdminAuditService,
	userID int64,
	email string,
) error {
	fmt.Println("\nPromoting user to super admin...")

	// Update user account type
	_, err := dbManager.AdminDB.Exec(`
		UPDATE users
		SET account_type = 'super_admin',
		    is_admin = TRUE,
		    updated_at = $1
		WHERE id = $2
	`, time.Now(), userID)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	fmt.Println("✓ Updated account type")

	// Assign super_admin role if not already assigned
	var roleID int64
	err = dbManager.AdminDB.QueryRow("SELECT id FROM roles WHERE name = 'super_admin'").Scan(&roleID)
	if err != nil {
		return fmt.Errorf("failed to find super_admin role: %w", err)
	}

	_, err = dbManager.AdminDB.Exec(`
		INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID, userID, time.Now())

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	fmt.Println("✓ Assigned super_admin role")

	// Log the promotion
	auditReq := models.CreateAuditLogRequest{
		UserID:       &userID,
		Username:     email,
		Action:       "user.promote_to_super_admin",
		ResourceType: models.ResourceTypeUser,
		ResourceID:   fmt.Sprintf("%d", userID),
		Details: map[string]interface{}{
			"method":  "cli_script",
			"promoted_to": "super_admin",
		},
		IPAddress: "127.0.0.1",
		UserAgent: "create-admin-cli",
		Status:    models.AuditStatusSuccess,
		Severity:  models.AuditSeverityCritical,
	}

	if err := auditService.LogAction(context.Background(), auditReq); err != nil {
		log.Printf("Warning: Failed to log promotion in audit log: %v", err)
	} else {
		fmt.Println("✓ Logged in audit trail")
	}

	return nil
}
