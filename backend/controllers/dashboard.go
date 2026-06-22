package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

type ScannerCount struct {
	ScannerType string `json:"scannerType"`
	Count       int64  `json:"count"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type TopFinding struct {
	RuleID string `json:"ruleId"`
	Title  string `json:"title"`
	Count  int64  `json:"count"`
}

type DashboardStats struct {
	TotalUsers      int64          `json:"totalUsers"`
	AdminUsers      int64          `json:"adminUsers"`
	MemberUsers     int64          `json:"memberUsers"`
	BannedUsers     int64          `json:"bannedUsers"`
	RegisteredAt    string         `json:"registeredAt"`
	TotalFindings   int64          `json:"totalFindings"`
	BySeverity      []SeverityCount `json:"bySeverity"`
	ByScanner       []ScannerCount  `json:"byScanner"`
	ByStatus        []StatusCount   `json:"byStatus"`
	RecentScans     int64          `json:"recentScans"`
	TopFindings     []TopFinding   `json:"topFindings"`
	MyOpenFindings  int64          `json:"myOpenFindings"`
	OverdueFindings int64          `json:"overdueFindings"`
}

func GetDashboardStats(c *gin.Context) {
	var stats DashboardStats

	config.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&stats.AdminUsers)
	config.DB.Model(&models.User{}).Where("role = ?", "member").Count(&stats.MemberUsers)
	config.DB.Model(&models.User{}).Where("banned = ?", true).Count(&stats.BannedUsers)

	var firstUser models.User
	if err := config.DB.Order("created_at asc").First(&firstUser).Error; err == nil {
		stats.RegisteredAt = firstUser.CreatedAt.Format("2006-01-02")
	}

	config.DB.Model(&models.Finding{}).Count(&stats.TotalFindings)

	config.DB.Model(&models.Finding{}).
		Select("severity, count(*) as count").
		Group("severity").
		Order("count DESC").
		Find(&stats.BySeverity)

	config.DB.Model(&models.Finding{}).
		Joins("JOIN scanner_types ON scanner_types.id = findings.scanner_type_id").
		Select("scanner_types.name as scanner_type, count(*) as count").
		Group("scanner_types.name").
		Find(&stats.ByScanner)

	config.DB.Model(&models.Finding{}).
		Select("status, count(*) as count").
		Group("status").
		Order("count DESC").
		Find(&stats.ByStatus)

	config.DB.Model(&models.Scan{}).
		Where("status = ?", "completed").
		Count(&stats.RecentScans)

	config.DB.Model(&models.Finding{}).
		Select("rule_id, title, count(*) as count").
		Group("rule_id, title").
		Order("count DESC").
		Limit(5).
		Find(&stats.TopFindings)

	userID, _ := c.Get("user_id")
	config.DB.Model(&models.Finding{}).
		Where("assigned_to = ? AND status IN ?", userID, []string{"open", "confirmed"}).
		Count(&stats.MyOpenFindings)

	config.DB.Raw(`SELECT COUNT(*) FROM findings 
		WHERE due_date IS NOT NULL 
		AND due_date < NOW() 
		AND status NOT IN ('fixed', 'closed', 'false_positive')`).
		Scan(&stats.OverdueFindings)

	utils.OKResponse(c, stats)
}
