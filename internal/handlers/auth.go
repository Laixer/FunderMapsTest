package handlers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fundermaps/internal/config"
	"fundermaps/internal/database"
	puser "fundermaps/internal/platform/user"
	"fundermaps/pkg/utils"
)

const accessTokenExp = 3600
const refreshTokenExp = 365

type AuthToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token"`
}

type AuthContext struct {
	db        *gorm.DB
	ipAddress string
}

func (ctx *AuthContext) generateTokens(clientID string, userID uuid.UUID) (AuthToken, error) {
	const tokenType = "Bearer"

	authAccessToken := database.AuthAccessToken{
		AccessToken:   fmt.Sprintf("fmat%s", utils.GenerateRandomString(40)),
		IPAddress:     ctx.ipAddress,
		ApplicationID: clientID,
		UserID:        userID,
		ExpiredAt:     time.Now().Add(accessTokenExp * time.Second),
	}
	authRefreshToken := database.AuthRefreshToken{
		Token:         fmt.Sprintf("fmrt%s", utils.GenerateRandomString(40)),
		ApplicationID: clientID,
		UserID:        userID,
		ExpiredAt:     time.Now().AddDate(0, 0, refreshTokenExp),
	}

	// TODO: Transaction should be handled by the service
	ctx.db.Transaction(func(tx *gorm.DB) error {
		tx.Create(&authAccessToken)
		tx.Create(&authRefreshToken)

		return nil
	})

	authToken := AuthToken{
		AccessToken:  authAccessToken.AccessToken,
		TokenType:    tokenType,
		ExpiresIn:    accessTokenExp,
		ExpiresAt:    authAccessToken.ExpiredAt,
		RefreshToken: authRefreshToken.Token,
	}
	return authToken, nil
}

func (ctx *AuthContext) generateTokensFromUser(clientID string, user database.User) (AuthToken, error) {
	return ctx.generateTokens(clientID, user.ID)
}

func (ctx *AuthContext) generateTokensFromAuthCode(authCode database.AuthCode) (AuthToken, error) {
	return ctx.generateTokens(authCode.ApplicationID, authCode.UserID)
}

func (ctx *AuthContext) generateTokensFromRefreshToken(refreshToken database.AuthRefreshToken) (AuthToken, error) {
	return ctx.generateTokens(refreshToken.ApplicationID, refreshToken.UserID)
}

func (ctx *AuthContext) revokeAuthCode(authCode database.AuthCode) error {
	return ctx.db.Delete(&database.AuthCode{}, "code = ?", authCode.Code).Error
}

func (ctx *AuthContext) revokeRefreshToken(refreshToken database.AuthRefreshToken) error {
	return ctx.db.Delete(&database.AuthRefreshToken{}, "token = ?", refreshToken.Token).Error
}

func SigninWithPassword(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	userService := puser.NewService(db)

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

	// TODO: From this point on, move into a platform service

	var user database.User
	result := db.First(&user, "email = ?", strings.ToLower(input.Email))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if userService.IsLocked(&user) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
	}

	if user.PasswordHash == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
	} else if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
		if !utils.VerifyPassword(input.Password, user.PasswordHash) {
			userService.IncrementAccessFailedCount(&user)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	} else {
		if !utils.VerifyLegacyPassword(input.Password, user.PasswordHash) {
			userService.IncrementAccessFailedCount(&user)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	}

	db.Exec("UPDATE application.user SET access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)

	// TODO: Move into database stored procedure
	// db.Transaction(func(tx *gorm.DB) error {
	// 	tx.Exec("UPDATE application.user SET access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)

	// 	return nil
	// })

	ctx := AuthContext{db: db, ipAddress: c.IP()}
	authToken, err := ctx.generateTokensFromUser(cfg.ApplicationID, user)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if err := revokeAPIKey(db, user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}

	return c.JSON(authToken)
}

func RefreshToken(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	type RefreshInput struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	var input RefreshInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// TODO: From this point on, move into a platform service

	refreshToken, err := getRefreshToken(db, cfg.ApplicationID, input.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
	}

	var user database.User
	result := db.First(&user, "id = ?", refreshToken.UserID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	ctx := AuthContext{db: db, ipAddress: c.IP()}
	authToken, err := ctx.generateTokensFromRefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}
	if err := ctx.revokeRefreshToken(refreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// if err := revokeAPIKey(db, user); err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	// }

	return c.JSON(authToken)
}

func ChangePassword(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	userService := puser.NewService(db)

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

	if userService.IsLocked(&user) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
	}

	if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
		if !utils.VerifyPassword(input.CurrentPassword, user.PasswordHash) {
			userService.IncrementAccessFailedCount(&user)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	} else {
		if !utils.VerifyLegacyPassword(input.CurrentPassword, user.PasswordHash) {
			userService.IncrementAccessFailedCount(&user)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	}

	hash := utils.HashLegacyPassword(input.NewPassword)

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE application.user SET password_hash = ?, access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", hash, user.ID)
		// tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)
		tx.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)

		return nil
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func ForgotPassword(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := puser.NewService(db)

	type ForgotPasswordInput struct {
		Email string `json:"email" validate:"required,email"`
	}

	var input ForgotPasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	user, err := userService.GetUserByEmail(input.Email)
	if err != nil {
		if errors.Is(err, errors.New("user not found")) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_email"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}

	if userService.IsLocked(user) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
	}

	resetKey := database.ResetKey{
		Key:    uuid.New(),
		UserID: user.ID,
	}

	if err := db.Create(&resetKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}

	// Send reset email (implementation depends on your email service)
	// if err := userService.SendResetEmail(user, resetKey.Token); err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	// }

	return c.SendStatus(fiber.StatusNoContent)
}

func ResetPassword(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := puser.NewService(db)

	type ResetPasswordInput struct {
		ResetKey    string `json:"reset_key" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=6"`
	}

	var input ResetPasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var resetKey database.ResetKey
	result := db.First(&resetKey, "key = ?", input.ResetKey)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_reset_key"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}

	user, err := userService.GetUserByID(resetKey.UserID)
	if err != nil {
		if errors.Is(err, errors.New("user not found")) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_reset_key"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}

	if userService.IsLocked(user) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
	}

	hash := utils.HashLegacyPassword(input.NewPassword)

	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE application.user SET password_hash = ?, access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", hash, user.ID)
		// tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)
		tx.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)

		return nil
	})

	return c.SendStatus(fiber.StatusNoContent)
}
