package pro

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SSOProvider interface {
	RegisterRoutes(router *gin.Engine)
}

type noopSSOProvider struct{}

func (n *noopSSOProvider) RegisterRoutes(router *gin.Engine) {
	ssoRoutes := router.Group("/auth")
	{
		ssoRoutes.GET("/providers", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"providers": []interface{}{},
			})
		})
	}
}

var SSO SSOProvider = &noopSSOProvider{}
