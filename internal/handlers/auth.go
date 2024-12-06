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

const JWTTokenValidity = time.Hour * 72

// func Hash(c *fiber.Ctx) error {
// 	type HashInput struct {
// 		Password string `json:"password"`
// 	}

// 	var input HashInput
// 	if err := c.BodyParser(&input); err != nil {
// 		return c.SendStatus(fiber.StatusBadRequest)
// 	}

// 	hash := utils.HashPassword(input.Password)

// 	return c.JSON(fiber.Map{"hash": hash})
// }

func SigninWithPassword(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	type LoginInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email and password required")
	}

	var user database.User
	result := db.First(&user, "email = ?", strings.ToLower(input.Email))

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	// TODO: From this point on, move into a platform service

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	// TODO: Check if account is locked

	if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
		// if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		//  // TODO: Increment access failed count
		// 	return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
		// }
		return c.Status(fiber.StatusNotImplemented).SendString("Not implemented")
	} else {
		if !utils.VerifyLegacyPassword(input.Password, user.PasswordHash) {
			// TODO: Increment access failed count
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
		}

		// TODO: Trigger password change if legacy password
	}

	ip := c.IP()
	if len(c.IPs()) > 1 {
		ip = c.IPs()[0]
	}

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE application.user SET access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)
		tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, ip, cfg.ApplicationID)
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

func RefreshToken(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	// TODO: From this point on, move into a platform service

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	// TODO: Check if account is locked

	ip := c.IP()
	if len(c.IPs()) > 1 {
		ip = c.IPs()[0]
	}

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, ip, cfg.ApplicationID)
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
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	var input ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.CurrentPassword == "" || input.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Current and new password required")
	}

	// TODO: From this point on, move into a platform service

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	// TODO: Check if account is locked

	if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
		// if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		//  // TODO: Increment access failed count
		// 	return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
		// }
		return c.Status(fiber.StatusNotImplemented).SendString("Not implemented")
	} else {
		if !utils.VerifyLegacyPassword(input.CurrentPassword, user.PasswordHash) {
			// TODO: Increment access failed count
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
		}
	}

	hash := utils.HashPassword(input.NewPassword)

	ip := c.IP()
	if len(c.IPs()) > 1 {
		ip = c.IPs()[0]
	}

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE application.user SET password_hash = ?, access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", hash, user.ID)
		tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, ip, cfg.ApplicationID)
		tx.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)

		return nil
	})

	return c.SendStatus(fiber.StatusNoContent)
}
