package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type TeamHandlers struct {
	teamService *services.TeamService
	rbacService *services.RBACService
	userService *services.UserService
}

// NewTeamHandlers creates new team handlers
func NewTeamHandlers(teamService *services.TeamService, rbacService *services.RBACService, userService *services.UserService) *TeamHandlers {
	return &TeamHandlers{
		teamService: teamService,
		rbacService: rbacService,
		userService: userService,
	}
}

// Team CRUD Operations

// CreateTeam creates a new team
// POST /api/v1/teams
func (h *TeamHandlers) CreateTeam(c *gin.Context) {
	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	ownerID := userID.(int64)

	team, err := h.teamService.CreateTeam(&req, ownerID)
	if err != nil {
		if err.Error() == "team name already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "team_name_exists",
				Message: "Team name already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "create_team_failed",
			Message: "Failed to create team",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team created successfully",
		"team":    team,
	})
}

// GetTeam retrieves a team by ID
// GET /api/v1/teams/:id
func (h *TeamHandlers) GetTeam(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Check if user has access to this team
	canAccess, err := h.teamService.UserCanAccessTeam(userID.(int64), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "access_check_failed",
			Message: "Failed to check team access",
		})
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "team_access_denied",
			Message: "You don't have access to this team",
		})
		return
	}

	team, err := h.teamService.GetTeamWithMembers(teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "team_not_found",
			Message: "Team not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team": team,
	})
}

// ListUserTeams lists all teams for the current user
// GET /api/v1/teams
func (h *TeamHandlers) ListUserTeams(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	teams, err := h.teamService.GetUserTeams(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_teams_failed",
			Message: "Failed to retrieve teams",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
		"count": len(teams),
	})
}

// UpdateTeam updates an existing team
// PUT /api/v1/teams/:id
func (h *TeamHandlers) UpdateTeam(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	var req models.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Check if user can manage this team
	canManage, err := h.teamService.UserCanManageTeam(userID.(int64), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "permission_check_failed",
			Message: "Failed to check team management permission",
		})
		return
	}

	if !canManage {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "team_management_denied",
			Message: "You don't have permission to manage this team",
		})
		return
	}

	team, err := h.teamService.UpdateTeam(teamID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "team_not_found",
				Message: "Team not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_team_failed",
			Message: "Failed to update team",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team updated successfully",
		"team":    team,
	})
}

// DeleteTeam deletes a team (soft delete)
// DELETE /api/v1/teams/:id
func (h *TeamHandlers) DeleteTeam(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Check if user is the team owner
	isOwner, err := h.teamService.IsTeamOwner(userID.(int64), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "ownership_check_failed",
			Message: "Failed to check team ownership",
		})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "team_owner_required",
			Message: "Only team owner can delete the team",
		})
		return
	}

	err = h.teamService.DeleteTeam(teamID, userID.(int64))
	if err != nil {
		if err.Error() == "team not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "team_not_found",
				Message: "Team not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_team_failed",
			Message: "Failed to delete team",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team deleted successfully",
	})
}

// Team Member Management

// InviteTeamMember invites a user to join a team
// POST /api/v1/teams/:id/invitations
func (h *TeamHandlers) InviteTeamMember(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	var req models.InviteTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Check if user can invite members to this team
	canInvite, err := h.teamService.UserCanInviteToTeam(userID.(int64), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "permission_check_failed",
			Message: "Failed to check invitation permission",
		})
		return
	}

	if !canInvite {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "invite_permission_denied",
			Message: "You don't have permission to invite members to this team",
		})
		return
	}

	invitation, err := h.teamService.InviteTeamMember(teamID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "user already member of team" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "user_already_member",
				Message: "User is already a member of this team",
			})
			return
		}

		if err.Error() == "user already invited to team" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "user_already_invited",
				Message: "User has already been invited to this team",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "invite_failed",
			Message: "Failed to invite team member",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Team member invited successfully",
		"invitation": invitation,
	})
}

// AcceptTeamInvitation accepts a team invitation
// POST /api/v1/teams/invitations/accept
func (h *TeamHandlers) AcceptTeamInvitation(c *gin.Context) {
	var req models.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	teamMember, err := h.teamService.AcceptTeamInvitation(req.Token, userID.(int64))
	if err != nil {
		if err.Error() == "invitation not found or expired" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "invitation_invalid",
				Message: "Invitation not found or has expired",
			})
			return
		}

		if err.Error() == "invitation not for this user" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "invitation_mismatch",
				Message: "This invitation is not for your account",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "accept_invitation_failed",
			Message: "Failed to accept team invitation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Team invitation accepted successfully",
		"team_member": teamMember,
	})
}

