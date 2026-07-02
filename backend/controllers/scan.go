package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

// GetScans returns scans with optional filters
// @Summary List scans
// @Tags Scans
// @Produce json
// @Param applicationId query string false "Filter by application ID"
// @Param applicationVersionId query string false "Filter by version ID"
// @Success 200 {array} models.Scan "List of scans"
// @Failure 500 {object} gin.H "Failed to fetch scans"
// @Router /scans [get]
func GetScans(c *gin.Context) {
	appID := c.Query("applicationId")
	versionID := c.Query("applicationVersionId")
	var scans []models.Scan
	query := config.DB.
		Select("scans.*, (SELECT COUNT(*) FROM findings WHERE findings.scan_id = scans.id) as findings_count").
		Preload("ScannerType").
		Preload("ApplicationVersion").
		Joins("JOIN application_versions ON application_versions.id = scans.application_version_id").
		Order("scans.created_at DESC")

	accessibleIDs := utils.GetAccessibleAppIDs(c)
	if accessibleIDs != nil {
		if len(accessibleIDs) == 0 {
			utils.OKResponse(c, []models.Scan{})
			return
		}
		query = query.Where("application_versions.application_id IN ?", accessibleIDs)
	}

	if appID != "" {
		query = query.Where("application_versions.application_id = ?", appID)
	}
	if versionID != "" {
		query = query.Where("scans.application_version_id = ?", versionID)
	}

	if err := query.Find(&scans).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch scans")
		return
	}
	utils.OKResponse(c, scans)
}

// GetScan returns a single scan by ID
// @Summary Get scan
// @Tags Scans
// @Produce json
// @Param id path string true "Scan ID"
// @Success 200 {object} models.Scan "Scan details"
// @Failure 404 {object} gin.H "Scan not found"
// @Router /scans/{id} [get]
func GetScan(c *gin.Context) {
	var scan models.Scan
	if err := config.DB.Preload("ScannerType").Preload("ApplicationVersion").First(&scan, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Scan not found")
		return
	}
	utils.OKResponse(c, scan)
}
