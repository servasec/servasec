package controllers

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/parsers"
	"github.com/servasec/servasec/backend/pro"
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
// @Router /api/ingest [post]
// @Router /applications/{id}/ingest [post]
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
		file, _, err := c.Request.FormFile("file")
		if err == nil {
			data, _ := io.ReadAll(file)
			file.Close()
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
	}
	if scannerType.ID == 0 {
		utils.BadRequestError(c, "could not determine scanner type")
		return
	}

	if !scannerType.Enabled {
		utils.BadRequestError(c, fmt.Sprintf("scanner type '%s' has been disabled by an administrator", scannerType.Name))
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

	parserName := scannerType.Parser
	if format == "sarif" && scannerType.Name != "sarif" {
		parserName = "sarif"
	}
	parser, ok := parsers.Get(parserName)
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

	// Deduplicate within the application version: a finding's identity is the
	// hash of (scanner, rule, file, line, severity). Skip any finding whose hash
	// already exists for this version or repeats within this upload. The same
	// finding may reappear in a newer version (e.g. after a code change) and
	// should not be suppressed.
	var existingHashes []string
	config.DB.Model(&models.Finding{}).
		Where("application_version_id = ?", version.ID).
		Where("dedupe_hash <> ''").
		Pluck("dedupe_hash", &existingHashes)
	seen := make(map[string]bool, len(existingHashes))
	for _, h := range existingHashes {
		seen[h] = true
	}

	var inserted []models.Finding
	skipped := 0
	for _, f := range findingsInput {
		hash := services.DedupeHash(scannerType.Name, f.RuleID, f.FilePath, f.LineStart, f.Severity)
		if seen[hash] {
			skipped++
			continue
		}
		seen[hash] = true

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
			DedupeHash:           hash,
		}
		if s := pro.Risk.CalculateScore(f.Severity, nil, app.AssetCriticality, time.Now()); s != nil {
			finding.RiskScore = s
		}
		if err := config.DB.Create(&finding).Error; err != nil {
			utils.InternalServerError(c, "failed to record findings")
			return
		}
		inserted = append(inserted, finding)
	}

	scan.Status = "completed"
	scan.CompletedAt = &now
	config.DB.Save(&scan)

	for i := range inserted {
		services.EvaluatePolicies("finding.created", &inserted[i], app, 0)
	}

	utils.CreatedResponse(c, gin.H{
		"id":                scan.ID,
		"status":            scan.Status,
		"versionId":         version.ID,
		"findingsCount":     len(inserted),
		"skippedDuplicates": skipped,
	})
}
