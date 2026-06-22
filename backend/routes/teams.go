package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterTeamRoutes(router *gin.Engine) {
	t := router.Group("/teams", middleware.AuthRequired(), middleware.CSRFProtection())
	{
		t.GET("", middleware.CheckPolicy("/teams", "read"), controllers.GetTeams)
		t.POST("", middleware.CheckPolicy("/teams/*", "write"), controllers.CreateTeam)
		t.GET("/:id", middleware.RequireResourceAccessByParam("teams", "id", "read"), controllers.GetTeam)
		t.PUT("/:id", middleware.RequireResourceAccessByParam("teams", "id", "write"), controllers.UpdateTeam)
		t.DELETE("/:id", middleware.RequireResourceAccessByParam("teams", "id", "write"), controllers.DeleteTeam)
		t.POST("/:id/members", middleware.RequireResourceAccessByParam("teams", "id", "write"), controllers.AddTeamMember)
		t.DELETE("/:id/members/:userId", middleware.RequireResourceAccessByParam("teams", "id", "write"), controllers.RemoveTeamMember)
	}
}
