package socket

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"github.com/lazy-myst/go-web.git/internal/config"
	"github.com/lazy-myst/go-web.git/internal/models"
	"github.com/lazy-myst/go-web.git/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var cfg = config.LoadConfig()

func SetupSocketServer(mongoClient *mongo.Client) *socketio.Server {
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			polling.Default,
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					// Socket.IO requests are handled outside the Fiber app's CORS
					// middleware (mounted directly on the http.ServeMux), so the
					// websocket upgrade needs its own origin allow-list.
					origin := r.Header.Get("Origin")
					switch origin {
					case "http://localhost:5173", "http://10.0.2.2:5173":
						return true
					default:
						return false
					}
				},
			},
		},
	})

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Printf("Socket connection attempt, ID: %s, URL: %s, RemoteAddr: %s", s.ID(), s.URL(), s.RemoteAddr())
		// Parse query parameters from URL
		query, err := url.ParseQuery(s.URL().RawQuery)
		if err != nil {
			log.Printf("Failed to parse query for socket %s: %v", s.ID(), err)
			s.Close()
			return nil
		}
		tokenString := query.Get("token")
		log.Printf("Socket %s: Received token: %s", s.ID(), tokenString)
		if tokenString == "" {
			log.Printf("Socket %s: No token provided", s.ID())
			s.Close()
			return nil
		}

		userId, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			log.Printf("Socket %s: Invalid token: %v", s.ID(), err)
			s.Close()
			return nil
		}

		log.Printf("Socket %s: Authenticated user ID: %s", s.ID(), userId)
		s.SetContext(userId)
		s.Join("user:" + userId)

		// Join chat rooms
		chatsColl := mongoClient.Database("chatdb").Collection("chats")
		cursor, err := chatsColl.Find(context.TODO(), bson.M{"userIds": userId})
		if err != nil {
			log.Printf("Socket %s: Failed to fetch chats: %v", s.ID(), err)
			s.Close()
			return nil
		}
		defer cursor.Close(context.TODO())

		var chat models.Chat
		for cursor.Next(context.TODO()) {
			if err := cursor.Decode(&chat); err != nil {
				log.Printf("Socket %s: Failed to decode chat: %v", s.ID(), err)
				continue
			}
			log.Printf("Socket %s: Joining room chat:%s for user %s", s.ID(), chat.ID, userId)
			s.Join("chat:" + chat.ID)
		}

		log.Printf("Socket %s: Connected successfully", s.ID())
		return nil
	})

	server.OnEvent("/", "newMessage", func(s socketio.Conn, data map[string]interface{}) {
		log.Printf("Socket %s: Received newMessage event: %v", s.ID(), data)
		userId := s.Context().(string)
		chatId, ok := data["chatId"].(string)
		if !ok {
			log.Printf("Socket %s: Invalid chatId", s.ID())
			return
		}
		text, ok := data["text"].(string)
		if !ok {
			log.Printf("Socket %s: Invalid text", s.ID())
			return
		}

		log.Printf("Socket %s: Processing newMessage: chatId=%s, text=%s, userId=%s", s.ID(), chatId, text, userId)
		chatsColl := mongoClient.Database("chatdb").Collection("chats")
		var chat models.Chat
		if err := chatsColl.FindOne(context.TODO(), bson.M{"_id": chatId}).Decode(&chat); err != nil {
			log.Printf("Socket %s: Chat not found: %v", s.ID(), err)
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
			log.Printf("Socket %s: User not authorized for chat", s.ID())
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
			log.Printf("Socket %s: Failed to insert message: %v", s.ID(), err)
			return
		}

		// Broadcast to each participant's personal room rather than the
		// chat's room: "chat:"+chatId is only joined by sockets that were
		// already connected when the chat existed (see OnConnect), so a
		// brand-new chat's other participant would otherwise never see the
		// first message arrive. "user:"+id is joined unconditionally on
		// every connect, so it always reaches anyone currently online.
		log.Printf("Socket %s: Broadcasting messageCreated for chat %s to %d participant(s)", s.ID(), chatId, len(chat.UserIDs))
		for _, id := range chat.UserIDs {
			server.BroadcastToRoom("", "user:"+id, "messageCreated", message)
		}
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		// go-socket.io passes a nil Conn here when OnConnect itself returns
		// a non-nil error (see server.serveConn), so this must stay nil-safe.
		if s == nil {
			log.Printf("Socket error: %v", e)
			return
		}
		log.Printf("Socket %s: Error: %v", s.ID(), e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("Socket %s: Disconnected, reason: %s", s.ID(), reason)
	})

	return server
}
