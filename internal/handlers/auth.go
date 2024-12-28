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
	"fundermaps/internal/platform/user"
	"fundermaps/pkg/utils"
)

const accessTokenExp = 3600
const refreshTokenExp = 365

type AuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
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

	if user.AccessFailedCount >= 5 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Account locked"})
	}

	if user.PasswordHash == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid credentials"})
	} else if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
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
	if err := revokeAuthKey(db, user); err != nil {
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

	// if err := revokeAuthKey(db, user); err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	// }

	return c.JSON(authToken)
}

func ChangePassword(c *fiber.Ctx) error {
	// cfg := c.Locals("config").(*config.Config)
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
		// tx.Exec("INSERT INTO application.auth_session (user_id, ip_address, application_id, provider, updated_at) VALUES (?, ?, ?, 'jwt', now()) ON CONFLICT ON constraint auth_session_pkey DO UPDATE SET updated_at = excluded.updated_at, ip_address = excluded.ip_address;", user.ID, c.IP(), cfg.ApplicationID)
		tx.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)

		return nil
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func AuthorizationRequest(c *fiber.Ctx) error {
	// db := c.Locals("db").(*gorm.DB)

	// clientID := c.Query("client_id")
	// redirectURI := c.Query("redirect_uri")
	// responseType := c.Query("response_type")
	// scope := c.Query("scope")
	// state := c.Query("state") // Optional, for CSRF protection

	// client, err := getClient(db, clientID)
	// if err != nil {
	// 	return c.Status(http.StatusBadRequest).SendString("Invalid client ID")
	// }
	// if !isValidRedirectURI(client, redirectURI) {
	// 	return c.Status(http.StatusBadRequest).SendString("Invalid redirect URI")
	// }

	// Validate response_type (should be "code" for authorization code flow)
	// if responseType != "code" {
	// 	return c.Status(http.StatusUnsupportedMediaType).SendString("Unsupported response type")
	// }

	// 2. Authenticate the user (if not already authenticated)
	// This might involve redirecting to a login page or checking an existing session
	// userID, err := authenticateUser(c)
	// if err != nil {
	// 	return c.Status(http.StatusUnauthorized).SendString("User authentication failed")
	// }

	// // 3. (Optional) Display a consent screen to the user
	// if shouldShowConsentScreen(client, userID, scope) {
	// 	// Render a consent screen with details about the requested permissions
	// 	// If the user denies access, redirect back with an "access_denied" error
	// 	if !userConsents(c) { // Implement your consent logic
	// 		return redirectWithError(redirectURI, "access_denied", state)
	// 	}
	// }

	// 4. Generate an authorization code
	// authCode, err := generateAuthCode(config, clientID, userID, redirectURI, scope)
	// if err != nil {
	// 	return c.Status(http.StatusInternalServerError).SendString("Failed to generate authorization code")
	// }

	// return c.Redirect(redirectURI + "?code=" + authCode + "&state=" + state)
	return c.SendStatus(fiber.StatusNotImplemented)
}

func getClient(db *gorm.DB, clientID string) (database.Application, error) {
	if clientID == "" {
		return database.Application{}, errors.New("client_id is required")
	}

	var client database.Application
	result := db.First(&client, "application_id = ?", clientID)
	if result.Error != nil {
		return client, result.Error
	}
	return client, nil
}

func getAuthCode(db *gorm.DB, clientID string, code string) (database.AuthCode, error) {
	if code == "" {
		return database.AuthCode{}, errors.New("code is required")
	}

	var authToken database.AuthCode
	result := db.First(&authToken, "code = ? AND application_id = ? AND expired_at > now()", code, clientID)
	if result.Error != nil {
		return authToken, result.Error
	}
	return authToken, nil
}

func getRefreshToken(db *gorm.DB, clientID string, refreshToken string) (database.AuthRefreshToken, error) {
	if refreshToken == "" {
		return database.AuthRefreshToken{}, errors.New("refresh_token is required")
	}

	var token database.AuthRefreshToken
	result := db.First(&token, "token = ? AND application_id = ? AND expired_at > now()", refreshToken, clientID)
	if result.Error != nil {
		return token, result.Error
	}
	return token, nil
}

func revokeAuthKey(db *gorm.DB, user database.User) error {
	db.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)
	return nil
}

// TODO: Move to other file
func TokenRequest(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	type tokenRequest struct{}

	if err := c.BodyParser(&tokenRequest{}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid_request"})
	}

	clientID := c.FormValue("client_id")
	clientSecret := c.FormValue("client_secret")
	client, err := getClient(db, clientID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_client"})
	}

	if !utils.VerifyPassword(clientSecret, client.Secret) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_client"})
	}

	grantType := c.FormValue("grant_type")
	switch grantType {
	case "authorization_code":
		code := c.FormValue("code")
		authCode, err := getAuthCode(db, clientID, code)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
		}

		user, err := userService.GetUserByID(authCode.UserID)
		if err != nil {
			if errors.Is(err, errors.New("user not found")) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		if userService.IsLocked(user) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
		}

		ctx := AuthContext{db: db, ipAddress: c.IP()}
		authToken, err := ctx.generateTokensFromAuthCode(authCode)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}
		if err := ctx.revokeAuthCode(authCode); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		return c.JSON(authToken)

	case "client_credentials":
		userID := uuid.MustParse("7a015c0a-55ce-4b8e-84b5-784bd3363d5b")

		user, err := userService.GetUserByID(userID)
		if err != nil {
			if errors.Is(err, errors.New("user not found")) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		if userService.IsLocked(user) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
		}

		ctx := AuthContext{db: db, ipAddress: c.IP()}
		authToken, err := ctx.generateTokensFromUser(clientID, *user) // TODO: pass pointer
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		return c.JSON(authToken)

	case "refresh_token":
		refreshToken := c.FormValue("refresh_token")
		refresh, err := getRefreshToken(db, clientID, refreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
		}

		user, err := userService.GetUserByID(refresh.UserID)
		if err != nil {
			if errors.Is(err, errors.New("user not found")) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		if userService.IsLocked(user) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
		}

		ctx := AuthContext{db: db, ipAddress: c.IP()}
		authToken, err := ctx.generateTokensFromRefreshToken(refresh)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}
		if err := ctx.revokeRefreshToken(refresh); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		return c.JSON(authToken)

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported_grant_type"})
	}
}
