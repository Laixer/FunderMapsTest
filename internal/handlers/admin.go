package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/database"
	"fundermaps/pkg/utils"
)

func CreateApplication(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var application database.Application
	if err := c.BodyParser(&application); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if application.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Name required",
		})
	}

	result := db.Create(&application)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(application)
}

func CreateOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type OrganizationInput struct {
		Name string `json:"name"`
	}

	var input OrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name required")
	}

	org := database.Organization{
		Name:  input.Name,
		Email: fmt.Sprintf("info@%s.com", utils.GenerateRandomString(10)),
	}

	result := db.Create(&org)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(org)
}

func CreateUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type UserInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input UserInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email and password required")
	}

	user := database.User{
		Email:        input.Email,
		PasswordHash: utils.HashPassword(input.Password),
		Role:         "user",
	}

	result := db.Create(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(user)
}

func AddUserToOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type AddUserToOrganizationInput struct {
		UserID         string  `json:"user_id"`
		OrganizationID string  `json:"organization_id"`
		Role           *string `json:"role"`
	}

	var input AddUserToOrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.UserID == "" || input.OrganizationID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("User ID and Organization ID required")
	}

	var user database.User
	result := db.First(&user, "id = ?", input.UserID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("User not found")
	}

	var org database.Organization
	result = db.First(&org, "id = ?", input.OrganizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	if input.Role == nil {
		role := "user"
		input.Role = &role
	}

	// TODO: Return the organization user combination
	result = db.Exec("INSERT INTO application.organization_user (user_id, organization_id, role) VALUES (?, ?, ?)", user.ID, org.ID, input.Role)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	return c.SendStatus(fiber.StatusCreated)
}

func AddMapsetToOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type AddMapsetToOrganizationInput struct {
		MapsetID       string `json:"mapset_id"`
		OrganizationID string `json:"organization_id"`
	}

	var input AddMapsetToOrganizationInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.MapsetID == "" || input.OrganizationID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Mapset ID and Organization ID required")
	}

	// TODO: Just do an insert into the database, the foreign key constraints will handle the rest
	// var mapset database.Mapset
	// result := db.First(&mapset, "id = ?", input.MapsetID)
	// if result.Error != nil {
	// 	return c.Status(fiber.StatusBadRequest).SendString("Mapset not found")
	// }

	// TODO: Just do an insert into the database, the foreign key constraints will handle the rest
	var org database.Organization
	result := db.First(&org, "id = ?", input.OrganizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	result = db.Exec("INSERT INTO maplayer.map_organization (map_id, organization_id) VALUES (?, ?)", input.MapsetID, org.ID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	return c.SendStatus(fiber.StatusCreated)
}

func CreateAuthKey(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type APITokenInput struct {
		UserID string `json:"user_id"`
	}

	var input APITokenInput
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID required",
		})
	}

	var user database.User
	result := db.First(&user, "id = ?", input.UserID)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	authKey := database.AuthKey{
		UserID: user.ID,
	}

	result = db.Create(&authKey)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(authKey)
}
