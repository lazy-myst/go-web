package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI    string
	Port        string
	JWTSecret   string
	CORSOrigins string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return Config{
		MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017/chatdb"),
		Port:      getEnv("PORT", "3000"),
		JWTSecret: getEnv("JWT_SECRET", "secret"),
		// Comma-separated, matches gofiber/cors' AllowOrigins format. Defaults
		// to the local dev origins so nothing changes for `docker compose up`.
		CORSOrigins: getEnv("CORS_ORIGINS", "http://10.0.2.2:8080,http://localhost:8080,http://localhost:5173,http://10.0.2.2:5173"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
