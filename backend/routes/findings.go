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
		f.GET("/:id", middleware.CheckPolicy("/findings/*", "read"), middleware.RequireFindingAccess("read"), controllers.GetFinding)
		f.PATCH("/:id/status", middleware.CheckPolicy("/findings/*", "write"), middleware.RequireFindingAccess("write"), controllers.UpdateFindingStatus)
		f.PATCH("/:id/assign", middleware.CheckPolicy("/findings/*", "write"), middleware.RequireFindingAccess("write"), controllers.AssignFinding)
		f.PATCH("/:id/review", middleware.CheckPolicy("/findings/*", "write"), middleware.RequireFindingAccess("write"), controllers.ReviewFinding)
		f.POST("/:id/comments", middleware.CheckPolicy("/findings/*", "write"), middleware.RequireFindingAccess("write"), controllers.CreateComment)
		f.GET("/:id/comments", middleware.CheckPolicy("/findings/*", "read"), middleware.RequireFindingAccess("read"), controllers.GetComments)
	}
}
