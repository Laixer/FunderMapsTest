package handlers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/config"
	"fundermaps/internal/database"
	"fundermaps/pkg/utils"
)

type AuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// TODO: Pass clientID as a parameter
func generateTokensFromUser(c *fiber.Ctx, user database.User) (AuthToken, error) {
	cfg := c.Locals("config").(*config.Config) // TODO: Don't use cfg
	db := c.Locals("db").(*gorm.DB)

	authAccessToken := database.AuthAccessToken{
		AccessToken:   fmt.Sprintf("fmat%s", utils.GenerateRandomString(40)),
		IPAddress:     c.IP(),
		ApplicationID: cfg.ApplicationID,
		UserID:        user.ID,
		ExpiredAt:     time.Now().Add(1 * time.Hour), // TODO: Move into config
	}
	db.Create(&authAccessToken)

	authRefreshToken := database.AuthRefreshToken{
		Token:         fmt.Sprintf("fmrt%s", utils.GenerateRandomString(40)),
		ApplicationID: cfg.ApplicationID,
		UserID:        user.ID,
		ExpiredAt:     time.Now().AddDate(1, 0, 0), // TODO: Move into config
	}
	db.Create(&authRefreshToken)

	// TODO; Dont need this here
	// TODO: Move into database stored procedure
	db.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE application.user SET access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)

		return nil
	})

	authToken := AuthToken{
		AccessToken:  authAccessToken.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, //config.AccessTokenExp,
		RefreshToken: authRefreshToken.Token,
	}
	return authToken, nil
}

func generateTokensFromAuthCode(c *fiber.Ctx, authCode database.AuthCode) (AuthToken, error) {
	db := c.Locals("db").(*gorm.DB)

	authAccessToken := database.AuthAccessToken{
		AccessToken:   fmt.Sprintf("fmat%s", utils.GenerateRandomString(40)),
		IPAddress:     c.IP(),
		ApplicationID: authCode.ApplicationID,
		UserID:        authCode.UserID,
		ExpiredAt:     time.Now().Add(1 * time.Hour),
	}
	db.Create(&authAccessToken)

	authRefreshToken := database.AuthRefreshToken{
		Token:         fmt.Sprintf("fmrt%s", utils.GenerateRandomString(40)),
		ApplicationID: authCode.ApplicationID,
		UserID:        authCode.UserID,
		ExpiredAt:     time.Now().AddDate(1, 0, 0),
	}
	db.Create(&authRefreshToken)

	authToken := AuthToken{
		AccessToken:  authAccessToken.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, //config.AccessTokenExp,
		RefreshToken: authRefreshToken.Token,
	}
	return authToken, nil
}

func generateTokensFromRefreshToken(c *fiber.Ctx, refreshToken database.AuthRefreshToken) (AuthToken, error) {
	db := c.Locals("db").(*gorm.DB)

	authAccessToken := database.AuthAccessToken{
		AccessToken:   fmt.Sprintf("fmat%s", utils.GenerateRandomString(40)),
		IPAddress:     c.IP(),
		ApplicationID: refreshToken.ApplicationID,
		UserID:        refreshToken.UserID,
		ExpiredAt:     time.Now().Add(1 * time.Hour),
	}
	db.Create(&authAccessToken)

	authRefreshToken := database.AuthRefreshToken{
		Token:         fmt.Sprintf("fmrt%s", utils.GenerateRandomString(40)),
		ApplicationID: refreshToken.ApplicationID,
		UserID:        refreshToken.UserID,
		ExpiredAt:     time.Now().AddDate(1, 0, 0),
	}
	db.Create(&authRefreshToken)

	authToken := AuthToken{
		AccessToken:  authAccessToken.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, //config.AccessTokenExp,
		RefreshToken: authRefreshToken.Token,
	}
	return authToken, nil
}

func SigninWithPassword(c *fiber.Ctx) error {
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

	authToken, err := generateTokensFromUser(c, user)
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

	authToken, err := generateTokensFromRefreshToken(c, refreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}
	if authToken.RefreshToken != "" {
		if err := revokeRefreshToken(db, refreshToken.Token); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}
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

func revokeAuthCode(db *gorm.DB, code string) error {
	return db.Delete(&database.AuthCode{}, "code = ?", code).Error
}

func revokeRefreshToken(db *gorm.DB, refreshToken string) error {
	return db.Delete(&database.AuthRefreshToken{}, "token = ?", refreshToken).Error
}

func revokeAuthKey(db *gorm.DB, user database.User) error {
	db.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)
	return nil
}

func TokenRequest(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

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

		var user database.User
		result := db.First(&user, "id = ?", authCode.UserID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		if user.AccessFailedCount >= 5 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
		}

		authToken, err := generateTokensFromAuthCode(c, authCode)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}
		if err := revokeAuthCode(db, authCode.Code); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		return c.JSON(authToken)

	case "refresh_token":
		refreshToken := c.FormValue("refresh_token")
		refresh, err := getRefreshToken(db, clientID, refreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
		}

		var user database.User
		result := db.First(&user, "id = ?", refresh.UserID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid_grant"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}

		if user.AccessFailedCount >= 5 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account_locked"})
		}

		authToken, err := generateTokensFromRefreshToken(c, refresh)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
		}
		if authToken.RefreshToken != "" {
			if err := revokeRefreshToken(db, refresh.Token); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
			}
		}

		return c.JSON(authToken)

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported_grant_type"})
	}
}
