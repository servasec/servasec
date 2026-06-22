package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetScannerTypes(c *gin.Context) {
	var scannerTypes []models.ScannerType
	if err := config.DB.Order("name ASC").Find(&scannerTypes).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch scanner types")
		return
	}
	utils.OKResponse(c, scannerTypes)
}
