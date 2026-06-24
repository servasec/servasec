package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func resolveClaims(c *gin.Context) *utils.TokenClaims {
	claims, _ := utils.GetClaimsFromCookie(c)
	if claims != nil {
		return claims
	}
	return getClaimsFromBearer(c)
}

func CheckPolicy(obj string, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := resolveClaims(c)

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

func getClaimsFromBearer(c *gin.Context) *utils.TokenClaims {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	if utils.IsTokenBlacklisted(tokenStr) {
		return nil
	}
	token, err := jwt.ParseWithClaims(tokenStr, &utils.TokenClaims{}, func(t *jwt.Token) (any, error) {
		return utils.AccessSecret, nil
	})
	if err != nil {
		return nil
	}
	return token.Claims.(*utils.TokenClaims)
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := utils.GetClaimsFromCookie(c)
		if claims == nil {
			claims = getClaimsFromBearer(c)
		}
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
