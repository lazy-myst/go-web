package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lazy-myst/go-web.git/internal/models"
	"github.com/lazy-myst/go-web.git/internal/repository"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo repository.UserRepository
	secret   string // JWT secret key
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repository.UserRepository, secret string) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		secret:   secret,
	}
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	// Check if email exists
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Create user
	user := &models.User{
		Name:     name,
		Email:    email,
		Password: password,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login authenticates a user and returns a JWT
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if user == nil || !user.ComparePassword(password) {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", nil, err
	}
	return tokenString, user, nil
}

// UpdateSocketID updates the user's WebSocket ID
func (s *AuthService) UpdateSocketID(ctx context.Context, userID, socketID string) error {
	return s.userRepo.UpdateSocketID(ctx, userID, socketID)
}
