package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"fundermaps/internal/auth"
	"fundermaps/internal/config"
	"fundermaps/internal/database"
	"fundermaps/pkg/utils"
)

// TODO: Move into config
const JWTTokenValidity = time.Hour * 72

func SigninWithPassword(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	type LoginInput struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var user database.User
	result := db.First(&user, "email = ?", strings.ToLower(input.Email))

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// TODO: From this point on, move into a platform service

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
		if !utils.VerifyPassword(input.Password, user.PasswordHash) {
			db.Exec("UPDATE application.user SET access_failed_count = access_failed_count + 1 WHERE id = ?", user.ID)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	} else {
		if !utils.VerifyLegacyPassword(input.Password, user.PasswordHash) {
			db.Exec("UPDATE application.user SET access_failed_count = access_failed_count + 1 WHERE id = ?", user.ID)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	}

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE application.user SET access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)
		tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)
		tx.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)

		return nil
	})

	claims := jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(JWTTokenValidity).Unix(),
	}

	token, err := auth.GenerateJWT(claims)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": token})
}

// TODO: Succeeded by Oauth2 Refresh Token
func RefreshToken(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	// TODO: From this point on, move into a platform service

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)
		tx.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)

		return nil
	})

	claims := jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(JWTTokenValidity).Unix(),
	}

	token, err := auth.GenerateJWT(claims)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": token})
}

func ChangePassword(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	type ChangePasswordInput struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,min=6"`
	}

	var input ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// TODO: From this point on, move into a platform service

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
		if !utils.VerifyPassword(input.CurrentPassword, user.PasswordHash) {
			db.Exec("UPDATE application.user SET access_failed_count = access_failed_count + 1 WHERE id = ?", user.ID)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	} else {
		if !utils.VerifyLegacyPassword(input.CurrentPassword, user.PasswordHash) {
			db.Exec("UPDATE application.user SET access_failed_count = access_failed_count + 1 WHERE id = ?", user.ID)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	}

	hash := utils.HashPassword(input.NewPassword)

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE application.user SET password_hash = ?, access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", hash, user.ID)
		tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)
		tx.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)

		return nil
	})

	return c.SendStatus(fiber.StatusNoContent)
}
