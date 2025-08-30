package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/lazy-myst/go-web.git/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupMessageRoutes(api fiber.Router, mongoClient *mongo.Client) {
	messages := api.Group("/chats/:chatId/messages", authMiddleware)
	messages.Get("/", fetchMessages(mongoClient))
}

func fetchMessages(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		chatId := c.Params("chatId")
		userId := c.Locals("userId").(string)

		// Check if user is in chat
		chatsColl := mongoClient.Database("chatdb").Collection("chats")
		var chat models.Chat
		if err := chatsColl.FindOne(context.TODO(), bson.M{"_id": chatId}).Decode(&chat); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Chat not found"})
		}

		authorized := false
		for _, id := range chat.UserIDs {
			if id == userId {
				authorized = true
				break
			}
		}
		if !authorized {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not authorized"})
		}

		messagesColl := mongoClient.Database("chatdb").Collection("messages")

		opts := options.Find().SetSort(bson.D{{"createdAt", 1}})
		cursor, err := messagesColl.Find(context.TODO(), bson.M{"chatId": chatId}, opts)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch messages"})
		}
		defer cursor.Close(context.TODO())

		var messages []models.Message
		if err := cursor.All(context.TODO(), &messages); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode messages"})
		}

		return c.JSON(messages)
	}
}
