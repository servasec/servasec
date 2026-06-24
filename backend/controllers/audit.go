package controllers

import (
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetAuditLogs(c *gin.Context) {
	query := config.DB.Model(&models.AuditLog{}).Order("created_at DESC")

	if v := c.Query("userId"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			query = query.Where("user_id = ?", id)
		}
	}
	if v := c.Query("action"); v != "" {
		query = query.Where("action = ?", v)
	}
	if v := c.Query("resource"); v != "" {
		decoded, err := url.QueryUnescape(v)
		if err == nil {
			query = query.Where("resource ILIKE ?", "%"+decoded+"%")
		}
	}
	if v := c.Query("status"); v != "" {
		if s, err := strconv.Atoi(v); err == nil {
			query = query.Where("status = ?", s)
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "50"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	var total int64
	query.Count(&total)

	var logs []models.AuditLog
	query.Offset((page - 1) * perPage).Limit(perPage).Find(&logs)

	utils.OKResponse(c, gin.H{
		"data":       logs,
		"total":      total,
		"page":       page,
		"perPage":    perPage,
		"totalPages": (int(total) + perPage - 1) / perPage,
	})
}
