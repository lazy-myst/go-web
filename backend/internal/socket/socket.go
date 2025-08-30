package socket

import (
	"context"
	"log"
	"net/url"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/lazy-myst/go-web.git/internal/config"
	"github.com/lazy-myst/go-web.git/internal/models"
	"github.com/lazy-myst/go-web.git/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var cfg = config.LoadConfig()

func SetupSocketServer(mongoClient *mongo.Client) *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		// Parse query parameters from URL
		query, err := url.ParseQuery(s.URL().RawQuery)
		if err != nil {
			log.Printf("Failed to parse query: %v", err)
			s.Close()
			return nil
		}
		tokenString := query.Get("token")
		if tokenString == "" {
			log.Println("No token provided")
			s.Close()
			return nil
		}

		userId, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			log.Printf("Invalid token: %v", err)
			s.Close()
			return nil
		}

		s.SetContext(userId)
		s.Join("user:" + userId)

		// Join chat rooms
		chatsColl := mongoClient.Database("chatdb").Collection("chats")
		cursor, err := chatsColl.Find(context.TODO(), bson.M{"userIds": userId})
		if err != nil {
			log.Printf("Failed to fetch chats for socket: %v", err)
			return nil
		}
		defer cursor.Close(context.TODO())

		var chat models.Chat
		for cursor.Next(context.TODO()) {
			if err := cursor.Decode(&chat); err != nil {
				continue
			}
			s.Join("chat:" + chat.ID)
		}

		return nil
	})

	server.OnEvent("/", "newMessage", func(s socketio.Conn, data map[string]interface{}) {
		userId := s.Context().(string)
		chatId, ok := data["chatId"].(string)
		if !ok {
			log.Println("Invalid chatId")
			return
		}
		text, ok := data["text"].(string)
		if !ok {
			log.Println("Invalid text")
			return
		}

		chatsColl := mongoClient.Database("chatdb").Collection("chats")
		var chat models.Chat
		if err := chatsColl.FindOne(context.TODO(), bson.M{"_id": chatId}).Decode(&chat); err != nil {
			log.Printf("Chat not found: %v", err)
			return
		}

		authorized := false
		for _, id := range chat.UserIDs {
			if id == userId {
				authorized = true
				break
			}
		}
		if !authorized {
			log.Println("User not authorized for chat")
			return
		}

		messagesColl := mongoClient.Database("chatdb").Collection("messages")

		message := models.Message{
			ID:        auth.NewObjectIDHex(),
			ChatID:    chatId,
			SenderID:  userId,
			Text:      text,
			CreatedAt: time.Now(),
		}

		_, err := messagesColl.InsertOne(context.TODO(), message)
		if err != nil {
			log.Printf("Failed to insert message: %v", err)
			return
		}

		// Use BroadcastToRoom instead of BroadcastTo
		server.BroadcastToRoom("", "chat:"+chatId, "messageCreated", message)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Printf("Socket error: %v", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("Disconnected: %s", reason)
	})

	return server
}
