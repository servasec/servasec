package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterFindingRoutes(router *gin.Engine) {
	f := router.Group("/findings", middleware.AuthRequired())
	{
		f.GET("", middleware.CheckPolicy("/findings", "read"), controllers.GetFindings)
		f.GET("/:id", middleware.CheckPolicy("/findings/*", "read"), controllers.GetFinding)
		f.PATCH("/:id/status", middleware.CheckPolicy("/findings/*", "write"), controllers.UpdateFindingStatus)
	}
}
