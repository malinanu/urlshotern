package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Admin User Management Methods
// ============================================================================

// CreateUserAsAdmin creates a new user with admin-specified roles
func (s *AdminService) CreateUserAsAdmin(req *models.CreateUserAdminRequest, createdBy int64) (*models.CreateUserAdminResponse, error) {
	// 1. Validate role hierarchy - can the admin create users with these roles?
	creatorRoles, err := s.rbacService.GetUserRoles(createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator roles: %w", err)
	}

	if len(creatorRoles) == 0 {
		return nil, fmt.Errorf("creator has no roles")
	}

	// Get the highest privilege role of the creator
	creatorRole := creatorRoles[0].Name
	for _, role := range creatorRoles {
		if role.Name == models.RoleSuperAdmin {
			creatorRole = models.RoleSuperAdmin
			break
		} else if role.Name == models.RoleAdmin && creatorRole != models.RoleSuperAdmin {
			creatorRole = models.RoleAdmin
		}
	}

	// Verify creator can assign all requested roles
	for _, roleID := range req.RoleIDs {
		targetRole, err := s.rbacService.GetRoleByID(roleID)
		if err != nil {
			return nil, fmt.Errorf("role %d not found: %w", roleID, err)
		}

		if !models.CanAssignRole(creatorRole, targetRole.Name) {
			return nil, fmt.Errorf("insufficient privileges to assign role '%s'", targetRole.Name)
		}
	}

	// 2. Check email uniqueness
	var existingUserID int64
	err = s.db.QueryRow("SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL", req.Email).Scan(&existingUserID)
	if err == nil {
		return nil, fmt.Errorf("user with email '%s' already exists", req.Email)
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}

	// 3. Generate Snowflake ID
	userID, _ := utils.GenerateSnowflakeID()

	// 4. Handle password/invitation logic
	var hashedPassword string
	var inviteURL *string
	var tempPassword *string

	if req.SendInvite {
		// Generate invitation token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return nil, fmt.Errorf("failed to generate invitation token: %w", err)
		}
		token := base64.URLEncoding.EncodeToString(tokenBytes)

		// Generate invite URL (you'll need to configure this based on your frontend)
		url := fmt.Sprintf("https://your-domain.com/accept-invite?token=%s", token)
		inviteURL = &url

		// Set a temporary password that will be changed on first login
		tempPass := generateTempPassword()
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(tempPass), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash temporary password: %w", err)
		}
		hashedPassword = string(hashedBytes)
		tempPassword = &tempPass
	} else {
		// Use provided password or generate temporary password
		var password string
		if req.Password != nil && *req.Password != "" {
			password = *req.Password
		} else {
			password = generateTempPassword()
			tempPassword = &password
		}

		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		hashedPassword = string(hashedBytes)
	}

	// 5. Create user record
	query := `
		INSERT INTO users (
			id, name, email, password, phone, account_type,
			is_active, is_email_verified, force_password_reset,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at
	`

	var createdAt time.Time
	err = s.db.QueryRow(
		query,
		userID,
		req.Name,
		req.Email,
		hashedPassword,
		req.Phone,
		req.AccountType,
		req.IsActive,
		req.EmailVerified,
		req.ForcePasswordReset,
	).Scan(&createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 6. Assign roles
	var assignedRoles []*models.Role
	for _, roleID := range req.RoleIDs {
		err = s.rbacService.AssignRoleToUser(userID, roleID, createdBy, nil)
		if err != nil {
			// Rollback: delete the user if role assignment fails
			s.db.Exec("DELETE FROM users WHERE id = $1", userID)
			return nil, fmt.Errorf("failed to assign role %d: %w", roleID, err)
		}

		role, _ := s.rbacService.GetRoleByID(roleID)
		if role != nil {
			assignedRoles = append(assignedRoles, role)
		}
	}

	// 7. Log admin action to RBAC audit log
	err = s.LogRBACAction(
		models.AuditActionUserCreated,
		&createdBy,
		&userID,
		nil,
		&createdBy,
		"",
		"",
	)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log RBAC action: %v\n", err)
	}

	// 8. Build response
	adminUser := &models.AdminUser{
		ID:              userID,
		Name:            req.Name,
		Email:           req.Email,
		IsActive:        req.IsActive,
		IsEmailVerified: req.EmailVerified,
		AccountType:     req.AccountType,
		CreatedAt:       createdAt,
		URLsCount:       0,
		TotalClicks:     0,
	}

	response := &models.CreateUserAdminResponse{
		User:         adminUser,
		Roles:        assignedRoles,
		InviteURL:    inviteURL,
		TempPassword: tempPassword,
	}

	return response, nil
}

// ValidateRoleAssignment validates if assigner can assign target roles
func (s *AdminService) ValidateRoleAssignment(assignerUserID int64, targetRoleIDs []int64) error {
	// Get assigner's roles
	assignerRoles, err := s.rbacService.GetUserRoles(assignerUserID)
	if err != nil {
		return fmt.Errorf("failed to get assigner roles: %w", err)
	}

	if len(assignerRoles) == 0 {
		return fmt.Errorf("assigner has no roles")
	}

	// Get highest privilege role
	assignerRole := assignerRoles[0].Name
	for _, role := range assignerRoles {
		if role.Name == models.RoleSuperAdmin {
			assignerRole = models.RoleSuperAdmin
			break
		} else if role.Name == models.RoleAdmin {
			assignerRole = models.RoleAdmin
		}
	}

	// Check each target role
	for _, roleID := range targetRoleIDs {
		targetRole, err := s.rbacService.GetRoleByID(roleID)
		if err != nil {
			return fmt.Errorf("role %d not found: %w", roleID, err)
		}

		if !models.CanAssignRole(assignerRole, targetRole.Name) {
			return fmt.Errorf("insufficient privileges to assign role '%s'", targetRole.Name)
		}
	}

	return nil
}

