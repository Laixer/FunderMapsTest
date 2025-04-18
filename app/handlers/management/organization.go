package mngmt

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
	limit := min(c.QueryInt("limit", 100), 100)
	offset := c.QueryInt("offset", 0)
	result := db.Limit(limit).Offset(offset).Order("name ASC").Find(&orgs)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(orgs)
}

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

	var existingOrg database.Organization
	if result := db.Where("name = ?", input.Name).First(&existingOrg); result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Organization with this name already exists"})
	}

	org := database.Organization{
		Name:  input.Name,
		Email: fmt.Sprintf("info@%s.com", utils.GenerateRandomString(10)), // TODO: Will be removed in the future
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

func UpdateOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	type OrganizationUpdateInput struct {
		Name              *string               `json:"name"`
		FenceMunicipality *database.StringArray `json:"fence_municipality"`
		FenceDistrict     *database.StringArray `json:"fence_district"`
		FenceNeighborhood *database.StringArray `json:"fence_neighborhood"`
	}

	var input OrganizationUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	// Update name if provided
	if input.Name != nil {
		// Check name uniqueness if it's being changed
		if *input.Name != org.Name {
			var existingOrg database.Organization
			if result := db.Where("name = ? AND id != ?", *input.Name, org.ID).First(&existingOrg); result.Error == nil {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Organization with this name already exists"})
			}
			org.Name = *input.Name
		}
	}

	// Update fence fields if provided
	if input.FenceMunicipality != nil {
		org.FenceMunicipality = *input.FenceMunicipality
	}

	if input.FenceDistrict != nil {
		org.FenceDistrict = *input.FenceDistrict
	}

	if input.FenceNeighborhood != nil {
		org.FenceNeighborhood = *input.FenceNeighborhood
	}

	result = db.Save(&org)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
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

	var count int64
	result = db.Table("application.organization_user").Where("user_id = ? AND organization_id = ?", user.ID, org.ID).Count(&count)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "User is already a member of this organization"})
	}

	if input.Role == nil || *input.Role == "" {
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
	var org database.Organization
	result := db.First(&org, "id = ?", c.Params("org_id"))
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Organization not found")
	}

	result = db.Exec("INSERT INTO maplayer.map_organization (map_id, organization_id) VALUES (?, ?)", input.MapsetID, org.ID)
	// TODO: This SQL statement can cause a unique constraint violation, handle this error
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
