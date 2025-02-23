package handlers

import (
	"errors"
	"fundermaps/internal/database"
	"fundermaps/internal/platform/user"
	"fundermaps/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func generateAuthCode(db *gorm.DB, clientID string, userID uuid.UUID) (string, error) {
	authToken := database.AuthCode{
		Code:          utils.GenerateRandomString(32),
		ApplicationID: clientID,
		UserID:        userID,
		ExpiredAt:     time.Now().Add(time.Minute * 5),
	}
	if err := db.Create(&authToken).Error; err != nil {
		return "", err
	}
	return authToken.Code, nil
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

func revokeAPIKey(db *gorm.DB, user database.User) error {
	db.Exec("DELETE FROM application.reset_key WHERE user_id = ?", user.ID)
	return nil
}

type UserInfo struct {
	Sub         string  `json:"sub"`
	Name        string  `json:"name"`
	GivenName   *string `json:"given_name,omitempty"`
	FamilyName  *string `json:"family_name,omitempty"`
	Email       string  `json:"email"`
	Picture     *string `json:"picture,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

func GetUserInfo(c *fiber.Ctx) error {
	user := c.Locals("user").(database.User)

	var name string
	if user.GivenName != nil {
		name = *user.GivenName
	}
	if user.LastName != nil {
		if name != "" {
			name += " "
		}
		name += *user.LastName
	}

	userInfo := UserInfo{
		Sub:         user.ID.String(),
		Name:        name,
		GivenName:   user.GivenName,
		FamilyName:  user.LastName,
		Email:       user.Email,
		Picture:     user.Avatar,
		PhoneNumber: user.PhoneNumber,
	}
	return c.JSON(userInfo)
}

func AuthorizationRequest(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	// scope := c.Query("scope")
	state := c.Query("state")

	_, err := getClient(db, clientID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid client ID")
	}
	// if !isValidRedirectURI(client, redirectURI) {
	// 	return c.Status(fiber.StatusBadRequest).SendString("Invalid redirect URI")
	// }

	if responseType != "code" {
		return c.Status(fiber.StatusUnsupportedMediaType).SendString("Unsupported response type")
	}

	userID, err := uuid.Parse("7a015c0a-55ce-4b8e-84b5-784bd3363d5b")
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("User authentication failed")
	}

	authCode, err := generateAuthCode(db, clientID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate authorization code")
	}

	return c.Redirect(redirectURI + "?code=" + authCode + "&state=" + state)
}

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
