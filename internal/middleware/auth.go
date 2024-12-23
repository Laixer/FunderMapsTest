package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/database"
)

const (
	HeaderXAPIKey = "X-API-Key"
)

func AuthMiddleware(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	xAPIKey := c.Get(HeaderXAPIKey)
	if xAPIKey != "" {
		var user database.User
		result := db.Joins("JOIN application.auth_key ON application.auth_key.user_id = application.user.id").
			Where("application.auth_key.key = ?", xAPIKey).
			Preload("Organizations").
			First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}

		// TODO: Fetch organization and organization role
		c.Locals("user", user)

		return c.Next()
	}

	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// if strings.HasPrefix(token, "fmat") {
	var user database.User
	result := db.Joins("JOIN application.auth_access_token ON application.auth_access_token.user_id = application.user.id").
		Where("application.auth_access_token.access_token = ? AND application.auth_access_token.expired_at > now()", token).
		Preload("Organizations").
		First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// TODO: Fetch organization and organization role
	c.Locals("user", user)

	return c.Next()
}
