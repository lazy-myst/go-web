package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateJWT creates a JWT token for a user
func GenerateJWT(userID, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(secret))
}

// ValidateJWT verifies a JWT token and returns the user ID
func ValidateJWT(tokenString, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", jwt.ErrTokenExpired
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	return userId, nil
}

// NewObjectIDHex generates a new MongoDB ObjectID as a hex string
func NewObjectIDHex() string {
	return primitive.NewObjectID().Hex()
}
