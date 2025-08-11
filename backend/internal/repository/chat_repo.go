package repository

import (
	"context"

	"github.com/lazy-myst/go-web.git/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoChatRepository implements ChatRepository for MongoDB
type MongoChatRepository struct {
	collection *mongo.Collection
}

// NewMongoChatRepository creates a new MongoChatRepository
func NewMongoChatRepository(db *mongo.Database) *MongoChatRepository {
	collection := db.Collection("chats")
	// Create indexes
	for _, idx := range models.ChatIndexes() {
		collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.D{{Key: idx, Value: 1}},
		})
	}
	return &MongoChatRepository{collection: collection}
}

// Create inserts a new chat into the database
func (r *MongoChatRepository) Create(ctx context.Context, chat *models.Chat) error {
	chat.ID = primitive.NewObjectID()
	chat.CreatedAt = chat.ID.Timestamp()
	chat.UpdatedAt = chat.CreatedAt
	_, err := r.collection.InsertOne(ctx, chat)
	return err
}

// FindByID retrieves a chat by ID
func (r *MongoChatRepository) FindByID(ctx context.Context, id string) (*models.Chat, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var chat models.Chat
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&chat)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &chat, err
}

// FindByUserID retrieves chats for a user
func (r *MongoChatRepository) FindByUserID(ctx context.Context, userID string) ([]*models.Chat, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	cursor, err := r.collection.Find(ctx, bson.M{"users": oid})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var chats []*models.Chat
	if err := cursor.All(ctx, &chats); err != nil {
		return nil, err
	}
	return chats, nil
}

// UpdateLastMessage updates the chat's last message
func (r *MongoChatRepository) UpdateLastMessage(ctx context.Context, chatID, messageID string) error {
	oidChat, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return err
	}
	oidMessage, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": oidChat},
		bson.M{"$set": bson.M{"lastMessage": oidMessage, "updatedAt": primitive.NewObjectID().Timestamp()}},
	)
	return err
}
