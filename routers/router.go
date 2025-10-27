package routes

import (
	"github.com/gin-gonic/gin"
	"go-fundraising/handlers"
)

func InitRouter(router *gin.Engine) {
	commentGroup := router.Group("/comments")
	{
		commentGroup.POST("/", handlers.CreateCommentHandler)
		commentGroup.GET("/:post_id", handlers.GetCommentsByPostIDHandler)
	}
}
