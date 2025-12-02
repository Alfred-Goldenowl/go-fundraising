package routes

import (
	"go-fundraising/campaign/handlers"
	"go-fundraising/middleware"

	"github.com/gin-gonic/gin"
)

func InitCampaignRouter(route *gin.Engine) {
	campaignGroup := route.Group("/campaign")
	{
		campaignGroup.POST("", middleware.AuthMiddleware(), handlers.CreateCampaignHandler)
		campaignGroup.GET("", handlers.SearchCampaignHandler)
		campaignGroup.GET("/:campaign_id", handlers.GetCampaignHandler)
	}
}
