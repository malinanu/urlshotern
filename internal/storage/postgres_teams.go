package storage

import (
	"database/sql"
	"fmt"

	"github.com/URLshorter/url-shortener/internal/models"
)

// roleNameToID maps a role name string (stored in DB) to a numeric ID used in the model.
func roleNameToID(role string) int64 {
	switch role {
	case models.RoleTeamOwner:
		return 1
	case models.RoleTeamAdmin:
		return 2
	case models.RoleTeamViewer:
		return 4
	default: // team_member
		return 3
	}
}

// roleIDToName maps a numeric role ID back to the role name string for DB storage.
func roleIDToName(id int64) string {
	switch id {
	case 1:
		return models.RoleTeamOwner
	case 2:
		return models.RoleTeamAdmin
	case 4:
		return models.RoleTeamViewer
	default:
		return models.RoleTeamMember
	}
}

// RoleToModel builds a models.Role value from a role name string.
func RoleToModel(role string) *models.Role {
	switch role {
	case models.RoleTeamOwner:
		return &models.Role{ID: 1, Name: models.RoleTeamOwner, DisplayName: "Team Owner", IsSystem: true, IsActive: true}
	case models.RoleTeamAdmin:
		return &models.Role{ID: 2, Name: models.RoleTeamAdmin, DisplayName: "Team Admin", IsSystem: true, IsActive: true}
	case models.RoleTeamViewer:
		return &models.Role{ID: 4, Name: models.RoleTeamViewer, DisplayName: "Team Viewer", IsSystem: true, IsActive: true}
	default:
		return &models.Role{ID: 3, Name: models.RoleTeamMember, DisplayName: "Team Member", IsSystem: true, IsActive: true}
	}
}

// ─── Team CRUD ────────────────────────────────────────────────────────────────

