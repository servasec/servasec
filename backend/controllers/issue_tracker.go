package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/providers"
	"github.com/servasec/servasec/backend/utils"
)

// @Summary Get issue tracker configuration
// @Tags Issue Tracker
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {object} gin.H "Issue tracker config"
// @Failure 404 {object} gin.H "Application or issue tracker not found"
// @Router /applications/{id}/issue-tracker [get]
func GetIssueTracker(c *gin.Context) {
	appID := c.Param("id")

	var app models.Application
	if err := config.DB.First(&app, appID).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	var tracker models.IssueTracker
	if err := config.DB.Where("application_id = ?", app.ID).First(&tracker).Error; err != nil {
		utils.NotFoundError(c, "Issue tracker not configured")
		return
	}

	resp := gin.H{
		"id":                tracker.ID,
		"provider":          tracker.Provider,
		"repositoryUrl":     app.RepositoryURL,
		"severityThreshold": tracker.SeverityThreshold,
		"isActive":          tracker.IsActive,
		"authType":          tracker.AuthType,
		"createdAt":         tracker.CreatedAt,
		"updatedAt":         tracker.UpdatedAt,
	}

	if tracker.Provider == "github" {
		resp["hasToken"] = tracker.EncryptedToken != ""
		resp["hasGitHubAppKey"] = tracker.EncryptedGitHubAppKey != ""
		if tracker.GitHubAppID != nil {
			resp["githubAppId"] = *tracker.GitHubAppID
		}
		if tracker.GitHubInstallationID != nil {
			resp["githubInstallationId"] = *tracker.GitHubInstallationID
		}
	} else {
		resp["hasToken"] = tracker.EncryptedToken != ""
	}

	utils.OKResponse(c, resp)
}

