package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Chat          primitive.ObjectID `bson:"chat"`
	Sender        primitive.ObjectID `bson:"sender"`
	Content       string             `bson:"content,omitempty"`
	Attachments   []Attachment       `bson:"attachments,omitempty"`
	ReplyTo       primitive.ObjectID `bson:"replyTo,omitempty"`
	Edited        bool               `bson:"edited"`
	EditHistory   []EditHistory      `bson:"editHistory,omitempty"`
	StatusPerUser []StatusPerUser    `bson:"statusPerUser,omitempty"`
	Reactions     []Reaction         `bson:"reactions,omitempty"`
	Pinned        bool               `bson:"pinned"`
	Deleted       bool               `bson:"deleted"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

type Attachment struct {
	URL      string `bson:"url"`
	FileName string `bson:"fileName"`
	FileType string `bson:"fileType"`
	FileSize int64  `bson:"fileSize"`
}

type EditHistory struct {
	Content  string    `bson:"content"`
	EditedAt time.Time `bson:"editedAt"`
}

type StatusPerUser struct {
	User      primitive.ObjectID `bson:"user"`
	Status    string             `bson:"status"` // pending, sent, delivered, seen
	UpdatedAt time.Time          `bson:"updatedAt"`
}

type Reaction struct {
	User     primitive.ObjectID `bson:"user"`
	Reaction string             `bson:"reaction"`
}

func MessageIndexes() []string {
	return []string{
		"chat_1_createdAt_-1", // Compound index for messages by chat and time
	}
}
