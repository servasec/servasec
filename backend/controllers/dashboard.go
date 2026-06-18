package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

type DashboardStats struct {
	TotalUsers   int64  `json:"totalUsers"`
	AdminUsers   int64  `json:"adminUsers"`
	MemberUsers  int64  `json:"memberUsers"`
	BannedUsers  int64  `json:"bannedUsers"`
	RegisteredAt string `json:"registeredAt"`
}

func GetDashboardStats(c *gin.Context) {
	var stats DashboardStats

	if err := config.DB.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		utils.InternalServerError(c, "Failed to load dashboard stats")
		return
	}
	if err := config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&stats.AdminUsers).Error; err != nil {
		utils.InternalServerError(c, "Failed to load dashboard stats")
		return
	}
	if err := config.DB.Model(&models.User{}).Where("role = ?", "member").Count(&stats.MemberUsers).Error; err != nil {
		utils.InternalServerError(c, "Failed to load dashboard stats")
		return
	}
	if err := config.DB.Model(&models.User{}).Where("banned = ?", true).Count(&stats.BannedUsers).Error; err != nil {
		utils.InternalServerError(c, "Failed to load dashboard stats")
		return
	}

	var firstUser models.User
	if err := config.DB.Order("created_at asc").First(&firstUser).Error; err == nil {
		stats.RegisteredAt = firstUser.CreatedAt.Format("2006-01-02")
	}

	utils.OKResponse(c, stats)
}
