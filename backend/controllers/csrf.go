package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/middleware"
	"github.com/servasec/servasec/backend/utils"
)

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
