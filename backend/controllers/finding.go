package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetFindings(c *gin.Context) {
	var findings []models.Finding
	query := config.DB.Order("created_at DESC")

	if appID := c.Query("applicationId"); appID != "" {
		query = query.Where("application_id = ?", appID)
	}
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if scannerType := c.Query("scannerType"); scannerType != "" {
		query = query.Where("scanner_type = ?", scannerType)
	}

	if err := query.Find(&findings).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch findings")
		return
	}
	utils.OKResponse(c, findings)
}

func GetFinding(c *gin.Context) {
	var finding models.Finding
	if err := config.DB.First(&finding, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Finding not found")
		return
	}
	utils.OKResponse(c, finding)
}

func UpdateFindingStatus(c *gin.Context) {
	var finding models.Finding
	if err := config.DB.First(&finding, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Finding not found")
		return
	}

	var input struct {
		Status string `json:"status" binding:"required,oneof=open confirmed false_positive fixed"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid status")
		return
	}

	finding.Status = input.Status
	if err := config.DB.Save(&finding).Error; err != nil {
		utils.InternalServerError(c, "failed to update finding status")
		return
	}
	utils.OKResponse(c, finding)
}
