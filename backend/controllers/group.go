package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetGroups(c *gin.Context) {
	var groups []models.Group
	if err := config.DB.Find(&groups).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch groups")
		return
	}
	utils.OKResponse(c, groups)
}

func GetGroup(c *gin.Context) {
	var group models.Group
	if err := config.DB.First(&group, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Group not found")
		return
	}
	utils.OKResponse(c, group)
}

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
