package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
	"gorm.io/gorm"
)

func generateApiToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetApplications returns all applications the user has access to
// @Summary List applications
// @Tags Applications
// @Produce json
// @Success 200 {array} models.Application "List of applications"
// @Router /applications [get]
func GetApplications(c *gin.Context) {
	accessibleIDs := utils.GetAccessibleAppIDs(c)
	query := config.DB.Model(&models.Application{})
	if accessibleIDs != nil {
		if len(accessibleIDs) == 0 {
			utils.OKResponse(c, []models.Application{})
			return
		}
		query = query.Where("id IN ?", accessibleIDs)
	}

	var apps []models.Application
	if err := query.Find(&apps).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch applications")
		return
	}
	utils.OKResponse(c, apps)
}

// GetApplication returns a single application by ID or slug
// @Summary Get application
// @Tags Applications
// @Produce json
// @Param id path string true "Application ID or slug"
// @Success 200 {object} object "Application with default version"
// @Failure 404 {object} gin.H "Application not found"
// @Router /applications/{id} [get]
func GetApplication(c *gin.Context) {
	var app models.Application
	query := config.DB.Preload("Versions", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	})

	if slug := c.Param("slug"); slug != "" {
		query = query.Where("slug = ?", slug)
	} else {
		query = query.Where("id = ?", c.Param("id"))
	}

	if err := query.First(&app).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}
	var defaultVersion models.ApplicationVersion
	config.DB.Where("application_id = ? AND is_default = ?", app.ID, true).First(&defaultVersion)
	type appResponse struct {
		models.Application
		DefaultVersion *models.ApplicationVersion `json:"defaultVersion"`
	}
	resp := appResponse{Application: app}
	if defaultVersion.ID != 0 {
		resp.DefaultVersion = &defaultVersion
	}
	utils.OKResponse(c, resp)
}

// CreateApplication creates a new application with an auto-generated API token
// @Summary Create application
// @Tags Applications
// @Accept json
// @Produce json
// @Param input body object true "Application details"
// @Success 201 {object} models.Application "Created application"
// @Failure 400 {object} gin.H "Invalid input"
// @Router /applications [post]
func CreateApplication(c *gin.Context) {
	var input struct {
		Name             string `json:"name" binding:"required,max=200"`
		Description      string `json:"description" binding:"max=1000"`
		Slug             string `json:"slug" binding:"required,max=100"`
		GroupID          uint   `json:"groupId" binding:"required"`
		RepositoryURL    string `json:"repositoryUrl" binding:"max=500"`
		AssetCriticality string `json:"assetCriticality" binding:"omitempty,oneof=critical high medium low"`
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
		Name:             input.Name,
		Description:      input.Description,
		Slug:             input.Slug,
		GroupID:          input.GroupID,
		RepositoryURL:    input.RepositoryURL,
		ApiToken:         generateApiToken(),
		AssetCriticality: input.AssetCriticality,
	}
	if app.AssetCriticality == "" {
		app.AssetCriticality = "medium"
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

// UpdateApplication updates an existing application
// @Summary Update application
// @Tags Applications
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param input body object true "Fields to update"
// @Success 200 {object} models.Application "Updated application"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 404 {object} gin.H "Application not found"
// @Router /applications/{id} [put]
func UpdateApplication(c *gin.Context) {
	var app models.Application
	if err := config.DB.First(&app, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	var input struct {
		Name             *string `json:"name" binding:"omitempty,max=200"`
		Description      *string `json:"description" binding:"omitempty,max=1000"`
		Slug             *string `json:"slug" binding:"omitempty,max=100"`
		GroupID          *uint   `json:"groupId"`
		RepositoryURL    *string `json:"repositoryUrl" binding:"omitempty,max=500"`
		AssetCriticality *string `json:"assetCriticality" binding:"omitempty,oneof=critical high medium low"`
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
	if input.AssetCriticality != nil {
		app.AssetCriticality = *input.AssetCriticality
	}

	if err := config.DB.Save(&app).Error; err != nil {
		utils.InternalServerError(c, "failed to update application")
		return
	}
	utils.OKResponse(c, app)
}

// DeleteApplication deletes an application by ID
// @Summary Delete application
// @Tags Applications
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {object} gin.H "Application deleted"
// @Failure 404 {object} gin.H "Application not found"
// @Router /applications/{id} [delete]
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

// RegenerateApiToken generates a new API token for an application
// @Summary Regenerate API token
// @Tags Applications
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {object} gin.H "New API token"
// @Failure 404 {object} gin.H "Application not found"
// @Router /applications/{id}/regenerate-token [post]
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
