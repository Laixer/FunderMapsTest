package mngmt

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fundermaps/internal/config"
	"fundermaps/internal/database"
	"fundermaps/internal/platform/user"
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

func GetUserByEmail(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	user, err := userService.GetUserByEmail(c.Params("email"))
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(user)
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

// TODO: There are likely better ways to update a user
// TODO: Related to /me endpoint
func UpdateUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userService := user.NewService(db)

	type UpdateUserInput struct {
		GivenName   *string `json:"given_name"`
		LastName    *string `json:"family_name"`
		Avatar      *string `json:"picture"`
		JobTitle    *string `json:"job_title"`
		PhoneNumber *string `json:"phone_number"`
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

	if input.GivenName != nil {
		user.GivenName = input.GivenName
	}
	if input.LastName != nil {
		user.LastName = input.LastName
	}
	if input.Avatar != nil {
		user.Avatar = input.Avatar
	}
	if input.JobTitle != nil {
		user.JobTitle = input.JobTitle
	}
	if input.PhoneNumber != nil {
		user.PhoneNumber = input.PhoneNumber
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User not found"})
	}

	err = userService.UpdatePassword(user, input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func CreateAuthKey(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var user database.User
	result := db.First(&user, "id = ?", c.Params("user_id"))
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User not found"})
	}

	authKey := database.AuthKey{
		UserID: user.ID,
	}

	result = db.Create(&authKey)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(authKey)
}
