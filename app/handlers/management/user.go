package mngmt

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
	"fundermaps/app/platform/user"
	"fundermaps/pkg/utils"
)

func CreateUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	type UserInput struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	var input UserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// TODO: Create a normalizer for email
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, _ := userService.GetUserByEmail(email)
	if user != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User already exists"})
	}

	user = &database.User{
		Email:        email,
		PasswordHash: utils.HashLegacyPassword(input.Password),
		Role:         "user",
	}

	err = userService.Create(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(user)
}

func GetAllUsers(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var users []database.User
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)
	result := db.Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(users)
}

func GetUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	uid, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID"})
	}
	user, err := userService.GetUserByID(uid)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	type UpdateUserInput struct {
		Email       *string `json:"email" validate:"omitempty,email"`
		GivenName   *string `json:"given_name"`
		LastName    *string `json:"family_name"`
		Avatar      *string `json:"picture"`
		JobTitle    *string `json:"job_title"`
		PhoneNumber *string `json:"phone_number"`
		Role        *string `json:"role" validate:"omitempty,oneof=user administrator service"`
	}

	var input UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	uid, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID"})
	}
	user, err := userService.GetUserByID(uid)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if input.Email != nil && *input.Email != "" {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		existingUser, _ := userService.GetUserByEmail(email)
		if existingUser != nil && existingUser.ID != user.ID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Email already exists"})
		}
		user.Email = email
	}
	if input.GivenName != nil && *input.GivenName != "" {
		user.GivenName = input.GivenName
	}
	if input.LastName != nil && *input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Avatar != nil && *input.Avatar != "" {
		user.Avatar = input.Avatar
	}
	if input.JobTitle != nil && *input.JobTitle != "" {
		user.JobTitle = input.JobTitle
	}
	if input.PhoneNumber != nil && *input.PhoneNumber != "" {
		user.PhoneNumber = input.PhoneNumber
	}
	if input.Role != nil && *input.Role != "" {
		user.Role = *input.Role
	}

	err = userService.Update(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(user)
}

func ResetUserPassword(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	type ResetPasswordInput struct {
		Password string `json:"password" validate:"required,min=6"`
	}

	var input ResetPasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	uid, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID"})
	}
	user, err := userService.GetUserByID(uid)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	err = userService.UpdatePassword(user, input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func GetApiKeys(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	uid, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID"})
	}

	var keys []database.AuthKey
	result := db.Where("user_id = ?", uid).Find(&keys)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(keys)
}

func CreateApiKey(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var user database.User
	result := db.First(&user, "id = ?", c.Params("user_id"))
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User not found"})
	}

	user.Role = "service"
	result = db.Save(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	apiKey := database.AuthKey{
		UserID: user.ID,
	}

	result = db.Create(&apiKey)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(apiKey)
}

func DeleteApiKey(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type DeleteApiKeyInput struct {
		Key string `json:"key" validate:"required"`
	}

	var input DeleteApiKeyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	result := db.Where("key = ? AND user_id = ?", input.Key, c.Params("user_id")).Delete(&database.AuthKey{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "API key not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func DeleteUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	uid, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID"})
	}

	user, err := userService.GetUserByID(uid)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// Remove user from all organizations first
	result := db.Exec("DELETE FROM application.organization_user WHERE user_id = ?", user.ID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// Remove API keys
	result = db.Where("user_id = ?", user.ID).Delete(&database.AuthKey{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// Delete the user
	result = db.Delete(user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
