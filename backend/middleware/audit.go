package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
)

func AuditLog() gin.HandlerFunc {
	methods := map[string]bool{
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}

	skipPaths := map[string]bool{
		"/login":    true,
		"/register": true,
		"/refresh":  true,
	}

	return func(c *gin.Context) {
		if !methods[c.Request.Method] {
			c.Next()
			return
		}

		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		c.Next()

		var uid uint
		if id, exists := c.Get("user_id"); exists {
			uid, _ = id.(uint)
		}

		user, _ := c.Get("user")
		var username string
		if u, ok := user.(*models.User); ok {
			username = u.Username
		}

		details := ""
		ct := c.Request.Header.Get("Content-Type")
		if strings.Contains(ct, "application/json") && len(bodyBytes) > 0 {
			if len(bodyBytes) <= 500 {
				details = string(bodyBytes)
			}
		}

		config.DB.Create(&models.AuditLog{
			UserID:    uid,
			Username:  username,
			Action:    c.Request.Method,
			Resource:  c.Request.URL.Path,
			Status:    c.Writer.Status(),
			Details:   details,
			IP:        c.ClientIP(),
			CreatedAt: time.Now(),
		})
	}
}
