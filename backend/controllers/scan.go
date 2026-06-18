package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

type IngestScanInput struct {
	ScannerType string          `json:"scannerType" binding:"required"`
	Status      string          `json:"status" binding:"omitempty"`
	Findings    []IngestFinding `json:"findings"`
}

type IngestFinding struct {
	RuleID      string `json:"ruleId"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FilePath    string `json:"filePath"`
	LineStart   *int   `json:"lineStart"`
	LineEnd     *int   `json:"lineEnd"`
	CWEID       string `json:"cweId"`
	Remediation string `json:"remediation"`
}

func IngestScan(c *gin.Context) {
	apiToken := c.GetHeader("X-Api-Token")
	if apiToken == "" {
		utils.UnauthorizedError(c, "missing API token")
		return
	}

	var app models.Application
	if err := config.DB.Where("api_token = ?", apiToken).First(&app).Error; err != nil {
		utils.UnauthorizedError(c, "invalid API token")
		return
	}

	var input IngestScanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	scannerType := input.ScannerType
	status := input.Status
	if status == "" {
		status = "completed"
	}

	now := time.Now()
	scan := models.Scan{
		ApplicationID: app.ID,
		ScannerType:   scannerType,
		Status:        status,
		StartedAt:     &now,
		CompletedAt:   &now,
	}
	if err := config.DB.Create(&scan).Error; err != nil {
		utils.InternalServerError(c, "failed to record scan")
		return
	}

	for _, f := range input.Findings {
		finding := models.Finding{
			ScanID:        scan.ID,
			ApplicationID: app.ID,
			ScannerType:   scannerType,
			RuleID:        f.RuleID,
			Title:         f.Title,
			Severity:      f.Severity,
			Description:   f.Description,
			FilePath:      f.FilePath,
			LineStart:     f.LineStart,
			LineEnd:       f.LineEnd,
			CWEID:         f.CWEID,
			Remediation:   f.Remediation,
			Status:        "open",
		}
		if err := config.DB.Create(&finding).Error; err != nil {
			utils.InternalServerError(c, "failed to record findings")
			return
		}
	}

	utils.CreatedResponse(c, gin.H{
		"id":     scan.ID,
		"status": scan.Status,
	})
}

func GetScans(c *gin.Context) {
	appID := c.Query("applicationId")
	var scans []models.Scan
	query := config.DB.Order("created_at DESC")
	if appID != "" {
		query = query.Where("application_id = ?", appID)
	}
	if err := query.Find(&scans).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch scans")
		return
	}
	utils.OKResponse(c, scans)
}

func GetScan(c *gin.Context) {
	var scan models.Scan
	if err := config.DB.First(&scan, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Scan not found")
		return
	}
	utils.OKResponse(c, scan)
}
