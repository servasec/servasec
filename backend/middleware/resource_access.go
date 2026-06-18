package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
)

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
