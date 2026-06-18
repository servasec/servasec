package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func CheckPolicy(obj string, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := utils.GetClaimsFromCookie(c)

		var sub string
		if claims == nil {
			sub = "anonymous"
		} else {
			sub = fmt.Sprint(claims.Role)
			if sub == "" {
				sub = "anonymous"
			}
		}

		ok, err := config.CEF.Enforce(sub, obj, act)
		if err != nil {
			debug.Log("Casbin error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization_error"})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized: wrong permissions"})
			return
		}

		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := utils.GetClaimsFromCookie(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err})
			return
		}
		var user models.User
		if err := config.DB.First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		if user.Banned {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "banned"})
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user", &user)
		c.Next()
	}
}
