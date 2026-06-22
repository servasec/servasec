package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterScannerTypeRoutes(router *gin.Engine) {
	s := router.Group("/scanner-types", middleware.AuthRequired())
	{
		s.GET("", middleware.CheckPolicy("/scanner-types", "read"), controllers.GetScannerTypes)
	}
}
