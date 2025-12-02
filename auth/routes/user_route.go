package routes

import (
	"go-fundraising/auth/handlers"
	"go-fundraising/middleware"

	"github.com/gin-gonic/gin"
)

func InitAuthRouter(route *gin.Engine) {
	userGroup := route.Group("/auth")
	{
		userGroup.POST("/register", handlers.CreateUserHandler)
		userGroup.POST("/login", handlers.LoginHandler)
		userGroup.GET("/current-user", middleware.AuthMiddleware(), handlers.CurrentUserHandler)
		userGroup.POST("/refresh", handlers.RefreshHandler)
		userGroup.POST("/logout", handlers.LogoutHandler)
	}
}
