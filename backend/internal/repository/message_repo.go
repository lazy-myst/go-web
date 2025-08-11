package repository

import (
	"context"
	"strings"

	"github.com/lazy-myst/go-web.git/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoMessageRepository implements MessageRepository for MongoDB
type MongoMessageRepository struct {
	collection *mongo.Collection
}

// NewMongoMessageRepository creates a new MongoMessageRepository
func NewMongoMessageRepository(db *mongo.Database) *MongoMessageRepository {
	collection := db.Collection("messages")
	// Create indexes
	for _, idx := range models.MessageIndexes() {
		keys := bson.D{}
		for _, part := range strings.Split(idx, "_") {
			key, order := part, 1
			if strings.HasSuffix(part, "-1") {
				key, order = strings.TrimSuffix(part, "-1"), -1
			}
			keys = append(keys, bson.E{Key: key, Value: order})
		}
		collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{Keys: keys})
	}
	return &MongoMessageRepository{collection: collection}
}

// Create inserts a new message into the database
func (r *MongoMessageRepository) Create(ctx context.Context, message *models.Message) error {
	message.ID = primitive.NewObjectID()
	message.CreatedAt = message.ID.Timestamp()
	message.UpdatedAt = message.CreatedAt
	_, err := r.collection.InsertOne(ctx, message)
	return err
}

// FindByChatID retrieves messages for a chat
func (r *MongoMessageRepository) FindByChatID(ctx context.Context, chatID string, limit int) ([]*models.Message, error) {
	oid, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, err
	}
	opts := options.Find().SetSort(bson.D{{"createdAt", -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"chat": oid}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var messages []*models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// UpdateStatus updates the message status for a user
func (r *MongoMessageRepository) UpdateStatus(ctx context.Context, messageID, userID, status string) error {
	oidMessage, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}
	oidUser, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": oidMessage},
		bson.M{
			"$set": bson.M{
				"statusPerUser.$[elem].status":    status,
				"statusPerUser.$[elem].updatedAt": primitive.NewObjectID().Timestamp(),
			},
		},
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{bson.M{"elem.user": oidUser}},
		}),
	)
	return err
}
