package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetVersions(c *gin.Context) {
	appID := c.Param("id")
	var versions []models.ApplicationVersion
	if err := config.DB.Where("application_id = ?", appID).Order("created_at DESC").Find(&versions).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch versions")
		return
	}
	utils.OKResponse(c, versions)
}

func GetVersion(c *gin.Context) {
	var version models.ApplicationVersion
	if err := config.DB.Preload("Scans").First(&version, c.Param("versionId")).Error; err != nil {
		utils.NotFoundError(c, "Version not found")
		return
	}
	utils.OKResponse(c, version)
}

func CreateVersion(c *gin.Context) {
	appID := c.Param("id")

	var input struct {
		Name      string `json:"name" binding:"required,max=100"`
		Branch    string `json:"branch" binding:"max=200"`
		Tag       string `json:"tag" binding:"max=100"`
		IsDefault bool   `json:"isDefault"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	var existing models.ApplicationVersion
	err := config.DB.Where("application_id = ? AND name = ?", appID, input.Name).First(&existing).Error
	if err == nil {
		utils.ConflictError(c, "Version with this name already exists")
		return
	}

	appIDUint, _ := strconv.ParseUint(appID, 10, 64)
	version := models.ApplicationVersion{
		ApplicationID: uint(appIDUint),
		Name:          input.Name,
		Branch:        input.Branch,
		Tag:           input.Tag,
		IsDefault:     input.IsDefault,
	}
	if input.IsDefault {
		config.DB.Model(&models.ApplicationVersion{}).
			Where("application_id = ?", appID).
			Update("is_default", false)
	}
	if err := config.DB.Create(&version).Error; err != nil {
		utils.InternalServerError(c, "failed to create version")
		return
	}
	utils.CreatedResponse(c, version)
}

func UpdateVersion(c *gin.Context) {
	var version models.ApplicationVersion
	if err := config.DB.First(&version, c.Param("versionId")).Error; err != nil {
		utils.NotFoundError(c, "Version not found")
		return
	}

	var input struct {
		Name      *string `json:"name" binding:"omitempty,max=100"`
		Branch    *string `json:"branch" binding:"omitempty,max=200"`
		Tag       *string `json:"tag" binding:"omitempty,max=100"`
		IsDefault *bool   `json:"isDefault"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Name != nil {
		version.Name = *input.Name
	}
	if input.Branch != nil {
		version.Branch = *input.Branch
	}
	if input.Tag != nil {
		version.Tag = *input.Tag
	}
	if input.IsDefault != nil && *input.IsDefault {
		config.DB.Model(&models.ApplicationVersion{}).
			Where("application_id = ?", version.ApplicationID).
			Update("is_default", false)
		version.IsDefault = true
	}

	if err := config.DB.Save(&version).Error; err != nil {
		utils.InternalServerError(c, "failed to update version")
		return
	}
	utils.OKResponse(c, version)
}

func DeleteVersion(c *gin.Context) {
	var version models.ApplicationVersion
	if err := config.DB.First(&version, c.Param("versionId")).Error; err != nil {
		utils.NotFoundError(c, "Version not found")
		return
	}
	if err := config.DB.Delete(&version).Error; err != nil {
		utils.InternalServerError(c, "failed to delete version")
		return
	}
	utils.NoContentResponse(c)
}

type CompareResult struct {
	From          *models.ApplicationVersion `json:"from"`
	To            *models.ApplicationVersion `json:"to"`
	Fixed         []models.Finding           `json:"fixed"`
	New           []models.Finding           `json:"new"`
	StillPresent  []models.Finding           `json:"stillPresent"`
}

func CompareVersions(c *gin.Context) {
	appID := c.Param("id")
	fromQuery := c.Query("from")
	toQuery := c.Query("to")

	if fromQuery == "" || toQuery == "" {
		utils.BadRequestError(c, "query params 'from' and 'to' are required")
		return
	}

	var fromVersion, toVersion models.ApplicationVersion
	for _, v := range []struct {
		dst   *models.ApplicationVersion
		query string
	}{
		{&fromVersion, fromQuery},
		{&toVersion, toQuery},
	} {
		if id, err := strconv.ParseUint(v.query, 10, 64); err == nil {
			config.DB.First(v.dst, id)
		} else {
			config.DB.Where("application_id = ? AND name = ?", appID, v.query).First(v.dst)
		}
		if v.dst.ID == 0 {
			utils.NotFoundError(c, "version not found: "+v.query)
			return
		}
	}

	var fromFindings, toFindings []models.Finding
	config.DB.Where("application_version_id = ?", fromVersion.ID).Find(&fromFindings)
	config.DB.Where("application_version_id = ?", toVersion.ID).Find(&toFindings)

	keyFn := func(f models.Finding) string { return f.RuleID + "|" + f.FilePath }

	toKeys := make(map[string]bool)
	for _, f := range toFindings {
		toKeys[keyFn(f)] = true
	}
	fromKeys := make(map[string]bool)
	for _, f := range fromFindings {
		fromKeys[keyFn(f)] = true
	}

	var fixed, news, stillPresent []models.Finding
	for _, f := range fromFindings {
		if !toKeys[keyFn(f)] {
			fixed = append(fixed, f)
		}
	}
	for _, f := range toFindings {
		if !fromKeys[keyFn(f)] {
			news = append(news, f)
		} else {
			stillPresent = append(stillPresent, f)
		}
	}

	if fixed == nil {
		fixed = []models.Finding{}
	}
	if news == nil {
		news = []models.Finding{}
	}
	if stillPresent == nil {
		stillPresent = []models.Finding{}
	}

	utils.OKResponse(c, CompareResult{
		From:         &fromVersion,
		To:           &toVersion,
		Fixed:        fixed,
		New:          news,
		StillPresent: stillPresent,
	})
}

func GetVersionFindings(c *gin.Context) {
	var version models.ApplicationVersion
	if err := config.DB.First(&version, c.Param("versionId")).Error; err != nil {
		utils.NotFoundError(c, "Version not found")
		return
	}

	var findings []models.Finding
	query := config.DB.Preload("ScannerType").Where("application_version_id = ?", version.ID)

	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&findings).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch findings")
		return
	}
	utils.OKResponse(c, findings)
}
