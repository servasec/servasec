package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/parsers"
	"github.com/servasec/servasec/backend/pro"
	"github.com/servasec/servasec/backend/services"
	"github.com/servasec/servasec/backend/utils"
	"gorm.io/gorm"
)

const maxIngestFileSize = 50 << 20 // 50 MB

func findApplication(c *gin.Context) (*models.Application, error) {
	if app, exists := c.Get("route_app"); exists {
		return app.(*models.Application), nil
	}

	apiToken := c.GetHeader("X-Api-Token")

	if apiToken != "" {
		var app models.Application
		if err := config.DB.Where("api_token = ?", apiToken).First(&app).Error; err != nil {
			return nil, fmt.Errorf("invalid API token")
		}
		return &app, nil
	}

	idParam := c.Param("id")
	if idParam == "" {
		return nil, fmt.Errorf("application not specified")
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid application ID")
	}

	var app models.Application
	if err := config.DB.First(&app, id).Error; err != nil {
		return nil, fmt.Errorf("application not found")
	}
	return &app, nil
}

func upsertVersion(appID uint, name, branch string) (*models.ApplicationVersion, error) {
	var version models.ApplicationVersion
	err := config.DB.Where("application_id = ? AND name = ?", appID, name).First(&version).Error
	if err == nil {
		if branch != "" && version.Branch != branch {
			config.DB.Model(&version).Update("branch", branch)
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

// IngestScan accepts a scan result file and creates findings
// @Summary Ingest scan results
// @Description Upload a scan results file. Authenticate via API token header, application ID in path, or application slug in path.
// @Tags Scans
// @Accept multipart/form-data
// @Produce json
// @Param id path string false "Application ID"
// @Param slug path string false "Application slug"
// @Param version formData string false "Version name"
// @Param branch formData string false "Branch name"
// @Param scannerType formData string false "Scanner type name"
// @Param file formData file true "Scan results file (SARIF, etc.)"
// @Param X-Api-Token header string false "API token (alternative to id/slug auth)"
// @Success 201 {object} gin.H "Scan created with findings count"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 404 {object} gin.H "Application not found"
// @Router /ingest [post]
// @Router /applications/{id}/ingest [post]
func IngestScan(c *gin.Context) {
	app, err := findApplication(c)
	if err != nil {
		utils.NotFoundError(c, err.Error())
		return
	}

	versionName := c.PostForm("version")
	if versionName == "" {
		versionName = c.DefaultQuery("version", "")
	}
	branch := c.PostForm("branch")

	// No version specified → use the app's default version
	var version *models.ApplicationVersion
	if versionName == "" {
		var defaultVersion models.ApplicationVersion
		if err := config.DB.Where("application_id = ? AND is_default = ?", app.ID, true).
			First(&defaultVersion).Error; err == nil {
			version = &defaultVersion
		} else {
			versionName = "latest"
		}
	}

	if version == nil {
		var err error
		version, err = upsertVersion(app.ID, versionName, branch)
		if err != nil {
			utils.InternalServerError(c, "failed to create version")
			return
		}
		if versionName == "latest" && branch == "" {
			config.DB.Model(&version).Update("is_default", true)
		}
	}

	scannerTypeName := c.PostForm("scannerType")
	if scannerTypeName == "" {
		scannerTypeName = c.DefaultQuery("scannerType", "")
	}

	// Read file once with size limit (Phase 2.2 + 2.5)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxIngestFileSize)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequestError(c, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		utils.BadRequestError(c, "failed to read file or file too large (max 50MB)")
		return
	}

	var scannerType models.ScannerType
	var format string
	if scannerTypeName != "" {
		if err := config.DB.Where("name = ?", scannerTypeName).First(&scannerType).Error; err != nil {
			utils.BadRequestError(c, fmt.Sprintf("unknown scanner type: %s", scannerTypeName))
			return
		}
		if scannerTypeName == "sarif" {
			format = "sarif"
		}
	} else {
		format = parsers.DetectScannerType(data)
		if format == "sarif" {
			if toolName := parsers.ExtractSarifToolName(data); toolName != "" {
				var st models.ScannerType
				if err := config.DB.Where("name = ?", toolName).First(&st).Error; err == nil && st.Enabled {
					scannerType = st
				}
			}
			if scannerType.ID == 0 {
				config.DB.Where("name = ?", "sarif").First(&scannerType)
			}
		} else if format != "" {
			config.DB.Where("name = ?", format).First(&scannerType)
		}
	}
	if scannerType.ID == 0 {
		utils.BadRequestError(c, "could not determine scanner type")
		return
	}

	if !scannerType.Enabled {
		utils.BadRequestError(c, fmt.Sprintf("scanner type '%s' has been disabled by an administrator", scannerType.Name))
		return
	}

	parserName := scannerType.Parser
	if format == "sarif" && scannerType.Name != "sarif" {
		parserName = "sarif"
	}
	parser, ok := parsers.Get(parserName)
	if !ok {
		now := time.Now()
		rawStr := string(data)
		scan := models.Scan{
			ApplicationVersionID: version.ID,
			ScannerTypeID:        scannerType.ID,
			Status:               "completed",
			StartedAt:            &now,
			CompletedAt:          &now,
			RawResults:           &rawStr,
		}
		if err := config.DB.Create(&scan).Error; err != nil {
			utils.InternalServerError(c, "failed to create scan")
			return
		}
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
		now := time.Now()
		rawStr := string(data)
		scan := models.Scan{
			ApplicationVersionID: version.ID,
			ScannerTypeID:        scannerType.ID,
			Status:               "failed",
			StartedAt:            &now,
			CompletedAt:          &now,
			RawResults:           &rawStr,
		}
		config.DB.Create(&scan)
		utils.InternalServerError(c, fmt.Sprintf("failed to parse results: %v", err))
		return
	}

	// Phase 1.1: Wrap scan + findings + policy eval in a transaction
	now := time.Now()
	rawStr := string(data)
	var scanResult models.Scan
	var inserted []models.Finding
	skipped := 0

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		scanResult = models.Scan{
			ApplicationVersionID: version.ID,
			ScannerTypeID:        scannerType.ID,
			Status:               "completed",
			StartedAt:            &now,
			CompletedAt:          &now,
			RawResults:           &rawStr,
		}
		if err := tx.Create(&scanResult).Error; err != nil {
			return fmt.Errorf("create scan: %w", err)
		}

		// Deduplicate within the application version
		var existingHashes []string
		tx.Model(&models.Finding{}).
			Where("application_version_id = ?", version.ID).
			Where("dedupe_hash <> ''").
			Pluck("dedupe_hash", &existingHashes)
		seen := make(map[string]bool, len(existingHashes))
		for _, h := range existingHashes {
			seen[h] = true
		}

		var batch []models.Finding
		for _, f := range findingsInput {
			hash := services.DedupeHash(scannerType.Name, f.RuleID, f.FilePath, f.LineStart, f.Severity)
			if seen[hash] {
				skipped++
				continue
			}
			seen[hash] = true

			finding := models.Finding{
				ScanID:               scanResult.ID,
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
				DedupeHash:           hash,
			}
			if s := pro.Risk.CalculateScore(f.Severity, nil, app.AssetCriticality, time.Now()); s != nil {
				finding.RiskScore = s
			}
			batch = append(batch, finding)
		}

		// Phase 2.1: Batch insert findings
		if len(batch) > 0 {
			if err := tx.CreateInBatches(&batch, 500).Error; err != nil {
				return fmt.Errorf("create findings: %w", err)
			}
			inserted = batch
		}

		return nil
	})

	if err != nil {
		utils.InternalServerError(c, "failed to record scan results")
		return
	}

	// Policy evaluation runs outside the transaction (may trigger webhooks)
	for i := range inserted {
		services.EvaluatePolicies("finding.created", &inserted[i], app, 0)
	}

	utils.CreatedResponse(c, gin.H{
		"id":                scanResult.ID,
		"status":            scanResult.Status,
		"versionId":         version.ID,
		"findingsCount":     len(inserted),
		"skippedDuplicates": skipped,
	})
}
