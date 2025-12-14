package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("[go-middle] .env not found, using system environment")
	}
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("[go-middle] WARNING: env %s not set\n", key)
	}
	return value
}
