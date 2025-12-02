package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go-fundraising/db"
	"go-fundraising/routers"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Not found file .env")
	}

	db.InitScylla()
	defer db.CloseScylla()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	routes.InitRouter(r)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🚀 Server is running at port", port)
	r.Run(fmt.Sprintf(":%s", port))
}
