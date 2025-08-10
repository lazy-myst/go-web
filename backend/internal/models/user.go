package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Name           string             `bson:"name"`
	Email          string             `bson:"email"`
	Password       string             `bson:"password"`
	ProfilePicture string             `bson:"profilePicture,omitempty"`
	StatusMessage  string             `bson:"statusMessage,omitempty"`
	LastSeen       time.Time          `bson:"lastSeen"`
	SocketID       string             `bson:"socketId,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt"`
}

func (u *User) HashPassword() error {
	if u.Password == "" {
		return nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func UserIndexes() []string {
	return []string{
		"email_1", // Unique index on email
	}
}
