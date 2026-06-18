package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/middleware"
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

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	routes.RegisterAuthRoutes(router)
	routes.RegisterUserRoutes(router)
	routes.RegisterDashboardRoutes(router)
	routes.RegisterGroupRoutes(router)
	routes.RegisterApplicationRoutes(router)
	routes.RegisterScanRoutes(router)
	routes.RegisterFindingRoutes(router)
	routes.RegisterTeamRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	debug.Log("Starting server on port %s", port)
	router.Run(":" + port)
}
