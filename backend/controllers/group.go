package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

// GetGroups returns all groups
// @Summary List groups
// @Tags Groups
// @Produce json
// @Success 200 {array} models.Group "List of groups"
// @Router /groups [get]
func GetGroups(c *gin.Context) {
	var groups []models.Group
	if err := config.DB.Find(&groups).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch groups")
		return
	}
	utils.OKResponse(c, groups)
}

// GetGroup returns a single group by ID
// @Summary Get group by ID
// @Tags Groups
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} models.Group "Group details"
// @Failure 404 {object} gin.H "Group not found"
// @Router /groups/{id} [get]
func GetGroup(c *gin.Context) {
	var group models.Group
	if err := config.DB.First(&group, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Group not found")
		return
	}
	utils.OKResponse(c, group)
}

// CreateGroup creates a new group
// @Summary Create group
// @Tags Groups
// @Accept json
// @Produce json
// @Param input body object true "Group details"
// @Success 201 {object} models.Group "Created group"
// @Failure 400 {object} gin.H "Invalid input"
// @Router /groups [post]
func CreateGroup(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required,max=100"`
		Description string `json:"description" binding:"max=500"`
		Path        string `json:"path" binding:"required,max=100"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	group := models.Group{
		Name:        input.Name,
		Description: input.Description,
		Path:        input.Path,
	}
	if err := config.DB.Create(&group).Error; err != nil {
		utils.InternalServerError(c, "failed to create group")
		return
	}

	userID, _ := c.Get("user_id")
	sub := fmt.Sprintf("user:%v", userID)
	config.CEF.AddPolicy(sub, fmt.Sprintf("/groups/%d", group.ID), "*")
	config.CEF.SavePolicy()

	utils.CreatedResponse(c, group)
}

// UpdateGroup updates an existing group
// @Summary Update group
// @Tags Groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param input body object true "Fields to update"
// @Success 200 {object} models.Group "Updated group"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 404 {object} gin.H "Group not found"
// @Router /groups/{id} [put]
func UpdateGroup(c *gin.Context) {
	var group models.Group
	if err := config.DB.First(&group, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Group not found")
		return
	}

	var input struct {
		Name        *string `json:"name" binding:"omitempty,max=100"`
		Description *string `json:"description" binding:"omitempty,max=500"`
		Path        *string `json:"path" binding:"omitempty,max=100"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Name != nil {
		group.Name = *input.Name
	}
	if input.Description != nil {
		group.Description = *input.Description
	}
	if input.Path != nil {
		group.Path = *input.Path
	}

	if err := config.DB.Save(&group).Error; err != nil {
		utils.InternalServerError(c, "failed to update group")
		return
	}
	utils.OKResponse(c, group)
}

// DeleteGroup deletes a group by ID
// @Summary Delete group
// @Tags Groups
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} gin.H "Group deleted"
// @Failure 404 {object} gin.H "Group not found"
// @Router /groups/{id} [delete]
func DeleteGroup(c *gin.Context) {
	var group models.Group
	if err := config.DB.First(&group, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Group not found")
		return
	}

	if err := config.DB.Delete(&group).Error; err != nil {
		utils.InternalServerError(c, "failed to delete group")
		return
	}
	utils.OKResponse(c, gin.H{"message": "Group deleted"})
}
