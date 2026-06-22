package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

type PermissionPolicy struct {
	ID       int    `json:"id"`
	Subject  string `json:"subject"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

func GetPermissions(c *gin.Context) {
	filter := c.Query("subject")

	var policies []PermissionPolicy
	allPolicies, _ := config.CEF.GetPolicy()

	for i, p := range allPolicies {
		if len(p) >= 3 {
			if filter != "" && !strings.HasPrefix(p[0], filter) {
				continue
			}
			policies = append(policies, PermissionPolicy{
				ID:       i + 1,
				Subject:  p[0],
				Resource: p[1],
				Action:   p[2],
			})
		}
	}

	utils.OKResponse(c, policies)
}

func GrantPermission(c *gin.Context) {
	var input struct {
		Subject  string `json:"subject"`
		UserID   uint   `json:"userId"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	var sub string
	if input.Subject != "" {
		sub = input.Subject
	} else if input.UserID != 0 {
		sub = fmt.Sprintf("user:%d", input.UserID)
	} else {
		utils.BadRequestError(c, "subject or userId is required")
		return
	}

	added, err := config.CEF.AddPolicy(sub, input.Resource, input.Action)
	if err != nil {
		utils.InternalServerError(c, "failed to add permission")
		return
	}
	if !added {
		utils.ConflictError(c, "Permission already exists")
		return
	}

	if err := config.CEF.SavePolicy(); err != nil {
		utils.InternalServerError(c, "failed to save policy")
		return
	}

	utils.CreatedResponse(c, gin.H{"message": "Permission granted", "subject": sub})
}

func RevokePermission(c *gin.Context) {
	var input struct {
		Subject  string `json:"subject"`
		UserID   uint   `json:"userId"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	var sub string
	if input.Subject != "" {
		sub = input.Subject
	} else if input.UserID != 0 {
		sub = fmt.Sprintf("user:%d", input.UserID)
	} else {
		utils.BadRequestError(c, "subject or userId is required")
		return
	}

	removed, err := config.CEF.RemovePolicy(sub, input.Resource, input.Action)
	if err != nil {
		utils.InternalServerError(c, "failed to remove permission")
		return
	}
	if !removed {
		utils.NotFoundError(c, "Permission not found")
		return
	}

	if err := config.CEF.SavePolicy(); err != nil {
		utils.InternalServerError(c, "failed to save policy")
		return
	}

	utils.OKResponse(c, gin.H{"message": "Permission revoked"})
}

func GetApplicationPermissions(c *gin.Context) {
	appID := c.Param("id")
	resource := fmt.Sprintf("/applications/%s", appID)

	var policies []PermissionPolicy
	allPolicies, _ := config.CEF.GetPolicy()

	for i, p := range allPolicies {
		if len(p) >= 3 && p[1] == resource {
			policies = append(policies, PermissionPolicy{
				ID:       i + 1,
				Subject:  p[0],
				Resource: p[1],
				Action:   p[2],
			})
		}
	}

	utils.OKResponse(c, policies)
}

func GrantApplicationPermission(c *gin.Context) {
	appID := c.Param("id")

	var input struct {
		Subject string `json:"subject" binding:"required"`
		Action  string `json:"action" binding:"required,oneof=read write"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input: subject (user:{id} or team:{id}) and action (read|write) are required")
		return
	}

	if !strings.HasPrefix(input.Subject, "user:") && !strings.HasPrefix(input.Subject, "team:") {
		utils.BadRequestError(c, "subject must be user:{id} or team:{id}")
		return
	}

	parts := strings.SplitN(input.Subject, ":", 2)
	subID, err := strconv.Atoi(parts[1])
	if err != nil || subID <= 0 {
		utils.BadRequestError(c, "invalid subject id")
		return
	}

	if parts[0] == "user" {
		var user models.User
		if err := config.DB.First(&user, subID).Error; err != nil {
			utils.NotFoundError(c, "User not found")
			return
		}
	} else {
		var team models.Team
		if err := config.DB.First(&team, subID).Error; err != nil {
			utils.NotFoundError(c, "Team not found")
			return
		}
	}

	var app models.Application
	if err := config.DB.First(&app, appID).Error; err != nil {
		utils.NotFoundError(c, "Application not found")
		return
	}

	resource := fmt.Sprintf("/applications/%s", appID)

	added, err := config.CEF.AddPolicy(input.Subject, resource, input.Action)
	if err != nil {
		utils.InternalServerError(c, "failed to grant permission")
		return
	}
	if !added {
		utils.ConflictError(c, "Permission already exists")
		return
	}
	if err := config.CEF.SavePolicy(); err != nil {
		utils.InternalServerError(c, "failed to save policy")
		return
	}

	utils.CreatedResponse(c, gin.H{"message": "Permission granted", "subject": input.Subject})
}

func RevokeApplicationPermission(c *gin.Context) {
	appID := c.Param("id")

	var input struct {
		Subject string `json:"subject" binding:"required"`
		Action  string `json:"action" binding:"required,oneof=read write"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	resource := fmt.Sprintf("/applications/%s", appID)

	removed, err := config.CEF.RemovePolicy(input.Subject, resource, input.Action)
	if err != nil {
		utils.InternalServerError(c, "failed to revoke permission")
		return
	}
	if !removed {
		utils.NotFoundError(c, "Permission not found")
		return
	}
	if err := config.CEF.SavePolicy(); err != nil {
		utils.InternalServerError(c, "failed to save policy")
		return
	}

	utils.OKResponse(c, gin.H{"message": "Permission revoked"})
}


