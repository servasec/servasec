// @title Servasec API
// @version 1.0
// @description Security scanner management platform API
// @contact.name API Support
// @contact.email support@servasec.local
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in cookie
// @name access_token
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/servasec/servasec/backend/config"
	_ "github.com/servasec/servasec/backend/docs"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/features"
	"github.com/servasec/servasec/backend/middleware"
	"github.com/servasec/servasec/backend/pro"
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

	pro.Risk.StartEPSSSync()

	var gReleaseMode string
	if os.Getenv("SSC_DEBUG_ENABLED") == "true" {
		gReleaseMode = gin.DebugMode
	} else {
		gReleaseMode = gin.ReleaseMode
	}
	gin.SetMode(gReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	middleware.InitCSRFProtection()

	router.SetTrustedProxies(getTrustedProxies())
	allowedOrigin := os.Getenv("SSC_PUBLIC_URL")
	if allowedOrigin == "" {
		log.Fatal("SSC_PUBLIC_URL environment variable is required but not set")
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{allowedOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.Use(pro.Audit.Middleware())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := config.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(503, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.RegisterWellKnownRoutes(router)
	routes.RegisterAuthRoutes(router)
	pro.SSO.RegisterRoutes(router)
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
	routes.RegisterPolicyRoutes(router)
	routes.RegisterUserApiKeyRoutes(router)

	pro.Audit.RegisterRoutes(router)
	pro.MCP.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		debug.Log("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
