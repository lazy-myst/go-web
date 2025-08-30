package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI  string
	Port      string
	JWTSecret string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return Config{
		MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017/chatdb"),
		Port:      getEnv("PORT", "3000"),
		JWTSecret: getEnv("JWT_SECRET", "secret"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
