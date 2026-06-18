package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterGroupRoutes(router *gin.Engine) {
	g := router.Group("/groups", middleware.AuthRequired(), middleware.CSRFProtection())
	{
	g.GET("", middleware.CheckPolicy("/groups", "read"), controllers.GetGroups)
	g.POST("", middleware.CheckPolicy("/groups/*", "write"), controllers.CreateGroup)
	g.GET("/:id", middleware.CheckPolicy("/groups/*", "read"), middleware.RequireResourceAccessByParam("groups", "id", "read"), controllers.GetGroup)
	g.PUT("/:id", middleware.CheckPolicy("/groups/*", "write"), middleware.RequireResourceAccessByParam("groups", "id", "write"), controllers.UpdateGroup)
	g.DELETE("/:id", middleware.CheckPolicy("/groups/*", "write"), middleware.RequireResourceAccessByParam("groups", "id", "write"), controllers.DeleteGroup)
	}
}
