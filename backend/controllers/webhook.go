package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func parseAppID(id string) uint {
	u, _ := strconv.ParseUint(id, 10, 64)
	return uint(u)
}

func GetWebhooks(c *gin.Context) {
	appID := c.Param("id")
	var webhooks []models.Webhook
	if err := config.DB.Where("application_id = ?", appID).Order("created_at DESC").Find(&webhooks).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch webhooks")
		return
	}
	utils.OKResponse(c, webhooks)
}

func CreateWebhook(c *gin.Context) {
	appID := c.Param("id")

	var input struct {
		URL      string `json:"url" binding:"required,max=500"`
		Secret   string `json:"secret" binding:"max=64"`
		Events   string `json:"events" binding:"required"`
		IsActive *bool  `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	webhook := models.Webhook{
		ApplicationID: parseAppID(appID),
		URL:           input.URL,
		Secret:        input.Secret,
		Events:        input.Events,
		IsActive:      isActive,
	}
	if err := config.DB.Create(&webhook).Error; err != nil {
		utils.InternalServerError(c, "failed to create webhook")
		return
	}
	utils.CreatedResponse(c, webhook)
}

func DeleteWebhook(c *gin.Context) {
	var webhook models.Webhook
	if err := config.DB.First(&webhook, c.Param("webhookId")).Error; err != nil {
		utils.NotFoundError(c, "Webhook not found")
		return
	}
	if err := config.DB.Delete(&webhook).Error; err != nil {
		utils.InternalServerError(c, "failed to delete webhook")
		return
	}
	utils.NoContentResponse(c)
}
