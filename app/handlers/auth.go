package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
	"fundermaps/app/mail"
	puser "fundermaps/app/platform/user"
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

// TODO: Move this to a service
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

// TODO: Move this to a service
func (ctx *AuthContext) generateTokensFromUser(clientID string, user database.User) (AuthToken, error) {
	return ctx.generateTokens(clientID, user.ID)
}

// TODO: Move this to a service
func (ctx *AuthContext) generateTokensFromAuthCode(authCode database.AuthCode) (AuthToken, error) {
	return ctx.generateTokens(authCode.ApplicationID, authCode.UserID)
}

// TODO: Move this to a service
func (ctx *AuthContext) generateTokensFromRefreshToken(refreshToken database.AuthRefreshToken) (AuthToken, error) {
	return ctx.generateTokens(refreshToken.ApplicationID, refreshToken.UserID)
}

// TODO: Move this to a service
func (ctx *AuthContext) revokeAuthCode(authCode database.AuthCode) error {
	return ctx.db.Delete(&database.AuthCode{}, "code = ?", authCode.Code).Error
}

// TODO: Move this to a service
func (ctx *AuthContext) revokeRefreshToken(refreshToken database.AuthRefreshToken) error {
	return ctx.db.Delete(&database.AuthRefreshToken{}, "token = ?", refreshToken.Token).Error
}

func Logout(c *fiber.Ctx) error {
	store := c.Locals("store").(*session.Store)

	sess, err := store.Get(c)
	if err != nil {
		return err
	}

	sess.Destroy()

	return c.Redirect("/")
}

func LoginWithForm(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	store := c.Locals("store").(*session.Store)

	sess, err := store.Get(c)
	if err != nil {
		return err
	}

	// if sess.Get("authenticated") != nil {
	// 	if redirectURI != "" && responseType == "code" {
	// 		userID, err := uuid.Parse(sess.Get("user_id").(string))
	// 		if err != nil {
	// 			return c.Status(fiber.StatusInternalServerError).SendString("Failed to parse user ID")
	// 		}

	// 		authCode, err := generateAuthCode(db, clientID, userID)
	// 		if err != nil {
	// 			return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate authorization code")
	// 		}
	// 		return c.Redirect(fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode, state))
	// 	}
	// 	return c.Redirect("/")
	// }

	userService := puser.NewService(db)

	email := c.FormValue("email")
	password := c.FormValue("password")

	// TODO: From this point on, move into a platform service

	var user database.User
	result := db.First(&user, "email = ?", strings.ToLower(email))
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if userService.IsLocked(&user) {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if user.PasswordHash == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	} else if strings.HasPrefix(user.PasswordHash, "$argon2id$") {
		if !utils.VerifyPassword(password, user.PasswordHash) {
			userService.IncrementAccessFailedCount(&user)
			return c.SendStatus(fiber.StatusUnauthorized)
		}
	} else {
		if !utils.VerifyLegacyPassword(password, user.PasswordHash) {
			userService.IncrementAccessFailedCount(&user)
			return c.SendStatus(fiber.StatusUnauthorized)
		}
	}

	db.Exec("UPDATE application.user SET access_failed_count = 0, login_count = login_count + 1, last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)

	// End platform service

	sess.Set("authenticated", true)
	sess.Set("user_id", user.ID.String())
	if err := sess.Save(); err != nil {
		return err
	}

	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	responseType := c.FormValue("response_type")
	state := c.FormValue("state")
	codeChallenge := c.FormValue("code_challenge")
	codeChallengeMethod := c.FormValue("code_challenge_method")

	// Check if we are dealing with an OAuth2 request
	if clientID != "" && responseType == "code" {
		_, err := getClient(db, clientID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid client ID")
		}

		var authCode string
		if codeChallengeMethod != "" && codeChallenge != "" {
			authCode, err = generateAuthCodeWithPKCE(db, clientID, user.ID, codeChallenge, codeChallengeMethod)
		} else {
			authCode, err = generateAuthCode(db, clientID, user.ID)
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to generate authorization code")
		}

		if redirectURI != "" && state != "" {
			return c.Redirect(fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode, state))
		} else if redirectURI != "" {
			return c.Redirect(fmt.Sprintf("%s?code=%s", redirectURI, authCode))
		}
	}

	if redirectURI != "" {
		return c.Redirect(redirectURI)
	}
	return c.Redirect("/")
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

	// End platform service

	ctx := AuthContext{db: db, ipAddress: c.IP()}
	authToken, err := ctx.generateTokensFromUser(cfg.ApplicationID, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}
	if err := revokeAPIKey(db, user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
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

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]any{
			"password_hash":       hash,
			"access_failed_count": 0,
			"login_count":         gorm.Expr("login_count + ?", 1),
			"last_login":          time.Now(),
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", user.ID).Delete(&database.ResetKey{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func ForgotPassword(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
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

	message := mail.Email{
		Subject:  "FunderMaps - Wachtwoord reset",
		From:     fmt.Sprintf("Fundermaps <no-reply@%s>", cfg.MailgunDomain),
		To:       []string{user.Email},
		Template: "reset-password",
		TemplateVars: map[string]any{
			"creatorName": user.Email, // TODO: Get the user's name
			"resetToken":  resetKey.Key,
		},
	}

	mailer := mail.NewMailer(cfg.MailgunDomain, cfg.MailgunAPIKey, cfg.MailgunAPIBase)
	if err := mailer.SendTemplatedMail(&message); err != nil {
		log.Printf("Failed to send email notification: %v\n", err)
	}

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
	result := db.First(&resetKey, "key = ?", input.ResetKey) // TODO: Add expiration date check
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

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]any{
			"password_hash":       hash,
			"access_failed_count": 0,
			"login_count":         gorm.Expr("login_count + ?", 1),
			"last_login":          time.Now(),
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", user.ID).Delete(&database.ResetKey{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server_error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