// RejectTeamInvitation rejects a team invitation
// POST /api/v1/teams/invitations/reject
func (h *TeamHandlers) RejectTeamInvitation(c *gin.Context) {
	var req models.AcceptInvitationRequest // Reuse the same request structure
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	err := h.teamService.RejectTeamInvitation(req.Token, userID.(int64))
	if err != nil {
		if err.Error() == "invitation not found or expired" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "invitation_invalid",
				Message: "Invitation not found or has expired",
			})
			return
		}

		if err.Error() == "invitation not for this user" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "invitation_mismatch",
				Message: "This invitation is not for your account",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "reject_invitation_failed",
			Message: "Failed to reject team invitation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team invitation rejected successfully",
	})
}

// ListTeamInvitations lists pending invitations for a team
// GET /api/v1/teams/:id/invitations
func (h *TeamHandlers) ListTeamInvitations(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Check if user can manage this team
	canManage, err := h.teamService.UserCanManageTeam(userID.(int64), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "permission_check_failed",
			Message: "Failed to check team management permission",
		})
		return
	}

	if !canManage {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "team_management_denied",
			Message: "You don't have permission to manage this team",
		})
		return
	}

	invitations, err := h.teamService.GetTeamInvitations(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_invitations_failed",
			Message: "Failed to retrieve team invitations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"count":       len(invitations),
	})
}

// UpdateTeamMember updates a team member's role or status
// PUT /api/v1/teams/:id/members/:user_id
func (h *TeamHandlers) UpdateTeamMember(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	memberIDStr := c.Param("user_id")
	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	var req models.UpdateTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Check if user can manage this team
	canManage, err := h.teamService.UserCanManageTeam(userID.(int64), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "permission_check_failed",
			Message: "Failed to check team management permission",
		})
		return
	}

	if !canManage {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "team_management_denied",
			Message: "You don't have permission to manage this team",
		})
		return
	}

	teamMember, err := h.teamService.UpdateTeamMember(teamID, memberID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "team member not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "member_not_found",
				Message: "Team member not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_member_failed",
			Message: "Failed to update team member",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Team member updated successfully",
		"team_member": teamMember,
	})
}

// RemoveTeamMember removes a member from the team
// DELETE /api/v1/teams/:id/members/:user_id
func (h *TeamHandlers) RemoveTeamMember(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	memberIDStr := c.Param("user_id")
	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Users can remove themselves, or managers can remove others
	if userID.(int64) != memberID {
		canManage, err := h.teamService.UserCanManageTeam(userID.(int64), teamID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "permission_check_failed",
				Message: "Failed to check team management permission",
			})
			return
		}

		if !canManage {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "team_management_denied",
				Message: "You don't have permission to remove this team member",
			})
			return
		}
	}

	err = h.teamService.RemoveTeamMember(teamID, memberID, userID.(int64))
	if err != nil {
		if err.Error() == "team member not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "member_not_found",
				Message: "Team member not found",
			})
			return
		}

		if err.Error() == "cannot remove team owner" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "cannot_remove_owner",
				Message: "Cannot remove team owner from team",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "remove_member_failed",
			Message: "Failed to remove team member",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team member removed successfully",
	})
}

// LeaveTeam allows a user to leave a team
// DELETE /api/v1/teams/:id/leave
func (h *TeamHandlers) LeaveTeam(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	err = h.teamService.LeaveTeam(teamID, userID.(int64))
	if err != nil {
		if err.Error() == "user not member of team" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_team_member",
				Message: "You are not a member of this team",
			})
			return
		}

		if err.Error() == "team owner cannot leave team" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "owner_cannot_leave",
				Message: "Team owner cannot leave team. Transfer ownership or delete the team instead.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "leave_team_failed",
			Message: "Failed to leave team",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully left the team",
	})
}

// TransferTeamOwnership transfers team ownership to another member
// POST /api/v1/teams/:id/transfer-ownership
func (h *TeamHandlers) TransferTeamOwnership(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_team_id",
			Message: "Invalid team ID",
		})
		return
	}

	var req struct {
		NewOwnerID int64 `json:"new_owner_id" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Check if current user is the team owner
	isOwner, err := h.teamService.IsTeamOwner(userID.(int64), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "ownership_check_failed",
			Message: "Failed to check team ownership",
		})
		return
	}

	if !isOwner {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "not_team_owner",
			Message: "Only team owner can transfer ownership",
		})
		return
	}

	team, err := h.teamService.TransferTeamOwnership(teamID, req.NewOwnerID, userID.(int64))
	if err != nil {
		if err.Error() == "new owner not member of team" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "new_owner_not_member",
				Message: "New owner must be a member of the team",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "transfer_ownership_failed",
			Message: "Failed to transfer team ownership",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team ownership transferred successfully",
		"team":    team,
	})
}