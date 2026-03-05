package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type TeamService struct {
	db           *storage.PostgresStorage
	redis        *storage.RedisStorage
	rbacService  *RBACService
	userService  *UserService
	emailService *EmailService
}

// NewTeamService creates a new team service
func NewTeamService(
	db *storage.PostgresStorage,
	redis *storage.RedisStorage,
	rbacService *RBACService,
	userService *UserService,
	emailService *EmailService,
) *TeamService {
	return &TeamService{
		db:           db,
		redis:        redis,
		rbacService:  rbacService,
		userService:  userService,
		emailService: emailService,
	}
}

// ─── Team CRUD ────────────────────────────────────────────────────────────────

// CreateTeam creates a new team and adds the creator as owner.
func (t *TeamService) CreateTeam(req *models.CreateTeamRequest, ownerID int64) (*models.Team, error) {
	exists, err := t.db.TeamNameExists(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check team name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("team name already exists")
	}

	now := time.Now()
	team := &models.Team{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		OwnerID:     ownerID,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := t.db.CreateTeam(team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Add owner as team member with owner role
	ownerMember := &models.TeamMember{
		TeamID:    team.ID,
		UserID:    ownerID,
		RoleID:    1, // team_owner
		Status:    models.StatusActive,
		JoinedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := t.db.CreateTeamMember(ownerMember); err != nil {
		return nil, fmt.Errorf("failed to add owner as team member: %w", err)
	}

	return team, nil
}

// GetTeamByID retrieves a team by ID.
func (t *TeamService) GetTeamByID(teamID int64) (*models.Team, error) {
	return t.db.GetTeamByID(teamID)
}

// GetTeamWithMembers retrieves a team together with its member list.
func (t *TeamService) GetTeamWithMembers(teamID int64) (*models.TeamWithMembers, error) {
	team, err := t.db.GetTeamByID(teamID)
	if err != nil {
		return nil, err
	}

	owner, err := t.userService.GetUserByID(team.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team owner: %w", err)
	}

	members, users, roleNames, err := t.db.GetTeamMembersWithUsers(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	memberInfos := make([]*models.TeamMemberInfo, len(members))
	for i, m := range members {
		memberInfos[i] = &models.TeamMemberInfo{
			PublicUser: users[i],
			Role:       storage.RoleToModel(roleNames[i]),
			Status:     m.Status,
			JoinedAt:   m.JoinedAt,
		}
	}

	return &models.TeamWithMembers{
		Team:    team,
		Owner:   owner.ToPublic(),
		Members: memberInfos,
	}, nil
}

// GetUserTeams retrieves all teams a user belongs to.
func (t *TeamService) GetUserTeams(userID int64) ([]*models.TeamInfo, error) {
	teams, roles, err := t.db.GetUserTeams(userID)
	if err != nil {
		return nil, err
	}

	infos := make([]*models.TeamInfo, len(teams))
	for i, team := range teams {
		infos[i] = &models.TeamInfo{
			Team:   team,
			Role:   storage.RoleToModel(roles[i]),
			Status: models.StatusActive,
		}
	}
	return infos, nil
}

// UpdateTeam updates an existing team.
func (t *TeamService) UpdateTeam(teamID int64, req *models.UpdateTeamRequest, updatedBy int64) (*models.Team, error) {
	team, err := t.db.GetTeamByID(teamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, err
	}

	if req.DisplayName != nil {
		team.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		team.Description = req.Description
	}
	if req.IsActive != nil {
		team.IsActive = *req.IsActive
	}

	if err := t.db.UpdateTeam(team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}
	return team, nil
}

// DeleteTeam soft-deletes a team.
func (t *TeamService) DeleteTeam(teamID int64, deletedBy int64) error {
	_, err := t.db.GetTeamByID(teamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("team not found")
		}
		return err
	}
	return t.db.SoftDeleteTeam(teamID)
}

// ─── Team Member Management ───────────────────────────────────────────────────

// InviteTeamMember invites an email address to join a team.
// If the email has no account yet the invitation is still created and the email
// is sent so the recipient can sign up and then accept.
func (t *TeamService) InviteTeamMember(teamID int64, req *models.InviteTeamMemberRequest, invitedBy int64) (*models.TeamInvitation, error) {
	// If the email belongs to an existing user, make sure they're not already a member.
	invitedUserID, err := t.db.GetUserIDByEmail(req.Email)
	if err == nil {
		isMember, err := t.db.TeamMemberExists(teamID, invitedUserID)
		if err != nil {
			return nil, err
		}
		if isMember {
			return nil, fmt.Errorf("user already member of team")
		}
	}
	// err == sql.ErrNoRows → no account yet, proceed anyway

	// Normalize role: accept string alias from frontend ("admin","member","viewer")
	if req.RoleID == 0 {
		switch req.Role {
		case "admin", "team_admin":
			req.RoleID = 2
		case "viewer", "team_viewer":
			req.RoleID = 4
		default:
			req.RoleID = 3 // team_member
		}
	}

	// If a pending invitation already exists, resend the email and return it.
	if existing, err := t.db.GetExistingInvitation(teamID, req.Email); err == nil {
		if t.emailService != nil {
			go func() {
				team, _ := t.db.GetTeamByID(teamID)
				inviter, _ := t.userService.GetUserByID(invitedBy)
				if team != nil && inviter != nil {
					t.emailService.SendTeamInvitation(req.Email, team.DisplayName, inviter.Name, existing.Token)
				}
			}()
		}
		return existing, nil
	}

	token, err := generateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	now := time.Now()
	invitation := &models.TeamInvitation{
		TeamID:    teamID,
		Email:     req.Email,
		RoleID:    req.RoleID,
		InvitedBy: invitedBy,
		Token:     token,
		Status:    models.InvitationStatusPending,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := t.db.CreateTeamInvitation(invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Send invitation email (non-blocking)
	if t.emailService != nil {
		go func() {
			team, _ := t.db.GetTeamByID(teamID)
			inviter, _ := t.userService.GetUserByID(invitedBy)
			if team != nil && inviter != nil {
				if err := t.emailService.SendTeamInvitation(req.Email, team.DisplayName, inviter.Name, token); err != nil {
					fmt.Printf("Failed to send team invitation email: %v\n", err)
				}
			}
		}()
	}

	return invitation, nil
}

// AcceptTeamInvitation accepts a pending invitation.
func (t *TeamService) AcceptTeamInvitation(token string, userID int64) (*models.TeamMember, error) {
	inv, err := t.db.GetTeamInvitationByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invitation not found or expired")
	}

	if inv.Status != models.InvitationStatusPending || inv.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invitation not found or expired")
	}

	user, err := t.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if inv.Email != user.Email {
		return nil, fmt.Errorf("invitation not for this user")
	}

	now := time.Now()
	member := &models.TeamMember{
		TeamID:    inv.TeamID,
		UserID:    userID,
		RoleID:    inv.RoleID,
		InvitedBy: &inv.InvitedBy,
		Status:    models.StatusActive,
		JoinedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := t.db.CreateTeamMember(member); err != nil {
		return nil, fmt.Errorf("failed to add team member: %w", err)
	}

	if err := t.db.UpdateTeamInvitationStatus(inv.ID, models.InvitationStatusAccepted); err != nil {
		return nil, fmt.Errorf("failed to update invitation status: %w", err)
	}

	return member, nil
}

// RejectTeamInvitation rejects a pending invitation.
func (t *TeamService) RejectTeamInvitation(token string, userID int64) error {
	inv, err := t.db.GetTeamInvitationByToken(token)
	if err != nil {
		return fmt.Errorf("invitation not found or expired")
	}

	user, err := t.userService.GetUserByID(userID)
	if err != nil {
		return err
	}
	if inv.Email != user.Email {
		return fmt.Errorf("invitation not for this user")
	}

	return t.db.UpdateTeamInvitationStatus(inv.ID, models.InvitationStatusRejected)
}

// GetTeamInvitations returns pending invitations for a team.
func (t *TeamService) GetTeamInvitations(teamID int64) ([]*models.TeamInvitation, error) {
	return t.db.GetTeamInvitations(teamID)
}

// UpdateTeamMember changes a member's role or status.
func (t *TeamService) UpdateTeamMember(teamID, memberID int64, req *models.UpdateTeamMemberRequest, updatedBy int64) (*models.TeamMember, error) {
	member, err := t.db.GetTeamMember(teamID, memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team member not found")
		}
		return nil, err
	}

	if req.RoleID != nil {
		member.RoleID = *req.RoleID
	}
	if req.Status != nil {
		member.Status = *req.Status
	}

	if err := t.db.UpdateTeamMember(member); err != nil {
		return nil, fmt.Errorf("failed to update team member: %w", err)
	}
	return member, nil
}

// RemoveTeamMember removes a user from the team.
func (t *TeamService) RemoveTeamMember(teamID, memberID, removedBy int64) error {
	team, err := t.db.GetTeamByID(teamID)
	if err != nil {
		return err
	}
	if team.OwnerID == memberID {
		return fmt.Errorf("cannot remove team owner")
	}

	exists, err := t.db.TeamMemberExists(teamID, memberID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("team member not found")
	}

	return t.db.DeleteTeamMember(teamID, memberID)
}

// LeaveTeam lets a non-owner member leave a team.
func (t *TeamService) LeaveTeam(teamID, userID int64) error {
	team, err := t.db.GetTeamByID(teamID)
	if err != nil {
		return err
	}
	if team.OwnerID == userID {
		return fmt.Errorf("team owner cannot leave team")
	}

	isMember, err := t.IsTeamMember(userID, teamID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user not member of team")
	}

	return t.db.DeleteTeamMember(teamID, userID)
}

// TransferTeamOwnership transfers ownership to an existing team member.
func (t *TeamService) TransferTeamOwnership(teamID, newOwnerID, currentOwnerID int64) (*models.Team, error) {
	isMember, err := t.IsTeamMember(newOwnerID, teamID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("new owner not member of team")
	}

	if err := t.db.UpdateTeamOwner(teamID, newOwnerID); err != nil {
		return nil, fmt.Errorf("failed to transfer ownership: %w", err)
	}

	// Update roles: new owner gets owner role, old owner becomes admin
	newOwnerMember, err := t.db.GetTeamMember(teamID, newOwnerID)
	if err == nil {
		newOwnerMember.RoleID = 1 // team_owner
		t.db.UpdateTeamMember(newOwnerMember)
	}
	oldOwnerMember, err := t.db.GetTeamMember(teamID, currentOwnerID)
	if err == nil {
		oldOwnerMember.RoleID = 2 // team_admin
		t.db.UpdateTeamMember(oldOwnerMember)
	}

	return t.db.GetTeamByID(teamID)
}

// ─── Permission and Access Control ───────────────────────────────────────────

// UserCanAccessTeam checks if a user can view a team (must be a member).
func (t *TeamService) UserCanAccessTeam(userID, teamID int64) (bool, error) {
	return t.IsTeamMember(userID, teamID)
}

// UserCanManageTeam checks if a user can manage a team (owner or admin).
func (t *TeamService) UserCanManageTeam(userID, teamID int64) (bool, error) {
	isOwner, err := t.IsTeamOwner(userID, teamID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}
	return t.IsTeamAdmin(userID, teamID)
}

// UserCanInviteToTeam checks if a user can invite others (owner or admin).
func (t *TeamService) UserCanInviteToTeam(userID, teamID int64) (bool, error) {
	return t.UserCanManageTeam(userID, teamID)
}

// IsTeamOwner checks if the user is the team's owner.
func (t *TeamService) IsTeamOwner(userID, teamID int64) (bool, error) {
	team, err := t.db.GetTeamByID(teamID)
	if err != nil {
		return false, err
	}
	return team.OwnerID == userID, nil
}

// IsTeamAdmin checks if the user has the team_admin role in the team.
func (t *TeamService) IsTeamAdmin(userID, teamID int64) (bool, error) {
	member, err := t.db.GetTeamMember(teamID, userID)
	if err != nil {
		return false, nil // not a member
	}
	return member.RoleID == 2, nil // 2 = team_admin
}

// IsTeamMember checks if the user is an active member of the team.
func (t *TeamService) IsTeamMember(userID, teamID int64) (bool, error) {
	return t.db.TeamMemberExists(teamID, userID)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func generateInvitationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
