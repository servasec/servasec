package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/controllers"
	"github.com/servasec/servasec/backend/middleware"
)

func RegisterUserRoutes(router *gin.Engine) {
	users := router.Group("/users", middleware.AuthRequired(), middleware.CSRFProtection())
	{
		users.GET("", middleware.CheckPolicy("/users", "read"), controllers.GetUsers)
		users.GET("/:id", middleware.CheckPolicy("/users/:id", "read"), controllers.GetUser)
		users.PUT("/:id", middleware.CheckPolicy("/users/:id", "write"), controllers.UpdateUser)
		users.DELETE("/:id", middleware.CheckPolicy("/users/:id", "write"), controllers.DeleteUser)
	}
}
