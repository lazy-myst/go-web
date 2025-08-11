package repository

import (
	"context"

	"github.com/lazy-myst/go-web.git/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateSocketID(ctx context.Context, id, socketID string) error
}

// ChatRepository defines the interface for chat data operations
type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat) error
	FindByID(ctx context.Context, id string) (*models.Chat, error)
	FindByUserID(ctx context.Context, userID string) ([]*models.Chat, error)
	UpdateLastMessage(ctx context.Context, chatID, messageID string) error
}

// MessageRepository defines the interface for message data operations
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	FindByChatID(ctx context.Context, chatID string, limit int) ([]*models.Message, error)
	UpdateStatus(ctx context.Context, messageID, userID, status string) error
}