// AssignUserRole assigns a role to a user with audit logging
func (s *AdminService) AssignUserRole(userID, roleID, assignedBy int64, expiresAt *time.Time) error {
	// Validate hierarchy
	err := s.ValidateRoleAssignment(assignedBy, []int64{roleID})
	if err != nil {
		return err
	}

	// Assign role
	err = s.rbacService.AssignRoleToUser(userID, roleID, assignedBy, expiresAt)
	if err != nil {
		return err
	}

	// Log action
	return s.LogRBACAction(
		models.AuditActionRoleAssigned,
		&assignedBy,
		&userID,
		&roleID,
		&assignedBy,
		"",
		"",
	)
}

// RemoveUserRole removes a role from a user with audit logging
func (s *AdminService) RemoveUserRole(userID, roleID, removedBy int64) error {
	// Remove role
	err := s.rbacService.RemoveRoleFromUser(userID, roleID, removedBy)
	if err != nil {
		return err
	}

	// Log action
	return s.LogRBACAction(
		models.AuditActionRoleRemoved,
		&removedBy,
		&userID,
		&roleID,
		&removedBy,
		"",
		"",
	)
}

// GetUserRoles retrieves user's roles and computed permissions
func (s *AdminService) GetUserRoles(userID int64) (*models.GetUserRolesResponse, error) {
	roles, err := s.rbacService.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.rbacService.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	return &models.GetUserRolesResponse{
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// UpdateAccountType updates user's account type
func (s *AdminService) UpdateAccountType(userID int64, accountType string, updatedBy int64) error {
	query := "UPDATE users SET account_type = $1, updated_at = NOW() WHERE id = $2"
	result, err := s.db.Exec(query, accountType, userID)
	if err != nil {
		return fmt.Errorf("failed to update account type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Log action
	return s.LogRBACAction(
		models.AuditActionUserUpdated,
		&updatedBy,
		&userID,
		nil,
		&updatedBy,
		"",
		"",
	)
}

// LogRBACAction logs an RBAC audit trail entry
func (s *AdminService) LogRBACAction(
	action string,
	userID, targetUserID, roleID, assignedBy *int64,
	ipAddress, userAgent string,
) error {
	query := `
		INSERT INTO rbac_audit_log (
			user_id, target_user_id, action, role_id, assigned_by,
			ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	var ipAddr, ua interface{}
	if ipAddress != "" {
		ipAddr = ipAddress
	}
	if userAgent != "" {
		ua = userAgent
	}

	_, err := s.db.Exec(query, userID, targetUserID, action, roleID, assignedBy, ipAddr, ua)
	if err != nil {
		return fmt.Errorf("failed to log RBAC action: %w", err)
	}

	return nil
}

// GetRBACAuditLog retrieves audit logs with filters
func (s *AdminService) GetRBACAuditLog(filters *models.RBACAuditFilters) ([]*models.RBACAuditEntry, error) {
	query := `
		SELECT
			ral.id,
			ral.user_id,
			ral.target_user_id,
			ral.action,
			ral.role_id,
			COALESCE(r.name, '') as role_name,
			ral.assigned_by,
			COALESCE(u.name, '') as assigned_by_name,
			ral.details,
			ral.ip_address,
			ral.user_agent,
			ral.created_at
		FROM rbac_audit_log ral
		LEFT JOIN roles r ON ral.role_id = r.id
		LEFT JOIN users u ON ral.assigned_by = u.id
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND ral.user_id = $%d", argPos)
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.TargetUserID != nil {
		query += fmt.Sprintf(" AND ral.target_user_id = $%d", argPos)
		args = append(args, *filters.TargetUserID)
		argPos++
	}

	if filters.Action != nil {
		query += fmt.Sprintf(" AND ral.action = $%d", argPos)
		args = append(args, *filters.Action)
		argPos++
	}

	if filters.RoleID != nil {
		query += fmt.Sprintf(" AND ral.role_id = $%d", argPos)
		args = append(args, *filters.RoleID)
		argPos++
	}

	if filters.AssignedBy != nil {
		query += fmt.Sprintf(" AND ral.assigned_by = $%d", argPos)
		args = append(args, *filters.AssignedBy)
		argPos++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND ral.created_at >= $%d", argPos)
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND ral.created_at <= $%d", argPos)
		args = append(args, *filters.EndDate)
		argPos++
	}

	query += " ORDER BY ral.created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filters.Limit)
		argPos++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filters.Offset)
		argPos++
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit log: %w", err)
	}
	defer rows.Close()

	var entries []*models.RBACAuditEntry
	for rows.Next() {
		entry := &models.RBACAuditEntry{}
		var details sql.NullString
		var ipAddr sql.NullString
		var userAgent sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.TargetUserID,
			&entry.Action,
			&entry.RoleID,
			&entry.RoleName,
			&entry.AssignedBy,
			&entry.AssignedByName,
			&details,
			&ipAddr,
			&userAgent,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit entry: %w", err)
		}

		if details.Valid {
			entry.Details = &details.String
		}
		if ipAddr.Valid {
			entry.IPAddress = &ipAddr.String
		}
		if userAgent.Valid {
			entry.UserAgent = &userAgent.String
		}

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit entries: %w", err)
	}

	return entries, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateTempPassword generates a secure temporary password
func generateTempPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	const passwordLength = 16

	b := make([]byte, passwordLength)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based password
		return fmt.Sprintf("TempPass%d!", time.Now().Unix())
	}

	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	return string(b)
}
