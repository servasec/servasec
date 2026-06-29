package controllers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/services"
	"github.com/servasec/servasec/backend/utils"
)

func buildFindingsQuery(c *gin.Context) *gorm.DB {
	q := config.DB.Model(&models.Finding{}).
		Joins("JOIN application_versions ON application_versions.id = findings.application_version_id")

	accessibleIDs := utils.GetAccessibleAppIDs(c)
	if accessibleIDs != nil {
		if len(accessibleIDs) == 0 {
			return q.Where("1 = 0")
		}
		q = q.Where("application_versions.application_id IN ?", accessibleIDs)
	}

	if appID := c.Query("applicationId"); appID != "" {
		q = q.Where("application_versions.application_id = ?", appID)
	}
	if versionID := c.Query("applicationVersionId"); versionID != "" {
		q = q.Where("findings.application_version_id = ?", versionID)
	}
	if severity := c.Query("severity"); severity != "" {
		q = q.Where("findings.severity = ?", strings.ToLower(severity))
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("findings.status = ?", status)
	}
	if scannerTypeID := c.Query("scannerTypeId"); scannerTypeID != "" {
		q = q.Where("findings.scanner_type_id = ?", scannerTypeID)
	}
	if scanID := c.Query("scanId"); scanID != "" {
		q = q.Where("findings.scan_id = ?", scanID)
	}
	if c.Query("assignedTo") == "me" {
		userID, _ := c.Get("user_id")
		q = q.Where("findings.assigned_to = ?", userID)
	}
	return q
}

func GetFindings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}
	offset := (page - 1) * perPage

	base := buildFindingsQuery(c)

	sortBy := c.Query("sortBy")
	sortOrder := c.DefaultQuery("order", "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	orderClause := "findings.created_at DESC"
	switch sortBy {
	case "risk_score":
		orderClause = "findings.risk_score " + sortOrder + " NULLS LAST"
	case "severity":
		orderClause = "findings.severity " + sortOrder
	case "title":
		orderClause = "findings.title " + sortOrder
	case "created_at":
		orderClause = "findings.created_at " + sortOrder
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		utils.InternalServerError(c, "failed to count findings")
		return
	}

	var findings []models.Finding
	if err := base.
		Session(&gorm.Session{}).
		Preload("ScannerType").
		Preload("ApplicationVersion").
		Preload("AssignedToUser").
		Preload("ReviewedByUser").
		Order(orderClause).
		Offset(offset).
		Limit(perPage).
		Find(&findings).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch findings")
		return
	}

	if findings == nil {
		findings = []models.Finding{}
	}

	utils.OKResponse(c, gin.H{
		"data":    findings,
		"total":   total,
		"page":    page,
		"perPage": perPage,
	})
}

func GetFinding(c *gin.Context) {
	var finding models.Finding
	if err := config.DB.
		Preload("ScannerType").
		Preload("ApplicationVersion").
		Preload("AssignedToUser").
		Preload("ReviewedByUser").
		First(&finding, c.Param("id")).Error; err != nil {
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
		Status string `json:"status" binding:"required,oneof=open confirmed false_positive fixed closed"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid status")
		return
	}

	finding.Status = input.Status
	if input.Status == "fixed" && finding.FixedAt == nil {
		now := time.Now()
		finding.FixedAt = &now
	}
	if err := config.DB.Save(&finding).Error; err != nil {
		utils.InternalServerError(c, "failed to update finding status")
		return
	}

	if app := findingApp(finding.ID); app != nil {
		services.EvaluatePolicies("finding.status_changed", &finding, app, 0)
	}

	utils.OKResponse(c, finding)
}

func AssignFinding(c *gin.Context) {
	var finding models.Finding
	if err := config.DB.First(&finding, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Finding not found")
		return
	}

	var input struct {
		UserID  *uint      `json:"userId"`
		DueDate *time.Time `json:"dueDate"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.UserID != nil {
		var user models.User
		if err := config.DB.First(&user, *input.UserID).Error; err != nil {
			utils.NotFoundError(c, "User not found")
			return
		}
		finding.AssignedTo = input.UserID
	} else {
		finding.AssignedTo = nil
	}
	finding.DueDate = input.DueDate

	if err := config.DB.Save(&finding).Error; err != nil {
		utils.InternalServerError(c, "failed to assign finding")
		return
	}

	if app := findingApp(finding.ID); app != nil {
		services.EvaluatePolicies("finding.reassigned", &finding, app, 0)
	}

	utils.OKResponse(c, finding)
}

func ReviewFinding(c *gin.Context) {
	var finding models.Finding
	if err := config.DB.First(&finding, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Finding not found")
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	var input struct {
		Status *string `json:"status" binding:"omitempty,oneof=fixed closed open"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	finding.ReviewedBy = &uid
	if input.Status != nil {
		finding.Status = *input.Status
		if *input.Status == "fixed" && finding.FixedAt == nil {
			now := time.Now()
			finding.FixedAt = &now
		}
	}

	if err := config.DB.Save(&finding).Error; err != nil {
		utils.InternalServerError(c, "failed to review finding")
		return
	}

	if app := findingApp(finding.ID); app != nil {
		services.EvaluatePolicies("finding.status_changed", &finding, app, 0)
	}

	config.DB.Preload("ReviewedByUser").First(&finding, finding.ID)
	utils.OKResponse(c, finding)
}

func CreateComment(c *gin.Context) {
	var finding models.Finding
	if err := config.DB.First(&finding, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Finding not found")
		return
	}

	var input struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Comment body is required")
		return
	}

	userID, _ := c.Get("user_id")

	comment := models.Comment{
		FindingID: finding.ID,
		UserID:    userID.(uint),
		Body:      input.Body,
	}
	if err := config.DB.Create(&comment).Error; err != nil {
		utils.InternalServerError(c, "failed to create comment")
		return
	}
	config.DB.Preload("User").First(&comment, comment.ID)
	utils.CreatedResponse(c, comment)
}

func findingApp(findingID uint) *models.Application {
	var finding models.Finding
	if err := config.DB.Preload("ApplicationVersion").First(&finding, findingID).Error; err != nil {
		return nil
	}
	var app models.Application
	if err := config.DB.First(&app, finding.ApplicationVersion.ApplicationID).Error; err != nil {
		return nil
	}
	return &app
}

func GetComments(c *gin.Context) {
	var finding models.Finding
	if err := config.DB.First(&finding, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Finding not found")
		return
	}

	var comments []models.Comment
	if err := config.DB.Preload("User").
		Where("finding_id = ?", finding.ID).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		utils.InternalServerError(c, "failed to fetch comments")
		return
	}
	utils.OKResponse(c, comments)
}
