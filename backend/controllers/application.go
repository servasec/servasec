package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func generateApiToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetApplications(c *gin.Context) {
	var apps []models.Application
	if err := config.DB.Find(&apps).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch applications")
		return
	}
	utils.OKResponse(c, apps)
}

func GetApplication(c *gin.Context) {
	var app models.Application
	if err := config.DB.First(&app, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}
	utils.OKResponse(c, app)
}

func CreateApplication(c *gin.Context) {
	var input struct {
		Name          string `json:"name" binding:"required,max=200"`
		Description   string `json:"description" binding:"max=1000"`
		Slug          string `json:"slug" binding:"required,max=100"`
		GroupID       uint   `json:"groupId" binding:"required"`
		RepositoryURL string `json:"repositoryUrl" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	var group models.Group
	if err := config.DB.First(&group, input.GroupID).Error; err != nil {
		utils.NotFoundError(c, "Group not found")
		return
	}

	app := models.Application{
		Name:          input.Name,
		Description:   input.Description,
		Slug:          input.Slug,
		GroupID:       input.GroupID,
		RepositoryURL: input.RepositoryURL,
		ApiToken:      generateApiToken(),
	}
	if err := config.DB.Create(&app).Error; err != nil {
		utils.InternalServerError(c, "failed to create application")
		return
	}

	userID, _ := c.Get("user_id")
	sub := fmt.Sprintf("user:%v", userID)
	config.CEF.AddPolicy(sub, fmt.Sprintf("/applications/%d", app.ID), "*")
	config.CEF.SavePolicy()

	utils.CreatedResponse(c, app)
}

func UpdateApplication(c *gin.Context) {
	var app models.Application
	if err := config.DB.First(&app, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	var input struct {
		Name          *string `json:"name" binding:"omitempty,max=200"`
		Description   *string `json:"description" binding:"omitempty,max=1000"`
		Slug          *string `json:"slug" binding:"omitempty,max=100"`
		GroupID       *uint   `json:"groupId"`
		RepositoryURL *string `json:"repositoryUrl" binding:"omitempty,max=500"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Name != nil {
		app.Name = *input.Name
	}
	if input.Description != nil {
		app.Description = *input.Description
	}
	if input.Slug != nil {
		app.Slug = *input.Slug
	}
	if input.GroupID != nil {
		app.GroupID = *input.GroupID
	}
	if input.RepositoryURL != nil {
		app.RepositoryURL = *input.RepositoryURL
	}

	if err := config.DB.Save(&app).Error; err != nil {
		utils.InternalServerError(c, "failed to update application")
		return
	}
	utils.OKResponse(c, app)
}

func DeleteApplication(c *gin.Context) {
	var app models.Application
	if err := config.DB.First(&app, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	if err := config.DB.Delete(&app).Error; err != nil {
		utils.InternalServerError(c, "failed to delete application")
		return
	}
	utils.OKResponse(c, gin.H{"message": "Application deleted"})
}

func RegenerateApiToken(c *gin.Context) {
	var app models.Application
	if err := config.DB.First(&app, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	app.ApiToken = generateApiToken()
	if err := config.DB.Save(&app).Error; err != nil {
		utils.InternalServerError(c, "failed to regenerate API token")
		return
	}
	utils.OKResponse(c, gin.H{"apiToken": app.ApiToken})
}