// @Summary Upsert issue tracker configuration
// @Tags Issue Tracker
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param input body object true "Issue tracker config (provider, authType, token, severityThreshold, etc.)"
// @Success 200 {object} gin.H "Updated issue tracker"
// @Success 201 {object} gin.H "Created issue tracker"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 404 {object} gin.H "Application not found"
// @Router /applications/{id}/issue-tracker [post]
func UpsertIssueTracker(c *gin.Context) {
	appID := c.Param("id")

	var app models.Application
	if err := config.DB.First(&app, appID).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	if app.RepositoryURL == "" {
		utils.BadRequestError(c, "Application must have a repository URL configured")
		return
	}

	if err := utils.ValidateRepositoryURL(app.RepositoryURL); err != nil {
		utils.BadRequestError(c, fmt.Sprintf("Invalid repository URL: %s", err.Error()))
		return
	}

	var input struct {
		Provider          string `json:"provider" binding:"required"`
		AuthType          string `json:"authType"`
		Token             string `json:"token"`
		GitHubAppID       *int64 `json:"githubAppId"`
		GitHubAppKey      string `json:"githubAppKey"`
		GitHubInstallID   *int64 `json:"githubInstallationId"`
		SeverityThreshold string `json:"severityThreshold" binding:"required"`
		IsActive          *bool  `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input: provider and severityThreshold are required")
		return
	}

	if input.Provider != "github" && input.Provider != "gitlab" {
		utils.BadRequestError(c, "Provider must be 'github' or 'gitlab'")
		return
	}

	authType := input.AuthType
	if authType == "" {
		authType = "pat"
	}

	if authType != "pat" && authType != "github_app" {
		utils.BadRequestError(c, "authType must be 'pat' or 'github_app'")
		return
	}

	if authType == "github_app" && input.Provider != "github" {
		utils.BadRequestError(c, "GitHub App authentication is only supported for GitHub provider")
		return
	}

	validSeverities := map[string]bool{"critical": true, "high": true, "medium": true, "low": true, "info": true}
	if !validSeverities[input.SeverityThreshold] {
		utils.BadRequestError(c, "severityThreshold must be one of: critical, high, medium, low, info")
		return
	}

	key, err := utils.GetEncryptionKey()
	if err != nil {
		utils.InternalServerError(c, "Encryption key not configured")
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	var existing models.IssueTracker
	result := config.DB.Where("application_id = ?", app.ID).First(&existing)

	if result.Error == nil {
		updates := map[string]interface{}{
			"provider":           input.Provider,
			"auth_type":          authType,
			"severity_threshold": input.SeverityThreshold,
			"is_active":          isActive,
		}

		if authType == "pat" {
			if input.Token != "" {
				encToken, encIV, err := utils.Encrypt([]byte(input.Token), key)
				if err != nil {
					utils.InternalServerError(c, "Failed to encrypt token")
					return
				}
				updates["encrypted_token"] = encToken
				updates["encrypted_token_iv"] = encIV
			}
			updates["github_app_id"] = nil
			updates["github_installation_id"] = nil
			updates["encrypted_github_app_key"] = nil
			updates["encrypted_github_app_key_iv"] = nil
		} else if authType == "github_app" {
			if input.GitHubAppID == nil || input.GitHubInstallID == nil {
				utils.BadRequestError(c, "githubAppId and githubInstallationId are required for GitHub App auth")
				return
			}
			if input.GitHubAppKey == "" && existing.EncryptedGitHubAppKey == "" {
				utils.BadRequestError(c, "githubAppKey is required when setting up GitHub App auth")
				return
			}

			updates["github_app_id"] = *input.GitHubAppID
			updates["github_installation_id"] = *input.GitHubInstallID

			if input.GitHubAppKey != "" {
				encKey, encIV, err := utils.Encrypt([]byte(input.GitHubAppKey), key)
				if err != nil {
					utils.InternalServerError(c, "Failed to encrypt GitHub App key")
					return
				}
				updates["encrypted_github_app_key"] = encKey
				updates["encrypted_github_app_key_iv"] = encIV
			}
			updates["encrypted_token"] = nil
			updates["encrypted_token_iv"] = nil
		}

		if err := config.DB.Model(&existing).Updates(updates).Error; err != nil {
			utils.InternalServerError(c, "Failed to update issue tracker")
			return
		}

		utils.OKResponse(c, gin.H{
			"id":                existing.ID,
			"provider":          input.Provider,
			"repositoryUrl":     app.RepositoryURL,
			"severityThreshold": input.SeverityThreshold,
			"isActive":          isActive,
			"authType":          authType,
			"hasToken":          authType == "pat" && (input.Token != "" || existing.EncryptedToken != ""),
			"hasGitHubAppKey":   authType == "github_app" && (input.GitHubAppKey != "" || existing.EncryptedGitHubAppKey != ""),
		})
		return
	}

	if authType == "pat" {
		if input.Token == "" {
			utils.BadRequestError(c, "Token is required when creating a new PAT issue tracker")
			return
		}

		encToken, encIV, err := utils.Encrypt([]byte(input.Token), key)
		if err != nil {
			utils.InternalServerError(c, "Failed to encrypt token")
			return
		}

		tracker := models.IssueTracker{
			ApplicationID:     app.ID,
			Provider:          input.Provider,
			AuthType:          authType,
			EncryptedToken:    encToken,
			EncryptedTokenIV:  encIV,
			SeverityThreshold: input.SeverityThreshold,
			IsActive:          isActive,
		}

		if err := config.DB.Create(&tracker).Error; err != nil {
			utils.InternalServerError(c, "Failed to create issue tracker")
			return
		}

		utils.CreatedResponse(c, gin.H{
			"id":                tracker.ID,
			"provider":          tracker.Provider,
			"repositoryUrl":     app.RepositoryURL,
			"severityThreshold": tracker.SeverityThreshold,
			"isActive":          tracker.IsActive,
			"authType":          authType,
			"hasToken":          true,
		})
	} else {
		if input.GitHubAppID == nil || input.GitHubInstallID == nil || input.GitHubAppKey == "" {
			utils.BadRequestError(c, "githubAppId, githubAppKey, and githubInstallationId are required for GitHub App auth")
			return
		}

		encKey, encIV, err := utils.Encrypt([]byte(input.GitHubAppKey), key)
		if err != nil {
			utils.InternalServerError(c, "Failed to encrypt GitHub App key")
			return
		}

		tracker := models.IssueTracker{
			ApplicationID:          app.ID,
			Provider:               input.Provider,
			AuthType:               authType,
			GitHubAppID:            input.GitHubAppID,
			GitHubInstallationID:   input.GitHubInstallID,
			EncryptedGitHubAppKey:  encKey,
			EncryptedGitHubAppKeyIV: encIV,
			SeverityThreshold:      input.SeverityThreshold,
			IsActive:               isActive,
		}

		if err := config.DB.Create(&tracker).Error; err != nil {
			utils.InternalServerError(c, "Failed to create issue tracker")
			return
		}

		utils.CreatedResponse(c, gin.H{
			"id":                tracker.ID,
			"provider":          tracker.Provider,
			"repositoryUrl":     app.RepositoryURL,
			"severityThreshold": tracker.SeverityThreshold,
			"isActive":          tracker.IsActive,
			"authType":          authType,
			"hasGitHubAppKey":   true,
		})
	}
}

// @Summary Delete issue tracker configuration
// @Tags Issue Tracker
// @Produce json
// @Param id path string true "Application ID"
// @Success 204 "No content"
// @Failure 404 {object} gin.H "Issue tracker not found"
// @Router /applications/{id}/issue-tracker [delete]
func DeleteIssueTracker(c *gin.Context) {
	appID := c.Param("id")
	var tracker models.IssueTracker
	if err := config.DB.Where("application_id = ?", appID).First(&tracker).Error; err != nil {
		utils.NotFoundError(c, "Issue tracker not found")
		return
	}

	if err := config.DB.Delete(&tracker).Error; err != nil {
		utils.InternalServerError(c, "Failed to delete issue tracker")
		return
	}

	utils.NoContentResponse(c)
}

// @Summary Test issue tracker connection
// @Tags Issue Tracker
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {object} gin.H "Connection successful"
// @Failure 400 {object} gin.H "Connection failed"
// @Failure 404 {object} gin.H "Application or issue tracker not found"
// @Router /applications/{id}/issue-tracker/test [post]
func TestIssueTracker(c *gin.Context) {
	appID := c.Param("id")

	var app models.Application
	if err := config.DB.First(&app, appID).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	var tracker models.IssueTracker
	if err := config.DB.Where("application_id = ?", app.ID).First(&tracker).Error; err != nil {
		utils.NotFoundError(c, "Issue tracker not configured")
		return
	}

	key, err := utils.GetEncryptionKey()
	if err != nil {
		utils.InternalServerError(c, "Encryption key not configured")
		return
	}

	cfg := providers.IssueTrackerConfig{
		Provider:      tracker.Provider,
		RepositoryURL: app.RepositoryURL,
		AuthType:      tracker.AuthType,
	}

	if tracker.AuthType == "github_app" {
		if tracker.GitHubAppID == nil || tracker.GitHubInstallationID == nil || tracker.EncryptedGitHubAppKey == "" {
			utils.BadRequestError(c, "GitHub App configuration incomplete")
			return
		}

		privateKey, err := utils.Decrypt(tracker.EncryptedGitHubAppKey, tracker.EncryptedGitHubAppKeyIV, key)
		if err != nil {
			utils.InternalServerError(c, "Failed to decrypt GitHub App key")
			return
		}

		cfg.GitHubAppID = tracker.GitHubAppID
		cfg.GitHubInstallationID = tracker.GitHubInstallationID
		cfg.GitHubPrivateKey = privateKey
	} else {
		if tracker.EncryptedToken == "" {
			utils.BadRequestError(c, "Token not configured")
			return
		}

		token, err := utils.Decrypt(tracker.EncryptedToken, tracker.EncryptedTokenIV, key)
		if err != nil {
			utils.InternalServerError(c, "Failed to decrypt token")
			return
		}
		cfg.Token = token
	}

	provider, ok := providers.Get(tracker.Provider)
	if !ok {
		utils.BadRequestError(c, fmt.Sprintf("Unknown provider: %s", tracker.Provider))
		return
	}

	if err := provider.TestConnection(cfg); err != nil {
		utils.BadRequestError(c, fmt.Sprintf("Connection test failed: %s", err.Error()))
		return
	}

	utils.OKResponse(c, gin.H{"message": "Connection successful"})
}

// @Summary List issue tracker issues associated with findings
// @Tags Issue Tracker
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {array} models.IssueTrackerIssue "List of issues"
// @Failure 404 {object} gin.H "Issue tracker not configured"
// @Router /applications/{id}/issue-tracker/issues [get]
func ListIssueTrackerIssues(c *gin.Context) {
	appID := c.Param("id")
	var tracker models.IssueTracker
	if err := config.DB.Where("application_id = ?", appID).First(&tracker).Error; err != nil {
		utils.NotFoundError(c, "Issue tracker not configured")
		return
	}

	var issues []models.IssueTrackerIssue
	if err := config.DB.
		Preload("Finding").
		Where("issue_tracker_id = ?", tracker.ID).
		Order("created_at DESC").
		Find(&issues).Error; err != nil {
		utils.InternalServerError(c, "Failed to fetch issues")
		return
	}

	utils.OKResponse(c, issues)
}
