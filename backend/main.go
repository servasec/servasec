package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/features"
	"github.com/servasec/servasec/backend/middleware"
	"github.com/servasec/servasec/backend/services"
	"github.com/servasec/servasec/backend/routes"
	"github.com/servasec/servasec/backend/utils"
)

func getTrustedProxies() []string {
	cidr := os.Getenv("TRUSTED_PROXIES_CIDR")
	if cidr == "" {
		cidr = "172.70.1.0/24"
	}
	return []string{cidr}
}

func main() {
	config.ConnectDB()
	config.InitCasbin()
	utils.ValidateSecrets()
	utils.StartBlacklistCleanup()

	features.Init(os.Getenv("SSC_LICENSE_KEY"))
	debug.Log("Enabled features: %v", features.F.EnabledFeatures())

	if features.F.IsEnabled(features.FeatureRiskScoring) {
		services.StartEPSSSync(services.NewEPSSClient())
	}

	var gReleaseMode string
	if os.Getenv("SSC_DEBUG_ENABLED") == "true" {
		gReleaseMode = gin.DebugMode
	} else {
		gReleaseMode = gin.ReleaseMode
	}
	gin.SetMode(gReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	if os.Getenv("SSC_DEBUG_ENABLED") == "true" {
		router.Use(gin.Recovery())
	}

	middleware.InitCSRFProtection()

	router.SetTrustedProxies(getTrustedProxies())
	allowedOrigin := os.Getenv("NEXT_PUBLIC_API_URL")
	if allowedOrigin == "" {
		log.Fatal("NEXT_PUBLIC_API_URL environment variable is required but not set")
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{allowedOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	if features.F.IsEnabled(features.FeatureAuditLog) {
		router.Use(middleware.AuditLog())
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	routes.RegisterWellKnownRoutes(router)
	routes.RegisterAuthRoutes(router)
	routes.RegisterOAuthRoutes(router)
	routes.RegisterUserRoutes(router)
	routes.RegisterDashboardRoutes(router)
	routes.RegisterGroupRoutes(router)
	routes.RegisterApplicationRoutes(router)
	routes.RegisterScanRoutes(router)
	routes.RegisterFindingRoutes(router)
	routes.RegisterTeamRoutes(router)
	routes.RegisterScannerTypeRoutes(router)
	routes.RegisterPermissionRoutes(router)

	if features.F.IsEnabled(features.FeatureAuditLog) {
		routes.RegisterAuditRoutes(router)
	}

	if features.F.IsEnabled(features.FeatureMCPServer) {
		routes.RegisterMCPRoutes(router)
		debug.Log("MCP server routes registered")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	debug.Log("Starting server on port %s", port)
	router.Run(":" + port)
}
