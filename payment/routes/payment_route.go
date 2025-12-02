package routes

import (
	"go-fundraising/middleware"
	handlers "go-fundraising/payment/handler"

	"github.com/gin-gonic/gin"
)

func InitPaymentRouter(route *gin.Engine) {
	paymentGroup := route.Group("/payment")
	{
		paymentGroup.POST("/create", middleware.AuthMiddleware(), handlers.CreatePaymentIntentHandler)
		paymentGroup.GET("/success", handlers.PaymentSuccessHandler)
		paymentGroup.GET("/fail", handlers.PaymentFailHandler)
	}
}
