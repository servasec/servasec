package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetScannerTypes(c *gin.Context) {
	var scannerTypes []models.ScannerType
	q := config.DB.Order("name ASC")
	if c.Query("enabled") == "true" {
		q = q.Where("enabled = ?", true)
	}
	if err := q.Find(&scannerTypes).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch scanner types")
		return
	}
	utils.OKResponse(c, scannerTypes)
}

func UpdateScannerType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequestError(c, "invalid scanner type id")
		return
	}

	var scannerType models.ScannerType
	if err := config.DB.First(&scannerType, id).Error; err != nil {
		utils.NotFoundError(c, "scanner type not found")
		return
	}

	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequestError(c, "invalid request body")
		return
	}
	if body.Enabled == nil {
		utils.BadRequestError(c, "enabled field is required")
		return
	}

	scannerType.Enabled = *body.Enabled
	if err := config.DB.Save(&scannerType).Error; err != nil {
		utils.InternalServerError(c, "failed to update scanner type")
		return
	}

	utils.OKResponse(c, scannerType)
}
