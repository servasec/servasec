package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetTeams(c *gin.Context) {
	var teams []models.Team
	if err := config.DB.Find(&teams).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch teams")
		return
	}
	utils.OKResponse(c, teams)
}

func GetTeam(c *gin.Context) {
	var team models.Team
	if err := config.DB.First(&team, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Team not found")
		return
	}
	utils.OKResponse(c, team)
}

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

	utils.CreatedResponse(c, team)
}

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

	c.JSON(http.StatusCreated, gin.H{"message": "Member added"})
}

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
	utils.OKResponse(c, gin.H{"message": "Member removed"})
}
