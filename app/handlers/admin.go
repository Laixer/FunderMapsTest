package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
	"fundermaps/pkg/utils"
)

func GetAllOrganizations(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var orgs []database.Organization
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)
	result := db.Limit(limit).Offset(offset).Order("name ASC").Find(&orgs)
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
		Role   *string `json:"role" validate:"omitempty,oneof=reader writer verifier superuser"`
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

	if input.Role == nil {
		role := "reader"
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
