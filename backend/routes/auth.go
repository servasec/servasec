package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterAuthRoutes(router *gin.Engine) {
	auth := router.Group("/")
	{
		auth.POST("login", middleware.RateLimitLogin(), middleware.CSRFProtection(), controllers.Login)
		auth.POST("register", middleware.RateLimit(5), middleware.CSRFProtection(), controllers.Register)
		auth.POST("logout", middleware.AuthRequired(), middleware.CSRFProtection(), controllers.Logout)

		auth.POST("refresh", middleware.RateLimit(10), middleware.CSRFProtection(), controllers.Refresh)
		auth.GET("me", middleware.AuthRequired(), controllers.GetCurrentUser)
		auth.GET("csrf-token", middleware.RateLimit(30), middleware.CSRFProtection(), controllers.GetCSRFToken)

		auth.PATCH("me", middleware.AuthRequired(), middleware.CSRFProtection(), controllers.UpdateCurrentUser)
		auth.PATCH("me/onboarding", middleware.AuthRequired(), middleware.CSRFProtection(), controllers.MarkOnboardingSeen)
		auth.PUT("me/password", middleware.AuthRequired(), middleware.CSRFProtection(), controllers.UpdateCurrentUserPassword)
	}
}
