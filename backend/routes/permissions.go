package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterPermissionRoutes(router *gin.Engine) {
	p := router.Group("/admin/permissions", middleware.AuthRequired(), middleware.CSRFProtection())
	{
		p.GET("", middleware.CheckPolicy("/admin/permissions", "read"), controllers.GetPermissions)
		p.POST("", middleware.CheckPolicy("/admin/permissions", "write"), controllers.GrantPermission)
		p.DELETE("", middleware.CheckPolicy("/admin/permissions", "write"), controllers.RevokePermission)
	}
}
