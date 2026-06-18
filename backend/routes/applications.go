package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterApplicationRoutes(router *gin.Engine) {
	a := router.Group("/applications", middleware.AuthRequired(), middleware.CSRFProtection())
	{
	a.GET("", middleware.CheckPolicy("/applications", "read"), controllers.GetApplications)
	a.POST("", middleware.CheckPolicy("/applications/*", "write"), controllers.CreateApplication)
	a.GET("/:id", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.GetApplication)
	a.PUT("/:id", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.UpdateApplication)
	a.DELETE("/:id", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.DeleteApplication)
	a.POST("/:id/regenerate-token", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.RegenerateApiToken)
	}
}
