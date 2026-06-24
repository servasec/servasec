package routes

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func getPublicURL() string {
	u := os.Getenv("NEXT_PUBLIC_API_URL")
	if u == "" {
		u = "http://localhost:8080"
	}
	return u
}

func getMCPResourceURL() string {
	base := getPublicURL()
	prefix := os.Getenv("SSC_API_PREFIX")
	return base + prefix + "/mcp"
}

func RegisterWellKnownRoutes(router *gin.Engine) {
	router.GET("/.well-known/oauth-authorization-server", func(c *gin.Context) {
		base := getPublicURL()
		c.JSON(http.StatusOK, gin.H{
			"issuer":                        base,
			"authorization_endpoint":        base + "/oauth/authorize",
			"token_endpoint":                base + "/oauth/token",
			"registration_endpoint":         base + "/oauth/register",
			"response_types_supported":      []string{"code"},
			"grant_types_supported":         []string{"authorization_code", "refresh_token"},
			"code_challenge_methods_supported": []string{"S256"},
			"token_endpoint_auth_methods_supported": []string{"none"},
			"scopes_supported":              []string{},
		})
	})

	router.GET("/.well-known/oauth-protected-resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"resource": getMCPResourceURL(),
		})
	})
}
