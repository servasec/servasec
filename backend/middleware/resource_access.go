package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
)

func resolveGroupAccess(uid uint, resourcePath string, action string) bool {
	if !strings.HasPrefix(resourcePath, "/applications/") {
		return false
	}
	parts := strings.Split(resourcePath, "/")
	if len(parts) != 3 {
		return false
	}
	appID, err := strconv.Atoi(parts[2])
	if err != nil {
		return false
	}

	var app models.Application
	if err := config.DB.Select("group_id").First(&app, appID).Error; err != nil {
		return false
	}
	groupPath := fmt.Sprintf("/groups/%d", app.GroupID)

	userSub := fmt.Sprintf("user:%d", uid)
	if ok, _ := config.CEF.Enforce(userSub, groupPath, action); ok {
		return true
	}

	var teamMembers []models.TeamMember
	config.DB.Where("user_id = ?", uid).Find(&teamMembers)
	for _, tm := range teamMembers {
		teamSub := fmt.Sprintf("team:%d", tm.TeamID)
		if ok, _ := config.CEF.Enforce(teamSub, groupPath, action); ok {
			return true
		}
	}

	return false
}

func RequireResourceAccess(resourcePath string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		currentUser := user.(*models.User)
		if currentUser.Role == "admin" {
			c.Next()
			return
		}

		uid := userID.(uint)
		userSub := fmt.Sprintf("user:%d", uid)
		if ok, _ := config.CEF.Enforce(userSub, resourcePath, action); ok {
			c.Next()
			return
		}

		var teamMembers []models.TeamMember
		if err := config.DB.Where("user_id = ?", uid).Find(&teamMembers).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
			return
		}
		for _, tm := range teamMembers {
			teamSub := fmt.Sprintf("team:%d", tm.TeamID)
			if ok, _ := config.CEF.Enforce(teamSub, resourcePath, action); ok {
				c.Next()
				return
			}
		}

		if resolveGroupAccess(uid, resourcePath, action) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized: insufficient permissions"})
	}
}

func RequireResourceAccessByParam(resourceType string, paramName string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID := c.Param(paramName)
		resourcePath := fmt.Sprintf("/%s/%s", resourceType, resourceID)
		RequireResourceAccess(resourcePath, action)(c)
	}
}
