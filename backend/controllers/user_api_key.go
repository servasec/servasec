package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

type createApiKeyInput struct {
	Name string `json:"name" binding:"required,max=100"`
}

func CreateApiKey(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.UnauthorizedError(c, "unauthorized")
		return
	}
	currentUser := user.(*models.User)

	var input createApiKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, fmt.Sprintf("invalid input: %v", err))
		return
	}

	raw, err := utils.GenerateRandomString(32)
	if err != nil {
		utils.InternalServerError(c, "failed to generate API key")
		return
	}

	prefix := raw[:8]
	hash := sha256.Sum256([]byte(raw))

	key := models.UserApiKey{
		UserID:    currentUser.ID,
		Name:      input.Name,
		KeyHash:   hex.EncodeToString(hash[:]),
		KeyPrefix: "sc_" + prefix,
	}

	if err := config.DB.Create(&key).Error; err != nil {
		utils.InternalServerError(c, "failed to save API key")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        key.ID,
		"name":      key.Name,
		"keyPrefix": key.KeyPrefix,
		"key":       "sc_" + raw,
	})
}

func ListApiKeys(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.UnauthorizedError(c, "unauthorized")
		return
	}
	currentUser := user.(*models.User)

	var keys []models.UserApiKey
	if err := config.DB.Where("user_id = ?", currentUser.ID).Find(&keys).Error; err != nil {
		utils.InternalServerError(c, "failed to list API keys")
		return
	}

	utils.OKResponse(c, keys)
}

func RevokeApiKey(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.UnauthorizedError(c, "unauthorized")
		return
	}
	currentUser := user.(*models.User)

	var key models.UserApiKey
	if err := config.DB.Where("id = ? AND user_id = ?", c.Param("id"), currentUser.ID).First(&key).Error; err != nil {
		utils.NotFoundError(c, "API key not found")
		return
	}

	if err := config.DB.Delete(&key).Error; err != nil {
		utils.InternalServerError(c, "failed to revoke API key")
		return
	}

	utils.OKResponse(c, gin.H{"message": "API key revoked"})
}
