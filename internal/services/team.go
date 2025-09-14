package services

import (
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
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

// Team CRUD Operations

// CreateTeam creates a new team
func (t *TeamService) CreateTeam(req *models.CreateTeamRequest, ownerID int64) (*models.Team, error) {
	// Validate team name uniqueness (simplified)
	// In a real implementation, this would query the database
	
	team := &models.Team{
		ID:          time.Now().Unix(), // Temporary ID generation
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		OwnerID:     ownerID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Implementation would save to database
	// For now, simulate success

	// Add owner as team member with owner role
	ownerRoleID, err := t.getTeamOwnerRoleID()
	if err != nil {
		return nil, fmt.Errorf("failed to get owner role: %w", err)
	}

	teamMember := &models.TeamMember{
		ID:        time.Now().Unix() + 1, // Avoid ID collision
		TeamID:    team.ID,
		UserID:    ownerID,
		RoleID:    ownerRoleID,
		Status:    models.StatusActive,
		JoinedAt:  &team.CreatedAt,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
	}

	// Implementation would save team member to database
	_ = teamMember // Use the teamMember

	return team, nil
}

// GetTeamByID retrieves a team by ID
func (t *TeamService) GetTeamByID(teamID int64) (*models.Team, error) {
	// Implementation would query database
	// For now, return simulated team
	return &models.Team{
		ID:          teamID,
		Name:        "sample-team",
		DisplayName: "Sample Team",
		Description: utils.StringPtr("A sample team for demonstration"),
		OwnerID:     1,
		IsActive:    true,
		CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
	}, nil
}

// GetTeamWithMembers retrieves a team with its members
func (t *TeamService) GetTeamWithMembers(teamID int64) (*models.TeamWithMembers, error) {
	team, err := t.GetTeamByID(teamID)
	if err != nil {
		return nil, err
	}

	// Get team owner
	owner, err := t.userService.GetUserByID(team.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team owner: %w", err)
	}

	// Get team members (simplified)
	members := []*models.TeamMemberInfo{
		{
			PublicUser: owner.ToPublic(),
			Role: &models.Role{
				ID:          1,
				Name:        models.RoleTeamOwner,
				DisplayName: "Team Owner",
			},
			Status:   models.StatusActive,
			JoinedAt: &team.CreatedAt,
		},
	}

	return &models.TeamWithMembers{
		Team:    team,
		Owner:   owner.ToPublic(),
		Members: members,
	}, nil
}

// GetUserTeams retrieves all teams for a user
func (t *TeamService) GetUserTeams(userID int64) ([]*models.TeamInfo, error) {
	// Implementation would query database
	// For now, return simulated teams
	teams := []*models.TeamInfo{
		{
			Team: &models.Team{
				ID:          1,
				Name:        "my-team",
				DisplayName: "My Team",
				Description: utils.StringPtr("My personal team"),
				OwnerID:     userID,
				IsActive:    true,
				CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
			},
			Role: &models.Role{
				ID:          1,
				Name:        models.RoleTeamOwner,
				DisplayName: "Team Owner",
			},
			Status: models.StatusActive,
		},
	}

	return teams, nil
}

// UpdateTeam updates an existing team
func (t *TeamService) UpdateTeam(teamID int64, req *models.UpdateTeamRequest, updatedBy int64) (*models.Team, error) {
	team, err := t.GetTeamByID(teamID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		team.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		team.Description = req.Description
	}
	if req.IsActive != nil {
		team.IsActive = *req.IsActive
	}

	team.UpdatedAt = time.Now()

	// Implementation would save to database
	return team, nil
}

// DeleteTeam soft deletes a team
func (t *TeamService) DeleteTeam(teamID int64, deletedBy int64) error {
	team, err := t.GetTeamByID(teamID)
	if err != nil {
		return err
	}

	now := time.Now()
	team.DeletedAt = &now
	team.UpdatedAt = now

	// Implementation would save to database
	return nil
}

// Team Member Management

// InviteTeamMember invites a user to join a team
func (t *TeamService) InviteTeamMember(teamID int64, req *models.InviteTeamMemberRequest, invitedBy int64) (*models.TeamInvitation, error) {
	// Check if user exists
	_, err := t.userService.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is already a member
	isMember, err := t.IsTeamMember(0, teamID) // We'd need to get user ID from email
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, fmt.Errorf("user already member of team")
	}

	// Generate invitation token
	token := t.generateInvitationToken()
	
	invitation := &models.TeamInvitation{
		ID:        time.Now().Unix(),
		TeamID:    teamID,
		Email:     req.Email,
		RoleID:    req.RoleID,
		InvitedBy: invitedBy,
		Token:     token,
		Status:    models.InvitationStatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Implementation would save to database

	// Send invitation email
	if t.emailService != nil {
		go func() {
			team, _ := t.GetTeamByID(teamID)
			inviter, _ := t.userService.GetUserByID(invitedBy)
			
			err := t.emailService.SendTeamInvitation(
				req.Email,
				team.DisplayName,
				inviter.Name,
				token,
			)
			if err != nil {
				// Log error but don't fail the invitation
				fmt.Printf("Failed to send team invitation email: %v\n", err)
			}
		}()
	}

	return invitation, nil
}

// AcceptTeamInvitation accepts a team invitation
func (t *TeamService) AcceptTeamInvitation(token string, userID int64) (*models.TeamMember, error) {
	// Get invitation by token
	invitation, err := t.getInvitationByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invitation not found or expired")
	}

	// Verify invitation is for this user's email
	user, err := t.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if invitation.Email != user.Email {
		return nil, fmt.Errorf("invitation not for this user")
	}

	// Check if invitation is still valid
	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invitation not found or expired")
	}

	// Create team member
	teamMember := &models.TeamMember{
		ID:        time.Now().Unix(),
		TeamID:    invitation.TeamID,
		UserID:    userID,
		RoleID:    invitation.RoleID,
		InvitedBy: &invitation.InvitedBy,
		Status:    models.StatusActive,
		JoinedAt:  utils.TimePtr(time.Now()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Update invitation status
	invitation.Status = models.InvitationStatusAccepted
	invitation.UpdatedAt = time.Now()

	// Implementation would save both to database
	return teamMember, nil
}

// RejectTeamInvitation rejects a team invitation
func (t *TeamService) RejectTeamInvitation(token string, userID int64) error {
	// Get invitation by token
	invitation, err := t.getInvitationByToken(token)
	if err != nil {
		return fmt.Errorf("invitation not found or expired")
	}

	// Verify invitation is for this user's email
	user, err := t.userService.GetUserByID(userID)
	if err != nil {
		return err
	}

	if invitation.Email != user.Email {
		return fmt.Errorf("invitation not for this user")
	}

	// Update invitation status
	invitation.Status = models.InvitationStatusRejected
	invitation.UpdatedAt = time.Now()

	// Implementation would save to database
	return nil
}

// GetTeamInvitations retrieves pending invitations for a team
func (t *TeamService) GetTeamInvitations(teamID int64) ([]*models.TeamInvitation, error) {
	// Implementation would query database
	// For now, return empty slice
	return []*models.TeamInvitation{}, nil
}

// UpdateTeamMember updates a team member's role or status
func (t *TeamService) UpdateTeamMember(teamID, memberID int64, req *models.UpdateTeamMemberRequest, updatedBy int64) (*models.TeamMember, error) {
	// Get current team member
	teamMember, err := t.getTeamMember(teamID, memberID)
	if err != nil {
		return nil, fmt.Errorf("team member not found")
	}

	// Update fields if provided
	if req.RoleID != nil {
		teamMember.RoleID = *req.RoleID
	}
	if req.Status != nil {
		teamMember.Status = *req.Status
	}

	teamMember.UpdatedAt = time.Now()

	// Implementation would save to database
	return teamMember, nil
}

// RemoveTeamMember removes a member from the team
func (t *TeamService) RemoveTeamMember(teamID, memberID, removedBy int64) error {
	// Check if member is team owner
	team, err := t.GetTeamByID(teamID)
	if err != nil {
		return err
	}

	if team.OwnerID == memberID {
		return fmt.Errorf("cannot remove team owner")
	}

	// Check if member exists
	_, err = t.getTeamMember(teamID, memberID)
	if err != nil {
		return fmt.Errorf("team member not found")
	}

	// Implementation would delete from database
	return nil
}

// LeaveTeam allows a user to leave a team
func (t *TeamService) LeaveTeam(teamID, userID int64) error {
	// Check if user is team owner
	team, err := t.GetTeamByID(teamID)
	if err != nil {
		return err
	}

	if team.OwnerID == userID {
		return fmt.Errorf("team owner cannot leave team")
	}

	// Check if user is a member
	isMember, err := t.IsTeamMember(userID, teamID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user not member of team")
	}

	// Implementation would remove user from team
	return nil
}

// TransferTeamOwnership transfers ownership to another team member
func (t *TeamService) TransferTeamOwnership(teamID, newOwnerID, currentOwnerID int64) (*models.Team, error) {
	// Check if new owner is a team member
	isMember, err := t.IsTeamMember(newOwnerID, teamID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("new owner not member of team")
	}

	// Update team ownership
	team, err := t.GetTeamByID(teamID)
	if err != nil {
		return nil, err
	}

	team.OwnerID = newOwnerID
	team.UpdatedAt = time.Now()

	// Update team member roles
	// Current owner becomes team admin
	// New owner becomes team owner
	
	// Implementation would update database

	return team, nil
}

// Permission and Access Control Methods

// UserCanAccessTeam checks if a user can access a team
func (t *TeamService) UserCanAccessTeam(userID, teamID int64) (bool, error) {
	// Check if user is a team member
	isMember, err := t.IsTeamMember(userID, teamID)
	if err != nil {
		return false, err
	}
	
	if isMember {
		return true, nil
	}

	// Check if user has system permissions to access all teams
	hasPermission, err := t.rbacService.UserHasPermission(userID, models.PermissionTeamsRead)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// UserCanManageTeam checks if a user can manage a team
func (t *TeamService) UserCanManageTeam(userID, teamID int64) (bool, error) {
	// Check if user is team owner
	isOwner, err := t.IsTeamOwner(userID, teamID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// Check if user is team admin
	isAdmin, err := t.IsTeamAdmin(userID, teamID)
	if err != nil {
		return false, err
	}
	if isAdmin {
		return true, nil
	}

	// Check if user has system permissions to manage all teams
	hasPermission, err := t.rbacService.UserHasPermission(userID, models.PermissionTeamsManage)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// UserCanInviteToTeam checks if a user can invite members to a team
func (t *TeamService) UserCanInviteToTeam(userID, teamID int64) (bool, error) {
	// Check if user can manage team (owners and admins can invite)
	canManage, err := t.UserCanManageTeam(userID, teamID)
	if err != nil {
		return false, err
	}
	if canManage {
		return true, nil
	}

	// Check if user has invite permission
	hasPermission, err := t.rbacService.UserHasPermission(userID, models.PermissionTeamsInvite)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// IsTeamOwner checks if a user is the owner of a team
func (t *TeamService) IsTeamOwner(userID, teamID int64) (bool, error) {
	team, err := t.GetTeamByID(teamID)
	if err != nil {
		return false, err
	}

	return team.OwnerID == userID, nil
}

// IsTeamAdmin checks if a user is an admin of a team
func (t *TeamService) IsTeamAdmin(userID, teamID int64) (bool, error) {
	// Get user's role in team
	teamMember, err := t.getTeamMember(teamID, userID)
	if err != nil {
		return false, nil // User is not a team member
	}

	// Check if role is team admin
	role, err := t.rbacService.GetRoleByID(teamMember.RoleID)
	if err != nil {
		return false, err
	}

	return role.Name == models.RoleTeamAdmin, nil
}

// IsTeamMember checks if a user is a member of a team
func (t *TeamService) IsTeamMember(userID, teamID int64) (bool, error) {
	// Implementation would query database
	// For now, simulate membership check
	if userID == 1 && teamID == 1 {
		return true, nil
	}
	return false, nil
}

// Helper Methods

func (t *TeamService) getInvitationByToken(token string) (*models.TeamInvitation, error) {
	// Implementation would query database by token
	// For now, return simulated invitation
	return &models.TeamInvitation{
		ID:        1,
		TeamID:    1,
		Email:     "user@example.com",
		RoleID:    3, // Team member role
		InvitedBy: 1,
		Token:     token,
		Status:    models.InvitationStatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}, nil
}

func (t *TeamService) getTeamMember(teamID, userID int64) (*models.TeamMember, error) {
	// Implementation would query database
	// For now, return simulated team member
	return &models.TeamMember{
		ID:        1,
		TeamID:    teamID,
		UserID:    userID,
		RoleID:    3, // Team member role
		Status:    models.StatusActive,
		JoinedAt:  utils.TimePtr(time.Now().Add(-30 * 24 * time.Hour)),
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * 24 * time.Hour),
	}, nil
}

func (t *TeamService) getTeamOwnerRoleID() (int64, error) {
	// Get team owner role ID
	role, err := t.rbacService.GetRoleByName(models.RoleTeamOwner)
	if err != nil {
		return 0, err
	}
	return role.ID, nil
}

func (t *TeamService) generateInvitationToken() string {
	// Generate a secure random token for invitations
	// In real implementation, this would use crypto/rand
	return "invitation_token_" + fmt.Sprintf("%d", time.Now().Unix())
}

// Helper functions for pointers
