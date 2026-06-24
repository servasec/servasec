package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterScanRoutes(router *gin.Engine) {
	s := router.Group("/scans", middleware.AuthRequired())
	{
		s.GET("", middleware.CheckPolicy("/scans", "read"), controllers.GetScans)
		s.GET("/:id", middleware.CheckPolicy("/scans/*", "read"), middleware.RequireScanAccess("read"), controllers.GetScan)
	}
}
