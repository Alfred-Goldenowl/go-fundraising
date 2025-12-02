package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env file not found, using system env variables")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
