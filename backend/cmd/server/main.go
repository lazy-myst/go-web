package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/lazy-myst/go-web.git/config/env"
	"github.com/lazy-myst/go-web.git/internal/db"
)

func main() {
	// fmt.Println("🚀 Starting Go Fiber Server...")

	// Load .env file manually
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Error loading .env file:", err)
	}

	// Load environment variables
	PORT := env.GetString("PORT", "3001")

	// Initialize MongoDB connection
	client, err := db.ConnectMongoDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(db.Ctx)

	fmt.Println("✅ Connected to MongoDB")

	// Initialize Fiber app
	app := fiber.New()

	// Define a simple health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "OK"})
	})

	// Start the server
	log.Printf("🚀 Server is running on port %s", PORT)
	log.Fatal(app.Listen(":" + PORT))
}
