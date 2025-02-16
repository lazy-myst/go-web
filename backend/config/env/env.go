package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Load the .env file once when the package is imported
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Warning: No .env file found. Using system environment variables.")
	}
}

// GetString fetches a string environment variable with a fallback
func GetString(key, fallback string) string {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return fallback
	}
	return val
}

// GetInt fetches an integer environment variable with a fallback
func GetInt(key string, fallback int) int {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}

// GetBool fetches a boolean environment variable with a fallback
func GetBool(key string, fallback bool) bool {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return boolVal
}
