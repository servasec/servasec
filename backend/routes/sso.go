package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/sso"
)

func RegisterSSORoutes(router *gin.Engine) {
	// Routes are under /auth because Caddy's handle_path /api/* strips the /api prefix
	ssoRoutes := router.Group("/auth")
	{
		ssoRoutes.GET("/providers", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"providers": sso.EnabledProviderInfos(),
			})
		})
		ssoRoutes.GET("/:provider", sso.HandleAuthRedirect)
		ssoRoutes.GET("/:provider/callback", sso.HandleCallback)
	}
}
