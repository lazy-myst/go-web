package models

type User struct {
	ID           string `bson:"_id" json:"id"`
	Name         string `bson:"name" json:"name"`
	Email        string `bson:"email" json:"email"`
	PasswordHash string `bson:"passwordHash" json:"-"`
}
