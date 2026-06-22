package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterWebhookRoutes(router *gin.Engine) {
	w := router.Group("/applications", middleware.AuthRequired(), middleware.CSRFProtection())
	{
		w.GET("/:id/webhooks", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.GetWebhooks)
		w.POST("/:id/webhooks", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.CreateWebhook)
		w.DELETE("/:id/webhooks/:webhookId", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.DeleteWebhook)
	}
}
