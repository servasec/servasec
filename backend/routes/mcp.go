package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/mcp"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterMCPRoutes(router *gin.Engine) {
	mcpGroup := router.Group("/mcp", middleware.AuthRequired(), middleware.CheckPolicy("/mcp", "*"))
	{
		mcpGroup.GET("", mcp.HandleSSE)
		mcpGroup.POST("", mcp.HandleStreamableHTTP)
		mcpGroup.POST("/message", mcp.HandleMessage)
	}
}
