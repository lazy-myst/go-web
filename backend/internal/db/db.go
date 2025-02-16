package db

import (
	"context"
	"fmt"
	"log"

	"github.com/lazy-myst/go-web.git/config/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variables
var Ctx = context.Background()
var Client *mongo.Client

// ConnectMongoDB establishes a connection to MongoDB and returns a client instance
func ConnectMongoDB() (*mongo.Client, error) {
	MONGO_URI := env.GetString("MONGO_URI", "")

	if MONGO_URI == "" {
		log.Fatal("❌ MONGO_URI is not set in environment variables")
	}

	// Explicitly check if the scheme is correct
	if len(MONGO_URI) < 10 {
		log.Fatal("❌ MONGO_URI seems invalid (too short)")
	}

	clientOptions := options.Client().ApplyURI(MONGO_URI)
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(Ctx, nil); err != nil {
		return nil, err
	}

	fmt.Println("✅ MongoDB connection established")
	Client = client
	return client, nil
}