// CreateTeam inserts a new team and populates team.ID from the RETURNING clause.
func (p *PostgresStorage) CreateTeam(team *models.Team) error {
	return p.QueryRow(`
		INSERT INTO teams (name, display_name, description, owner_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, team.Name, team.DisplayName, team.Description, team.OwnerID,
		team.IsActive, team.CreatedAt, team.UpdatedAt,
	).Scan(&team.ID)
}

// GetTeamByID retrieves a non-deleted team by its primary key.
func (p *PostgresStorage) GetTeamByID(id int64) (*models.Team, error) {
	t := &models.Team{}
	err := p.QueryRow(`
		SELECT id, name, display_name, description, owner_id, is_active, created_at, updated_at, deleted_at
		FROM teams WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&t.ID, &t.Name, &t.DisplayName, &t.Description, &t.OwnerID,
		&t.IsActive, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// TeamNameExists reports whether an active team with the given name already exists.
func (p *PostgresStorage) TeamNameExists(name string) (bool, error) {
	var exists bool
	err := p.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1 AND deleted_at IS NULL)`, name,
	).Scan(&exists)
	return exists, err
}

// UpdateTeam saves mutable fields (display_name, description, is_active).
func (p *PostgresStorage) UpdateTeam(team *models.Team) error {
	_, err := p.Exec(`
		UPDATE teams SET display_name = $1, description = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`, team.DisplayName, team.Description, team.IsActive, team.ID)
	return err
}

// SoftDeleteTeam marks a team as deleted without removing the row.
func (p *PostgresStorage) SoftDeleteTeam(id int64) error {
	_, err := p.Exec(`UPDATE teams SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`, id)
	return err
}

// UpdateTeamOwner transfers ownership.
func (p *PostgresStorage) UpdateTeamOwner(teamID, newOwnerID int64) error {
	_, err := p.Exec(`UPDATE teams SET owner_id = $1, updated_at = NOW() WHERE id = $2`, newOwnerID, teamID)
	return err
}

// GetUserTeams returns all active teams a user belongs to, along with their role in each.
func (p *PostgresStorage) GetUserTeams(userID int64) ([]*models.Team, []string, error) {
	rows, err := p.Query(`
		SELECT t.id, t.name, t.display_name, t.description, t.owner_id, t.is_active,
		       t.created_at, t.updated_at, t.deleted_at, tm.role
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id AND tm.user_id = $1 AND tm.status = 'active'
		WHERE t.deleted_at IS NULL AND t.is_active = TRUE
		ORDER BY t.created_at DESC
	`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	var roles []string
	for rows.Next() {
		t := &models.Team{}
		var role string
		if err := rows.Scan(
			&t.ID, &t.Name, &t.DisplayName, &t.Description, &t.OwnerID,
			&t.IsActive, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &role,
		); err != nil {
			return nil, nil, err
		}
		teams = append(teams, t)
		roles = append(roles, role)
	}
	return teams, roles, rows.Err()
}

// ─── Team Members ─────────────────────────────────────────────────────────────

// CreateTeamMember inserts a new team membership row.
func (p *PostgresStorage) CreateTeamMember(member *models.TeamMember) error {
	roleName := roleIDToName(member.RoleID)
	return p.QueryRow(`
		INSERT INTO team_members (team_id, user_id, role, invited_by, joined_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, member.TeamID, member.UserID, roleName, member.InvitedBy,
		member.JoinedAt, member.Status, member.CreatedAt, member.UpdatedAt,
	).Scan(&member.ID)
}

// GetTeamMember retrieves the membership row for a specific user in a team.
func (p *PostgresStorage) GetTeamMember(teamID, userID int64) (*models.TeamMember, error) {
	m := &models.TeamMember{}
	var role string
	err := p.QueryRow(`
		SELECT id, team_id, user_id, role, invited_by, joined_at, status, created_at, updated_at
		FROM team_members WHERE team_id = $1 AND user_id = $2
	`, teamID, userID).Scan(
		&m.ID, &m.TeamID, &m.UserID, &role, &m.InvitedBy,
		&m.JoinedAt, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.RoleID = roleNameToID(role)
	return m, nil
}

// GetTeamMembersWithUsers retrieves active team members joined with their user row.
// Returns (member, publicUser, roleName) slices aligned by index.
func (p *PostgresStorage) GetTeamMembersWithUsers(teamID int64) ([]*models.TeamMember, []*models.PublicUser, []string, error) {
	rows, err := p.Query(`
		SELECT tm.id, tm.team_id, tm.user_id, tm.role, tm.invited_by, tm.joined_at, tm.status, tm.created_at, tm.updated_at,
		       u.id, u.name, u.email, u.avatar_url, u.account_type, u.is_admin, u.is_active, u.created_at
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1 AND tm.status = 'active'
		ORDER BY tm.created_at ASC
	`, teamID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var members []*models.TeamMember
	var users []*models.PublicUser
	var roleNames []string

	for rows.Next() {
		m := &models.TeamMember{}
		pu := &models.PublicUser{}
		var role string
		if err := rows.Scan(
			&m.ID, &m.TeamID, &m.UserID, &role, &m.InvitedBy, &m.JoinedAt, &m.Status, &m.CreatedAt, &m.UpdatedAt,
			&pu.ID, &pu.Name, &pu.Email, &pu.AvatarURL, &pu.AccountType, &pu.IsAdmin, &pu.IsActive, &pu.CreatedAt,
		); err != nil {
			return nil, nil, nil, err
		}
		m.RoleID = roleNameToID(role)
		members = append(members, m)
		users = append(users, pu)
		roleNames = append(roleNames, role)
	}
	return members, users, roleNames, rows.Err()
}

// UpdateTeamMember updates a member's role and status.
func (p *PostgresStorage) UpdateTeamMember(member *models.TeamMember) error {
	roleName := roleIDToName(member.RoleID)
	_, err := p.Exec(`
		UPDATE team_members SET role = $1, status = $2, updated_at = NOW()
		WHERE team_id = $3 AND user_id = $4
	`, roleName, member.Status, member.TeamID, member.UserID)
	return err
}

// DeleteTeamMember removes a user from a team.
func (p *PostgresStorage) DeleteTeamMember(teamID, userID int64) error {
	_, err := p.Exec(`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID)
	return err
}

// ─── Team Invitations ─────────────────────────────────────────────────────────

// CreateTeamInvitation inserts a new invitation and populates inv.ID.
func (p *PostgresStorage) CreateTeamInvitation(inv *models.TeamInvitation) error {
	roleName := roleIDToName(inv.RoleID)
	return p.QueryRow(`
		INSERT INTO team_invitations (team_id, email, role, invited_by, token, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, inv.TeamID, inv.Email, roleName, inv.InvitedBy,
		inv.Token, inv.Status, inv.ExpiresAt, inv.CreatedAt, inv.UpdatedAt,
	).Scan(&inv.ID)
}

// GetTeamInvitationByToken retrieves an invitation by its unique token.
func (p *PostgresStorage) GetTeamInvitationByToken(token string) (*models.TeamInvitation, error) {
	inv := &models.TeamInvitation{}
	var role string
	err := p.QueryRow(`
		SELECT id, team_id, email, role, invited_by, token, status, expires_at, created_at, updated_at
		FROM team_invitations WHERE token = $1
	`, token).Scan(
		&inv.ID, &inv.TeamID, &inv.Email, &role, &inv.InvitedBy,
		&inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	inv.RoleID = roleNameToID(role)
	return inv, nil
}

// GetTeamInvitations retrieves pending, non-expired invitations for a team.
func (p *PostgresStorage) GetTeamInvitations(teamID int64) ([]*models.TeamInvitation, error) {
	rows, err := p.Query(`
		SELECT id, team_id, email, role, invited_by, token, status, expires_at, created_at, updated_at
		FROM team_invitations
		WHERE team_id = $1 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*models.TeamInvitation
	for rows.Next() {
		inv := &models.TeamInvitation{}
		var role string
		if err := rows.Scan(
			&inv.ID, &inv.TeamID, &inv.Email, &role, &inv.InvitedBy,
			&inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		inv.RoleID = roleNameToID(role)
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

// CancelTeamInvitation marks an invitation as cancelled/expired.
func (p *PostgresStorage) CancelTeamInvitation(teamID, invitationID int64) error {
	_, err := p.Exec(
		`UPDATE team_invitations SET status = 'expired', updated_at = NOW() WHERE id = $1 AND team_id = $2`,
		invitationID, teamID,
	)
	return err
}

// UpdateTeamInvitationStatus sets the status of an invitation by its ID.
func (p *PostgresStorage) UpdateTeamInvitationStatus(id int64, status string) error {
	_, err := p.Exec(
		`UPDATE team_invitations SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}

// GetExistingInvitation returns the pending, non-expired invitation for an email+team pair.
func (p *PostgresStorage) GetExistingInvitation(teamID int64, email string) (*models.TeamInvitation, error) {
	inv := &models.TeamInvitation{}
	var role string
	err := p.QueryRow(`
		SELECT id, team_id, email, role, invited_by, token, status, expires_at, created_at, updated_at
		FROM team_invitations
		WHERE team_id = $1 AND email = $2 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1
	`, teamID, email).Scan(
		&inv.ID, &inv.TeamID, &inv.Email, &role, &inv.InvitedBy,
		&inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	inv.RoleID = roleNameToID(role)
	return inv, nil
}

// TeamInvitationExists reports whether a pending, non-expired invitation exists for an email+team pair.
func (p *PostgresStorage) TeamInvitationExists(teamID int64, email string) (bool, error) {
	var exists bool
	err := p.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM team_invitations
			WHERE team_id = $1 AND email = $2 AND status = 'pending' AND expires_at > NOW()
		)
	`, teamID, email).Scan(&exists)
	return exists, err
}

// TeamMemberExists reports whether a user is already an active member of a team.
func (p *PostgresStorage) TeamMemberExists(teamID, userID int64) (bool, error) {
	var exists bool
	err := p.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2 AND status = 'active')
	`, teamID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetTeamLinks returns url_mappings assigned to a team, paginated.
func (p *PostgresStorage) GetTeamLinks(teamID int64, limit, offset int) ([]*models.URLMapping, int64, error) {
	rows, err := p.Query(`
		SELECT id, short_code, original_url, created_at, expires_at, click_count, is_active,
		       created_by_ip, user_id, password_hash, og_title, og_description, og_image, twitter_card,
		       custom_domain_id, team_id
		FROM url_mappings
		WHERE team_id = $1 AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get team links: %w", err)
	}
	defer rows.Close()

	var links []*models.URLMapping
	for rows.Next() {
		m := &models.URLMapping{}
		var expiresAt sql.NullTime
		var createdByIP, passwordHash, ogTitle, ogDescription, ogImage, twitterCard sql.NullString
		var userID, customDomainID, teamIDNull sql.NullInt64
		if err := rows.Scan(
			&m.ID, &m.ShortCode, &m.OriginalURL, &m.CreatedAt, &expiresAt, &m.ClickCount, &m.IsActive,
			&createdByIP, &userID, &passwordHash, &ogTitle, &ogDescription, &ogImage, &twitterCard,
			&customDomainID, &teamIDNull,
		); err != nil {
			return nil, 0, err
		}
		if expiresAt.Valid {
			m.ExpiresAt = &expiresAt.Time
		}
		if createdByIP.Valid {
			m.CreatedByIP = createdByIP.String
		}
		if userID.Valid {
			m.UserID = &userID.Int64
		}
		if customDomainID.Valid {
			m.CustomDomainID = &customDomainID.Int64
		}
		if teamIDNull.Valid {
			m.TeamID = &teamIDNull.Int64
		}
		links = append(links, m)
	}

	var total int64
	if err := p.QueryRow(`SELECT COUNT(*) FROM url_mappings WHERE team_id = $1 AND is_active = TRUE`, teamID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return links, total, nil
}

// GetUserByEmail is exposed here as a thin wrapper so team service can resolve email → userID.
// (The real method lives in postgres.go; this avoids import cycles.)
func (p *PostgresStorage) GetUserIDByEmail(email string) (int64, error) {
	var id int64
	err := p.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, sql.ErrNoRows
	}
	return id, err
}
