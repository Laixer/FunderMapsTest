package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/config"
	"fundermaps/internal/database"
)

const (
	AuthProviderJWT      = "jwt"
	AuthProviderAPIToken = "api_token"
)

const (
	HeaderXAPIKey = "X-API-Key"
)

func AuthMiddleware(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
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

		// TODO: Move into database stored procedure
		db.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'api_token', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)

		// TODO: Fetch organization and organization role
		// c.Locals("auth_provider", AuthProviderAPIToken)
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

	// var accessToken database.AuthAccessToken
	// result = db.First(&accessToken, "access_token = ?", token)
	// if result.Error != nil {
	// 	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	// 	}
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	// }

	// accessToken.UpdatedAt = time.Now()
	// if err := db.Save(&accessToken).Error; err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	// }

	// TODO: Fetch organization and organization role
	c.Locals("user", user)

	// return c.Next()
	// }

	// claims, err := auth.VerifyJWT(token)
	// if err != nil {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	// }

	// var user database.User
	// result := db.Model(&database.User{}).Where("id = ?", claims["id"]).Preload("Organizations").Find(&user)
	// if result.Error != nil {
	// 	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	// 	}
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	// }

	// // TODO: Move into database stored procedure
	// db.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)

	// // TODO: Fetch organization and organization role
	// c.Locals("auth_provider", AuthProviderJWT)
	// c.Locals("user", user)

	return c.Next()
}
