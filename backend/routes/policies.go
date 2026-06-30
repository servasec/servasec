package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterPolicyRoutes(router *gin.Engine) {
	p := router.Group("/policies", middleware.AuthRequired(), middleware.CSRFProtection())
	{
		p.GET("", controllers.GetPolicies)
		p.POST("", controllers.CreatePolicy)
		p.GET("/:id", controllers.GetPolicy)
		p.PUT("/:id", controllers.UpdatePolicy)
		p.DELETE("/:id", controllers.DeletePolicy)
		p.PATCH("/:id/toggle", controllers.TogglePolicy)
		p.GET("/logs", controllers.GetPolicyLogs)
		p.GET("/:id/logs", controllers.GetPolicyLogsByPolicy)
	}
}
