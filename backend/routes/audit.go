package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterAuditRoutes(router *gin.Engine) {
	a := router.Group("/admin/audit-logs", middleware.AuthRequired(), middleware.CSRFProtection())
	{
		a.GET("", middleware.CheckPolicy("/admin/audit-logs", "read"), controllers.GetAuditLogs)
	}
}
