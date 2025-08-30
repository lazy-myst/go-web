package models

import "time"

type Chat struct {
	ID        string    `bson:"_id" json:"id"`
	UserIDs   []string  `bson:"userIds" json:"userIds"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
