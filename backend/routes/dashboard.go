package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterDashboardRoutes(router *gin.Engine) {
	dashboard := router.Group("/dashboard", middleware.AuthRequired())
	{
		dashboard.GET("/stats", middleware.CheckPolicy("/dashboard", "read"), controllers.GetDashboardStats)
	}
}
