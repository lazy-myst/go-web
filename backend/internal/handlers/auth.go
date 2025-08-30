package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/lazy-myst/go-web.git/internal/config"
	"github.com/lazy-myst/go-web.git/internal/models"
	"github.com/lazy-myst/go-web.git/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var cfg = config.LoadConfig()

func SetupAuthRoutes(api fiber.Router, mongoClient *mongo.Client) {
	auth := api.Group("/auth")
	auth.Post("/register", register(mongoClient))
	auth.Post("/login", login(mongoClient))

	users := api.Group("/users", authMiddleware)
	users.Get("/", listUsers(mongoClient))
}

func register(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		usersColl := mongoClient.Database("chatdb").Collection("users")

		var existing models.User
		if err := usersColl.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&existing); err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email already exists"})
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
		}

		user := models.User{
			ID:           auth.NewObjectIDHex(),
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: string(hash),
		}

		_, err = usersColl.InsertOne(context.TODO(), user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register user"})
		}

		return c.JSON(fiber.Map{"message": "User registered"})
	}
}

func login(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
		}

		usersColl := mongoClient.Database("chatdb").Collection("users")

		var user models.User
		if err := usersColl.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&user); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		token, err := auth.GenerateJWT(user.ID, cfg.JWTSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		return c.JSON(fiber.Map{"token": token})
	}
}

func listUsers(mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		usersColl := mongoClient.Database("chatdb").Collection("users")

		cursor, err := usersColl.Find(context.TODO(), bson.M{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
		}
		defer cursor.Close(context.TODO())

		var users []models.User
		if err := cursor.All(context.TODO(), &users); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode users"})
		}

		// Remove passwordHash from response
		for i := range users {
			users[i].PasswordHash = ""
		}

		return c.JSON(users)
	}
}

func authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	tokenString := authHeader[7:]
	userId, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	c.Locals("userId", userId)
	return c.Next()
}
