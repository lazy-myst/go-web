package models

import "time"

type Message struct {
	ID        string    `bson:"_id" json:"id"`
	ChatID    string    `bson:"chatId" json:"chatId"`
	SenderID  string    `bson:"senderId" json:"senderId"`
	Text      string    `bson:"text" json:"text"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
