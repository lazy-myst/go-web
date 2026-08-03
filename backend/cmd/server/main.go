package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/lazy-myst/go-web.git/internal/config"
	"github.com/lazy-myst/go-web.git/internal/db"
	"github.com/lazy-myst/go-web.git/internal/grpcchat"
	"github.com/lazy-myst/go-web.git/internal/handlers"
	"github.com/lazy-myst/go-web.git/internal/pb"
	"google.golang.org/grpc"
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
    	AllowOriginsFunc: func(origin string) bool {
    	    // Allow anything explicitly configured (your Tailscale IP, etc.)
    	    for _, o := range strings.Split(cfg.CORSOrigins, ",") {
    	        if strings.TrimSpace(o) == origin {
    	            return true
    	        }
    	    }
    	    // Also allow any cloudflared quick-tunnel URL, since those change
    	    // every run and aren't worth hardcoding.
    	    return strings.HasSuffix(origin, ".trycloudflare.com")
    	},
    	AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    	AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// Setup routes
	api := app.Group("/api")
	handlers.SetupAuthRoutes(api, mongoClient)
	handlers.SetupChatRoutes(api, mongoClient)
	handlers.SetupMessageRoutes(api, mongoClient)

	// Setup gRPC chat service, wrapped for browser access (gRPC-Web)
	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, grpcchat.NewServer(mongoClient))
	wrappedGrpc := grpcweb.WrapServer(grpcServer, grpcweb.WithOriginFunc(func(origin string) bool { return true }))

	fiberHandler := adaptor.FiberApp(app)

	// Setup HTTP server with mux to handle both Fiber (REST) and gRPC-Web (chat)
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		fiberHandler.ServeHTTP(w, r)
	}))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
		// gRPC-Web streaming calls can stay open indefinitely, so these
		// timeouts can't apply server-wide the way they did with plain REST.
		ReadTimeout: 10 * time.Second,
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
