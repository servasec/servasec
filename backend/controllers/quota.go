package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/services"
	"github.com/servasec/servasec/backend/utils"
)

// GetQuotas returns the current license quota limits and usage.
// @Summary Get quota limits and usage
// @Tags Auth
// @Produce json
// @Success 200 {object} object "Quota status per resource"
// @Router /me/quotas [get]
func GetQuotas(c *gin.Context) {
	statuses, err := services.GetQuotaStatuses()
	if err != nil {
		utils.InternalServerError(c, "failed to fetch quota status")
		return
	}
	utils.OKResponse(c, gin.H{"resources": statuses})
}
