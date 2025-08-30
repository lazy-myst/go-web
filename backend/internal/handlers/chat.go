package handlers

import (
	"context"
	"time"

	// "chatapp/models"

	"github.com/gofiber/fiber/v2"
	"github.com/lazy-myst/go-web.git/internal/models"
	"github.com/lazy-myst/go-web.git/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupChatRoutes(api fiber.Router, mongoClient *mongo.Client) {
	chats := api.Group("/chats", authMiddleware)
	chats.Post("/", createChat(mongoClient))
	chats.Get("/", listChats(mongoClient))
}

func createChat(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Locals("userId").(string)

		var req struct {
			UserIDs []string `json:"userIds"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Ensure current user is included
		includesSelf := false
		for _, id := range req.UserIDs {
			if id == userId {
				includesSelf = true
				break
			}
		}
		if !includesSelf {
			req.UserIDs = append(req.UserIDs, userId)
		}

		chatsColl := mongoClient.Database("chatdb").Collection("chats")

		chat := models.Chat{
			ID:        auth.NewObjectIDHex(),
			UserIDs:   req.UserIDs,
			CreatedAt: time.Now(),
		}

		_, err := chatsColl.InsertOne(context.TODO(), chat)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create chat"})
		}

		return c.JSON(chat)
	}
}

func listChats(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Locals("userId").(string)

		chatsColl := mongoClient.Database("chatdb").Collection("chats")

		opts := options.Find().SetSort(bson.D{{"createdAt", -1}})
		cursor, err := chatsColl.Find(context.TODO(), bson.M{"userIds": userId}, opts)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch chats"})
		}
		defer cursor.Close(context.TODO())

		var chats []models.Chat
		if err := cursor.All(context.TODO(), &chats); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode chats"})
		}

		return c.JSON(chats)
	}
}
