package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/middleware"
	"github.com/servasec/servasec/backend/utils"
)

// GetCSRFToken returns a CSRF token for the current session
// @Summary Get CSRF token
// @Tags Auth
// @Produce json
// @Success 200 {object} gin.H "CSRF token"
// @Failure 500 {object} gin.H "CSRF token not available"
// @Router /csrf-token [get]
func GetCSRFToken(c *gin.Context) {
	token := middleware.GetCSRFToken(c)
	if token == "" {
		utils.InternalServerError(c, "csrf_token_not_available")
		return
	}

	utils.OKResponse(c, gin.H{
		"csrfToken": token,
	})
}
