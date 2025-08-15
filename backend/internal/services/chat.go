package service

import (
	"context"
	"errors"
	"time"

	"github.com/lazy-myst/go-web.git/internal/models"
	"github.com/lazy-myst/go-web.git/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatService handles chat and message logic
type ChatService struct {
	userRepo    repository.UserRepository
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
}

// NewChatService creates a new ChatService
func NewChatService(userRepo repository.UserRepository, chatRepo repository.ChatRepository, messageRepo repository.MessageRepository) *ChatService {
	return &ChatService{
		userRepo:    userRepo,
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

// CreateChat creates a new one-to-one or group chat
func (s *ChatService) CreateChat(ctx context.Context, isGroup bool, name string, userIDs []string) (*models.Chat, error) {
	// Validate users
	oids := make([]primitive.ObjectID, 0, len(userIDs))
	for _, id := range userIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		user, err := s.userRepo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, errors.New("user not found: " + id)
		}
		oids = append(oids, oid)
	}

	// Create chat
	chat := &models.Chat{
		IsGroup: isGroup,
		Name:    name,
		Users:   oids,
	}
	if isGroup {
		chat.Admins = oids[:1] // First user as admin
	}
	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}
	return chat, nil
}

// GetUserChats retrieves all chats for a user
func (s *ChatService) GetUserChats(ctx context.Context, userID string) ([]*models.Chat, error) {
	return s.chatRepo.FindByUserID(ctx, userID)
}

// SendMessage sends a new message in a chat
func (s *ChatService) SendMessage(ctx context.Context, chatID, senderID, content string) (*models.Message, error) {
	// Validate chat and sender
	chatOID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, err
	}
	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat == nil {
		return nil, errors.New("chat not found")
	}
	senderOID, err := primitive.ObjectIDFromHex(senderID)
	if err != nil {
		return nil, err
	}

	// Create message
	message := &models.Message{
		Chat:    chatOID,
		Sender:  senderOID,
		Content: content,
		StatusPerUser: []models.StatusPerUser{
			{User: senderOID, Status: "sent", UpdatedAt: time.Now()},
		},
	}
	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	// Update chat's last message
	if err := s.chatRepo.UpdateLastMessage(ctx, chatID, message.ID.Hex()); err != nil {
		return nil, err
	}
	return message, nil
}

// GetMessages retrieves messages for a chat
func (s *ChatService) GetMessages(ctx context.Context, chatID string, limit int) ([]*models.Message, error) {
	return s.messageRepo.FindByChatID(ctx, chatID, limit)
}

// UpdateMessageStatus updates the message status for a user
func (s *ChatService) UpdateMessageStatus(ctx context.Context, messageID, userID, status string) error {
	return s.messageRepo.UpdateStatus(ctx, messageID, userID, status)
}
