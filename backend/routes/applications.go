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

		a.GET("/:id/versions", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.GetVersions)
		a.POST("/:id/versions", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.CreateVersion)
		a.GET("/:id/versions/compare", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.CompareVersions)
		a.GET("/:id/versions/:versionId", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.GetVersion)
		a.PATCH("/:id/versions/:versionId", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.UpdateVersion)
		a.DELETE("/:id/versions/:versionId", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.DeleteVersion)
		a.GET("/:id/versions/:versionId/findings", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.GetVersionFindings)

		a.POST("/:id/ingest", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.IngestScan)
		a.GET("/:id/webhooks", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.GetWebhooks)
		a.POST("/:id/webhooks", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.CreateWebhook)
		a.DELETE("/:id/webhooks/:webhookId", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.DeleteWebhook)

		a.GET("/:id/permissions", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireResourceAccessByParam("applications", "id", "read"), controllers.GetApplicationPermissions)
		a.POST("/:id/permissions", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.GrantApplicationPermission)
		a.DELETE("/:id/permissions", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireResourceAccessByParam("applications", "id", "write"), controllers.RevokeApplicationPermission)
	}

	slug := router.Group("/applications/by-slug/:slug", middleware.AuthRequired(), middleware.CSRFProtection())
	slug.Use(func(c *gin.Context) {
		c.Params = append(c.Params, gin.Param{Key: "id", Value: c.Param("slug")})
		c.Next()
	})
	{
		slug.GET("", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireSlugAccess("read"), controllers.GetApplication)
		slug.POST("/ingest", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireSlugAccess("write"), controllers.IngestScan)
		slug.GET("/versions", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireSlugAccess("read"), controllers.GetVersions)
		slug.POST("/versions", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireSlugAccess("write"), controllers.CreateVersion)
		slug.GET("/versions/compare", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireSlugAccess("read"), controllers.CompareVersions)
		slug.GET("/versions/:versionId", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireSlugAccess("read"), controllers.GetVersion)
		slug.PATCH("/versions/:versionId", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireSlugAccess("write"), controllers.UpdateVersion)
		slug.DELETE("/versions/:versionId", middleware.CheckPolicy("/applications/*", "write"), middleware.RequireSlugAccess("write"), controllers.DeleteVersion)
		slug.GET("/versions/:versionId/findings", middleware.CheckPolicy("/applications/*", "read"), middleware.RequireSlugAccess("read"), controllers.GetVersionFindings)
	}

	publicIngest := router.Group("/")
	{
		publicIngest.POST("/ingest", controllers.IngestScan)
	}
}
