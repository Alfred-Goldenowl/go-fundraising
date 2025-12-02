package main

import (
	"fmt"
	authRouter "go-fundraising/auth/routes"
	campaignRouter "go-fundraising/campaign/routes"
	"go-fundraising/configs"
	paymentRouter "go-fundraising/payment/routes"
	"go-fundraising/worker"

	"go-fundraising/db"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	configs.LoadEnv()
	db.InitScylla()
	db.InitElastic()
	defer db.CloseScylla()

	r := gin.Default()

	worker.InitSyncWorkers(5)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	campaignRouter.InitCommentRouter(r)
	campaignRouter.InitCampaignRouter(r)
	authRouter.InitAuthRouter(r)
	paymentRouter.InitPaymentRouter(r)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	port := configs.GetEnv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	r.LoadHTMLGlob("./payment/templates/*")

	log.Println("🚀 Server is running at port", port)
	r.Run(fmt.Sprintf(":%s", port))
}
