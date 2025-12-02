package routes

import (
	"go-fundraising/campaign/handlers"
	"go-fundraising/middleware"

	"github.com/gin-gonic/gin"
)

func InitCommentRouter(route *gin.Engine) {
	commentGroup := route.Group("/comments")
	{
		commentGroup.POST("/:campaign_id", middleware.AuthMiddleware(), handlers.CreateCommentHandler)
		commentGroup.GET("/:campaign_id", handlers.GetCommentsByCampaignIDHandler)
	}
}
