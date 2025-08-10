package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty"`
	IsGroup     bool                 `bson:"isGroup"`
	Name        string               `bson:"name,omitempty"`
	Description string               `bson:"description,omitempty"`
	Avatar      string               `bson:"avatar,omitempty"`
	Users       []primitive.ObjectID `bson:"users"`
	Admins      []primitive.ObjectID `bson:"admins,omitempty"`
	LastMessage primitive.ObjectID   `bson:"lastMessage,omitempty"`
	CreatedAt   time.Time            `bson:"createdAt"`
	UpdatedAt   time.Time            `bson:"updatedAt"`
}

func ChatIndexes() []string {
	return []string{
		"users_1", // Index on users for efficient lookup
	}
}
