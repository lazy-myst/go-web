package repository

import (
	"context"

	"github.com/lazy-myst/go-web.git/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUserRepository implements UserRepository for MongoDB
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoUserRepository
func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	collection := db.Collection("users")
	// Create indexes
	for _, idx := range models.UserIndexes() {
		collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.D{{Key: idx, Value: 1}},
			Options: options.Index().SetUnique(idx == "email_1"),
		})
	}
	return &MongoUserRepository{collection: collection}
}

// Create inserts a new user into the database
func (r *MongoUserRepository) Create(ctx context.Context, user *models.User) error {
	if err := user.HashPassword(); err != nil {
		return err
	}
	user.ID = primitive.NewObjectID()
	user.CreatedAt = user.ID.Timestamp()
	user.UpdatedAt = user.CreatedAt
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindByID retrieves a user by ID
func (r *MongoUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// FindByEmail retrieves a user by email
func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// UpdateSocketID updates the user's socket ID
func (r *MongoUserRepository) UpdateSocketID(ctx context.Context, id, socketID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"socketId": socketID, "updatedAt": primitive.NewObjectID().Timestamp()}},
	)
	return err
}
