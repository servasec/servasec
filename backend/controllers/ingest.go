package controllers

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/features"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/parsers"
	"github.com/servasec/servasec/backend/services"
	"github.com/servasec/servasec/backend/utils"
)

func findApplication(c *gin.Context) (*models.Application, error) {
	apiToken := c.GetHeader("X-Api-Token")

	idParam := c.Param("id")
	slugParam := c.Param("slug")

	switch {
	case slugParam != "":
		var app models.Application
		if err := config.DB.Where("slug = ?", slugParam).First(&app).Error; err != nil {
			return nil, fmt.Errorf("application not found")
		}
		return &app, nil

	case apiToken != "":
		var app models.Application
		if err := config.DB.Where("api_token = ?", apiToken).First(&app).Error; err != nil {
			return nil, fmt.Errorf("invalid API token")
		}
		return &app, nil

	case idParam != "":
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid application ID")
		}
		var app models.Application
		if err := config.DB.First(&app, id).Error; err != nil {
			return nil, fmt.Errorf("application not found")
		}
		return &app, nil

	default:
		return nil, fmt.Errorf("application not specified")
	}
}

func upsertVersion(appID uint, name, branch string) (*models.ApplicationVersion, error) {
	var version models.ApplicationVersion
	err := config.DB.Where("application_id = ? AND name = ?", appID, name).First(&version).Error
	if err == nil {
		if branch != "" && version.Branch != branch {
			version.Branch = branch
			config.DB.Save(&version)
		}
		return &version, nil
	}

	version = models.ApplicationVersion{
		ApplicationID: appID,
		Name:          name,
		Branch:        branch,
	}
	if err := config.DB.Create(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func IngestScan(c *gin.Context) {
	app, err := findApplication(c)
	if err != nil {
		utils.NotFoundError(c, err.Error())
		return
	}

	versionName := c.PostForm("version")
	if versionName == "" {
		versionName = c.DefaultQuery("version", "latest")
	}
	branch := c.PostForm("branch")

	version, err := upsertVersion(app.ID, versionName, branch)
	if err != nil {
		utils.InternalServerError(c, "failed to create version")
		return
	}
	if versionName == "latest" && branch == "" {
		version.IsDefault = true
		config.DB.Save(version)
	}

	scannerTypeName := c.PostForm("scannerType")
	if scannerTypeName == "" {
		scannerTypeName = c.DefaultQuery("scannerType", "")
	}

	var scannerType models.ScannerType
	if scannerTypeName != "" {
		if err := config.DB.Where("name = ?", scannerTypeName).First(&scannerType).Error; err != nil {
			utils.BadRequestError(c, fmt.Sprintf("unknown scanner type: %s", scannerTypeName))
			return
		}
	} else {
		file, _, err := c.Request.FormFile("file")
		if err == nil {
			data, _ := io.ReadAll(file)
			file.Close()
			detected := parsers.DetectScannerType(data)
			if detected != "" {
				config.DB.Where("name = ?", detected).First(&scannerType)
			}
		}
	}
	if scannerType.ID == 0 {
		utils.BadRequestError(c, "could not determine scanner type")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequestError(c, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		utils.InternalServerError(c, "failed to read file")
		return
	}

	now := time.Now()
	rawStr := string(data)
	scan := models.Scan{
		ApplicationVersionID: version.ID,
		ScannerTypeID:        scannerType.ID,
		Status:               "processing",
		StartedAt:            &now,
		RawResults:           &rawStr,
	}
	if err := config.DB.Create(&scan).Error; err != nil {
		utils.InternalServerError(c, "failed to create scan")
		return
	}

	parser, ok := parsers.Get(scannerType.Parser)
	if !ok {
		scan.Status = "completed"
		scan.CompletedAt = &now
		config.DB.Save(&scan)
		utils.CreatedResponse(c, gin.H{
			"id":            scan.ID,
			"status":        scan.Status,
			"versionId":     version.ID,
			"findingsCount": 0,
			"warning":       fmt.Sprintf("no parser available for %s, raw results stored", scannerType.Name),
		})
		return
	}

	findingsInput, err := parser(data, header.Filename)
	if err != nil {
		scan.Status = "failed"
		scan.CompletedAt = &now
		config.DB.Save(&scan)
		utils.InternalServerError(c, fmt.Sprintf("failed to parse results: %v", err))
		return
	}

	for _, f := range findingsInput {
		finding := models.Finding{
			ScanID:               scan.ID,
			ApplicationVersionID: version.ID,
			ScannerTypeID:        scannerType.ID,
			RuleID:               f.RuleID,
			Title:                f.Title,
			Severity:             f.Severity,
			Description:          f.Description,
			FilePath:             f.FilePath,
			LineStart:            f.LineStart,
			LineEnd:              f.LineEnd,
			CWEID:                f.CWEID,
			Remediation:          f.Remediation,
			Status:               "open",
		}
		if features.F.IsEnabled(features.FeatureRiskScoring) {
			now := time.Now()
			score := services.CalculateRiskScore(f.Severity, nil, app.AssetCriticality, now)
			finding.RiskScore = &score
		}
		if err := config.DB.Create(&finding).Error; err != nil {
			utils.InternalServerError(c, "failed to record findings")
			return
		}
	}

	scan.Status = "completed"
	scan.CompletedAt = &now
	config.DB.Save(&scan)

	FireWebhooks(app.ID, scan.ID, findingsInput)

	utils.CreatedResponse(c, gin.H{
		"id":            scan.ID,
		"status":        scan.Status,
		"versionId":     version.ID,
		"findingsCount": len(findingsInput),
	})
}
