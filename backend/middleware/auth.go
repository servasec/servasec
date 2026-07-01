package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		authMethod := c.GetString("auth_method")

		var sub string
		switch authMethod {
		case "app_token":
			sub = "app_token"
		case "api_key":
			if user, exists := c.Get("user"); exists {
				sub = fmt.Sprint(user.(*models.User).Role)
			}
		default:
			claims := resolveClaims(c)
			if claims != nil {
				if user, exists := c.Get("user"); exists {
					sub = fmt.Sprint(user.(*models.User).Role)
				} else {
					sub = fmt.Sprint(claims.Role)
				}
			}
		}

		if sub == "" {
			sub = "anonymous"
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

func setUserContext(c *gin.Context, user *models.User) {
	c.Set("user_id", user.ID)
	c.Set("user", user)
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
		if claims != nil {
			var user models.User
			if err := config.DB.First(&user, claims.UserID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}
			if user.Banned {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "banned"})
				return
			}
			setUserContext(c, &user)
			c.Set("auth_method", "cookie")
			c.Next()
			return
		}

		claims = getClaimsFromBearer(c)
		if claims != nil {
			var user models.User
			if err := config.DB.First(&user, claims.UserID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}
			if user.Banned {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "banned"})
				return
			}
			setUserContext(c, &user)
			c.Set("auth_method", "bearer")
			c.Next()
			return
		}

		if apiToken := c.GetHeader("X-Api-Token"); apiToken != "" {
			var app models.Application
			if err := config.DB.Where("api_token = ?", apiToken).First(&app).Error; err == nil {
				c.Set("app", &app)
				c.Set("auth_method", "app_token")
				c.Next()
				return
			}
		}

		if apiKey := c.GetHeader("X-Api-Key"); apiKey != "" {
			hash := sha256.Sum256([]byte(strings.TrimPrefix(apiKey, "sc_")))
			hashStr := hex.EncodeToString(hash[:])
			var key models.UserApiKey
			if err := config.DB.Where("key_hash = ?", hashStr).First(&key).Error; err == nil {
				if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
					return
				}
				var user models.User
				if err := config.DB.First(&user, key.UserID).Error; err != nil {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
					return
				}
				if user.Banned {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "banned"})
					return
				}
				now := time.Now()
				config.DB.Model(&key).Update("last_used_at", now)
				setUserContext(c, &user)
				c.Set("auth_method", "api_key")
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}

func AppFromContext(c *gin.Context) (*models.Application, bool) {
	app, exists := c.Get("app")
	if !exists {
		return nil, false
	}
	return app.(*models.Application), true
}

func AppIDFromContext(c *gin.Context) (uint, bool) {
	app, exists := AppFromContext(c)
	if !exists {
		return 0, false
	}
	return app.ID, true
}

func AppIDStrFromContext(c *gin.Context) string {
	id, ok := AppIDFromContext(c)
	if !ok {
		return ""
	}
	return strconv.FormatUint(uint64(id), 10)
}
