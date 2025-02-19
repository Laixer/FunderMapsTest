package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fundermaps/internal/config"
	"fundermaps/internal/database"
	"fundermaps/internal/platform/user"
	"fundermaps/pkg/utils"
)

func GetAllOrganizations(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var orgs []database.Organization
	result := db.Find(&orgs)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(orgs)
}

// TODO: Check if organization name already exists
func CreateOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type OrganizationInput struct {
		Name string `json:"name" validate:"required"`
	}

	var input OrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	org := database.Organization{
		Name:  input.Name,
		Email: fmt.Sprintf("info@%s.com", utils.GenerateRandomString(10)),
	}

	result := db.Create(&org)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(org)
}

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
	result := db.Find(&users)
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

func GetOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	return c.JSON(org)
}

func GetAllOrganizationUsers(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	var users []database.User
	result = db.Joins("JOIN application.organization_user ON application.organization_user.user_id = application.user.id").
		Where("application.organization_user.organization_id = ?", org.ID).
		Find(&users)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	return c.JSON(users)
}

func AddUserToOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type AddUserToOrganizationInput struct {
		UserID string  `json:"user_id" validate:"required"`
		Role   *string `json:"role"` // TODO: Validate role
	}

	var input AddUserToOrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var user database.User
	result := db.First(&user, "id = ?", input.UserID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("User not found")
	}

	var org database.Organization
	result = db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	// TODO: Maybe use as default value in validation
	if input.Role == nil {
		role := "user"
		input.Role = &role
	}

	// TODO: Return the organization user combination
	result = db.Exec("INSERT INTO application.organization_user (user_id, organization_id, role) VALUES (?, ?, ?)", user.ID, org.ID, input.Role)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	return c.SendStatus(fiber.StatusCreated) // TODO: Only send status with no content
}

func RemoveUserFromOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type RemoveUserFromOrganizationInput struct {
		UserID string `json:"user_id" validate:"required"`
	}

	var input RemoveUserFromOrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var user database.User
	result := db.First(&user, "id = ?", input.UserID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("User not found")
	}

	var org database.Organization
	result = db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	result = db.Exec("DELETE FROM application.organization_user WHERE user_id = ? AND organization_id = ?", user.ID, org.ID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func AddMapsetToOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type AddMapsetToOrganizationInput struct {
		MapsetID string `json:"mapset_id" validate:"required"`
	}

	var input AddMapsetToOrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// TODO: Just do an insert into the database, the foreign key constraints will handle the rest
	// var mapset database.Mapset
	// result := db.First(&mapset, "id = ?", input.MapsetID)
	// if result.Error != nil {
	// 	return c.Status(fiber.StatusBadRequest).SendString("Mapset not found")
	// }

	// TODO: Just do an insert into the database, the foreign key constraints will handle the rest
	var org database.Organization
	result := db.First(&org, "id = ?", c.Params("org_id"))
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	result = db.Exec("INSERT INTO maplayer.map_organization (map_id, organization_id) VALUES (?, ?)", input.MapsetID, org.ID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	return c.SendStatus(fiber.StatusCreated)
}

func RemoveMapsetFromOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type RemoveMapsetFromOrganizationInput struct {
		MapsetID string `json:"mapset_id" validate:"required"`
	}

	var input RemoveMapsetFromOrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// TODO: Just do an insert into the database, the foreign key constraints will handle the rest
	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	result = db.Exec("DELETE FROM maplayer.map_organization WHERE map_id = ? AND organization_id = ?", input.MapsetID, org.ID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func CreateAuthKey(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	userID := c.Params("user_id")

	var user database.User
	result := db.First(&user, "id = ?", userID)
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
