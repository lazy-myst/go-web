package grpcchat

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/lazy-myst/go-web.git/internal/config"
	"github.com/lazy-myst/go-web.git/internal/models"
	"github.com/lazy-myst/go-web.git/internal/pb"
	"github.com/lazy-myst/go-web.git/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Server implements pb.ChatServiceServer, replacing the previous
// socket.io-based real-time transport with gRPC (server-streaming for
// receiving messages, unary for sending).
type Server struct {
	pb.UnimplementedChatServiceServer

	mongoClient *mongo.Client
	cfg         config.Config

	mu          sync.Mutex
	subscribers map[string]map[chan *pb.Message]struct{} // userId -> set of listener channels
}

func NewServer(mongoClient *mongo.Client) *Server {
	return &Server{
		mongoClient: mongoClient,
		cfg:         config.LoadConfig(),
		subscribers: make(map[string]map[chan *pb.Message]struct{}),
	}
}

func (s *Server) userIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization token")
	}

	tokenString := strings.TrimPrefix(tokens[0], "Bearer ")
	userId, err := auth.ValidateJWT(tokenString, s.cfg.JWTSecret)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	return userId, nil
}

func (s *Server) subscribe(userId string, ch chan *pb.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subscribers[userId] == nil {
		s.subscribers[userId] = make(map[chan *pb.Message]struct{})
	}
	s.subscribers[userId][ch] = struct{}{}
}

func (s *Server) unsubscribe(userId string, ch chan *pb.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subscribers[userId], ch)
	if len(s.subscribers[userId]) == 0 {
		delete(s.subscribers, userId)
	}
	close(ch)
}

func (s *Server) publish(userId string, msg *pb.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ch := range s.subscribers[userId] {
		select {
		case ch <- msg:
		default:
			// Slow consumer — drop rather than block every other subscriber.
			log.Printf("gRPC: dropping message for user %s, channel full", userId)
		}
	}
}

// StreamMessages pushes every new message for chats the caller belongs to,
// for as long as the client keeps the call open.
func (s *Server) StreamMessages(req *pb.StreamRequest, stream pb.ChatService_StreamMessagesServer) error {
	userId, err := s.userIDFromContext(stream.Context())
	if err != nil {
		return err
	}

	ch := make(chan *pb.Message, 16)
	s.subscribe(userId, ch)
	defer s.unsubscribe(userId, ch)

	log.Printf("gRPC: user %s subscribed to message stream", userId)

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(msg); err != nil {
				return err
			}
		case <-stream.Context().Done():
			log.Printf("gRPC: user %s stream closed: %v", userId, stream.Context().Err())
			return nil
		}
	}
}

// SendMessage validates the caller belongs to the chat, persists the message,
// then fans it out to every participant's active StreamMessages call.
func (s *Server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	userId, err := s.userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.ChatId == "" || req.Text == "" {
		return nil, status.Error(codes.InvalidArgument, "chatId and text are required")
	}

	chatsColl := s.mongoClient.Database("chatdb").Collection("chats")
	var chat models.Chat
	if err := chatsColl.FindOne(ctx, bson.M{"_id": req.ChatId}).Decode(&chat); err != nil {
		return nil, status.Error(codes.NotFound, "chat not found")
	}

	authorized := false
	for _, id := range chat.UserIDs {
		if id == userId {
			authorized = true
			break
		}
	}
	if !authorized {
		return nil, status.Error(codes.PermissionDenied, "not authorized for this chat")
	}

	message := models.Message{
		ID:        auth.NewObjectIDHex(),
		ChatID:    req.ChatId,
		SenderID:  userId,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}

	messagesColl := s.mongoClient.Database("chatdb").Collection("messages")
	if _, err := messagesColl.InsertOne(ctx, message); err != nil {
		return nil, status.Error(codes.Internal, "failed to save message")
	}

	pbMsg := &pb.Message{
		Id:        message.ID,
		ChatId:    message.ChatID,
		SenderId:  message.SenderID,
		Text:      message.Text,
		CreatedAt: message.CreatedAt.Format(time.RFC3339),
	}

	log.Printf("gRPC: broadcasting message for chat %s to %d participant(s)", req.ChatId, len(chat.UserIDs))
	for _, participantId := range chat.UserIDs {
		s.publish(participantId, pbMsg)
	}

	return pbMsg, nil
}
