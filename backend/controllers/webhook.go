package controllers

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func parseAppID(id string) uint {
	u, _ := strconv.ParseUint(id, 10, 64)
	return uint(u)
}

// GetWebhooks returns all webhooks for an application
// @Summary List webhooks
// @Tags Webhooks
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {array} models.Webhook "List of webhooks"
// @Router /applications/{id}/webhooks [get]
func GetWebhooks(c *gin.Context) {
	appID := c.Param("id")
	var webhooks []models.Webhook
	if err := config.DB.Where("application_id = ?", appID).Order("created_at DESC").Find(&webhooks).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch webhooks")
		return
	}
	utils.OKResponse(c, webhooks)
}

// CreateWebhook creates a new webhook for an application
// @Summary Create webhook
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param input body object true "Webhook details"
// @Success 201 {object} models.Webhook "Created webhook"
// @Failure 400 {object} gin.H "Invalid input"
// @Router /applications/{id}/webhooks [post]
func CreateWebhook(c *gin.Context) {
	appID := c.Param("id")

	var input struct {
		URL      string `json:"url" binding:"required,max=500"`
		Secret   string `json:"secret" binding:"max=64"`
		Events   string `json:"events" binding:"max=64"`
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

	// Phase 2.7: Validate webhook URL to prevent SSRF
	if err := validateWebhookURL(input.URL); err != nil {
		utils.BadRequestError(c, err.Error())
		return
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

// DeleteWebhook deletes a webhook by ID
// @Summary Delete webhook
// @Tags Webhooks
// @Produce json
// @Param id path string true "Application ID"
// @Param webhookId path string true "Webhook ID"
// @Success 204 "No content"
// @Failure 404 {object} gin.H "Webhook not found"
// @Router /applications/{id}/webhooks/{webhookId} [delete]
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

// validateWebhookURL checks that the URL is HTTPS and does not point to
// private/internal networks (SSRF prevention).
func validateWebhookURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if u.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS")
	}
	if u.Host == "" {
		return fmt.Errorf("webhook URL must have a valid host")
	}

	hostname := u.Hostname()
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return fmt.Errorf("webhook URL cannot point to localhost")
	}

	// Resolve hostname and check for private IPs
	ips, err := net.LookupIP(hostname)
	if err == nil {
		for _, ip := range ips {
			if isPrivateIP(ip) {
				return fmt.Errorf("webhook URL cannot point to private/internal networks")
			}
		}
	}

	// Also block common internal hostnames
	blocked := []string{"metadata.google.internal", "169.254.169.254"}
	for _, b := range blocked {
		if strings.EqualFold(hostname, b) {
			return fmt.Errorf("webhook URL cannot point to internal metadata services")
		}
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168) ||
			ip4[0] == 127 ||
			(ip4[0] == 169 && ip4[1] == 254)
	}
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}
