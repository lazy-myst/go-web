package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/lazy-myst/go-web.git/internal/config"
	"github.com/lazy-myst/go-web.git/internal/db"
	"github.com/lazy-myst/go-web.git/internal/handlers"
	"github.com/lazy-myst/go-web.git/internal/socket"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Connect to MongoDB
	mongoClient := db.ConnectMongo(cfg.MongoURI)
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 4 * 1024 * 1024, // 4MB
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://10.0.2.2:8080,http://localhost:8080,http://localhost:5173,http://10.0.2.2:5173", // Add frontend origins
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// Setup routes
	api := app.Group("/api")
	handlers.SetupAuthRoutes(api, mongoClient)
	handlers.SetupChatRoutes(api, mongoClient)
	handlers.SetupMessageRoutes(api, mongoClient)

	// Setup Socket.IO server
	socketServer := socket.SetupSocketServer(mongoClient)
	go func() {
		if err := socketServer.Serve(); err != nil {
			log.Printf("Socket.IO server error: %v", err)
		}
	}()
	defer socketServer.Close()

	// Setup HTTP server with mux to handle both Fiber and Socket.IO
	mux := http.NewServeMux()
	mux.Handle("/socket.io/", socketServer)
	mux.Handle("/", adaptor.FiberApp(app))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	log.Printf("Server running on port %s", cfg.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
