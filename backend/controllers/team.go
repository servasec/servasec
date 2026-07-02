package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

// GetTeams returns teams accessible to the current user
// @Summary List teams
// @Tags Teams
// @Produce json
// @Success 200 {array} models.Team "List of teams"
// @Router /teams [get]
func GetTeams(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(*models.User)

	query := config.DB.
		Model(&models.Team{}).
		Select("teams.*, (SELECT COUNT(*) FROM team_members WHERE team_members.team_id = teams.id) as member_count")
	if currentUser.Role != "admin" {
		userID, _ := c.Get("user_id")
		query = query.Joins("JOIN team_members ON team_members.team_id = teams.id").
			Where("team_members.user_id = ?", userID)
	}

	var teams []models.Team
	if err := query.Find(&teams).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch teams")
		return
	}
	utils.OKResponse(c, teams)
}

// GetTeam returns a single team by ID with its members
// @Summary Get team
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} object "Team with members"
// @Failure 404 {object} gin.H "Team not found"
// @Router /teams/{id} [get]
func GetTeam(c *gin.Context) {
	var team models.Team
	if err := config.DB.Preload("Members.User").First(&team, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Team not found")
		return
	}

	type memberInfo struct {
		UserID   uint   `json:"userId"`
		Role     string `json:"role"`
		UserName string `json:"userName"`
	}
	var members []memberInfo
	for _, m := range team.Members {
		members = append(members, memberInfo{
			UserID:   m.UserID,
			Role:     m.Role,
			UserName: m.User.Username,
		})
	}

	type teamResponse struct {
		models.Team
		Members []memberInfo `json:"members"`
	}

	utils.OKResponse(c, teamResponse{
		Team:    team,
		Members: members,
	})
}

// CreateTeam creates a new team and adds the creator as admin
// @Summary Create team
// @Tags Teams
// @Accept json
// @Produce json
// @Param input body object true "Team details"
// @Success 201 {object} models.Team "Created team"
// @Failure 400 {object} gin.H "Invalid input"
// @Router /teams [post]
func CreateTeam(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required,max=100"`
		Description string `json:"description" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	team := models.Team{
		Name:        input.Name,
		Description: input.Description,
	}
	if err := config.DB.Create(&team).Error; err != nil {
		utils.InternalServerError(c, "failed to create team")
		return
	}

	userID, _ := c.Get("user_id")
	config.DB.Create(&models.TeamMember{
		TeamID: team.ID,
		UserID: userID.(uint),
		Role:   "admin",
	})

	sub := fmt.Sprintf("user:%v", userID)
	config.CEF.AddPolicy(sub, fmt.Sprintf("/teams/%d", team.ID), "*")
	config.CEF.AddPolicy(fmt.Sprintf("team:%d", team.ID), fmt.Sprintf("/teams/%d", team.ID), "read")
	config.CEF.SavePolicy()

	utils.CreatedResponse(c, team)
}

// UpdateTeam updates an existing team
// @Summary Update team
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param input body object true "Fields to update"
// @Success 200 {object} models.Team "Updated team"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 404 {object} gin.H "Team not found"
// @Router /teams/{id} [put]
func UpdateTeam(c *gin.Context) {
	var team models.Team
	if err := config.DB.First(&team, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Team not found")
		return
	}

	var input struct {
		Name        *string `json:"name" binding:"omitempty,max=100"`
		Description *string `json:"description" binding:"omitempty,max=500"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Name != nil {
		team.Name = *input.Name
	}
	if input.Description != nil {
		team.Description = *input.Description
	}

	if err := config.DB.Save(&team).Error; err != nil {
		utils.InternalServerError(c, "failed to update team")
		return
	}
	utils.OKResponse(c, team)
}

// DeleteTeam deletes a team by ID
// @Summary Delete team
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} gin.H "Team deleted"
// @Failure 404 {object} gin.H "Team not found"
// @Router /teams/{id} [delete]
func DeleteTeam(c *gin.Context) {
	var team models.Team
	if err := config.DB.First(&team, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Team not found")
		return
	}

	if err := config.DB.Delete(&team).Error; err != nil {
		utils.InternalServerError(c, "failed to delete team")
		return
	}
	utils.OKResponse(c, gin.H{"message": "Team deleted"})
}

// AddTeamMember adds a user to a team
// @Summary Add team member
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param input body object true "User ID and role"
// @Success 201 {object} gin.H "Member added"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 404 {object} gin.H "User not found"
// @Failure 409 {object} gin.H "User is already a team member"
// @Router /teams/{id}/members [post]
func AddTeamMember(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestError(c, "Invalid team ID")
		return
	}

	var input struct {
		UserID uint   `json:"userId" binding:"required"`
		Role   string `json:"role" binding:"omitempty,oneof=admin member"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Role == "" {
		input.Role = "member"
	}

	var user models.User
	if err := config.DB.First(&user, input.UserID).Error; err != nil {
		utils.NotFoundError(c, "User not found")
		return
	}

	member := models.TeamMember{
		TeamID: uint(teamID),
		UserID: input.UserID,
		Role:   input.Role,
	}
	if err := config.DB.Create(&member).Error; err != nil {
		utils.ConflictError(c, "User is already a team member")
		return
	}

	if input.Role == "admin" {
		sub := fmt.Sprintf("user:%d", input.UserID)
		config.CEF.AddPolicy(sub, fmt.Sprintf("/teams/%d", teamID), "*")
		config.CEF.SavePolicy()
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added"})
}

// RemoveTeamMember removes a user from a team
// @Summary Remove team member
// @Tags Teams
// @Produce json
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Success 200 {object} gin.H "Member removed"
// @Failure 400 {object} gin.H "Invalid ID"
// @Failure 404 {object} gin.H "Team member not found"
// @Router /teams/{id}/members/{userId} [delete]
func RemoveTeamMember(c *gin.Context) {
	teamID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestError(c, "Invalid team ID")
		return
	}
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		utils.BadRequestError(c, "Invalid user ID")
		return
	}

	result := config.DB.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&models.TeamMember{})
	if result.Error != nil {
		utils.InternalServerError(c, "failed to remove team member")
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFoundError(c, "Team member not found")
		return
	}

	sub := fmt.Sprintf("user:%d", userID)
	config.CEF.RemovePolicy(sub, fmt.Sprintf("/teams/%d", teamID), "*")
	config.CEF.SavePolicy()

	utils.OKResponse(c, gin.H{"message": "Member removed"})
}


