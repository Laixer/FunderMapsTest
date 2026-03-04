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
		Name *string `json:"name"`
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

// Geolock District Handlers

func GetOrganizationGeolockDistricts(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	var districts []database.OrganizationGeolockDistrict
	result = db.Where("organization_id = ?", org.ID).Find(&districts)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(districts)
}

func AddDistrictToOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type AddDistrictInput struct {
		DistrictID string `json:"district_id" validate:"required"`
	}

	var input AddDistrictInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	// Check if the relationship already exists
	var existing database.OrganizationGeolockDistrict
	result = db.Where("organization_id = ? AND district_id = ?", org.ID, input.DistrictID).First(&existing)
	if result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "District is already associated with this organization"})
	}

	// Create the relationship
	geolockDistrict := database.OrganizationGeolockDistrict{
		OrganizationID: org.ID,
		DistrictID:     input.DistrictID,
	}

	result = db.Create(&geolockDistrict)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(geolockDistrict)
}

func RemoveDistrictFromOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type RemoveDistrictInput struct {
		DistrictID string `json:"district_id" validate:"required"`
	}

	var input RemoveDistrictInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	result = db.Where("organization_id = ? AND district_id = ?", org.ID, input.DistrictID).
		Delete(&database.OrganizationGeolockDistrict{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "District association not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Geolock Municipality Handlers

func GetOrganizationGeolockMunicipalities(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	var municipalities []database.OrganizationGeolockMunicipality
	result = db.Where("organization_id = ?", org.ID).Find(&municipalities)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(municipalities)
}

func AddMunicipalityToOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type AddMunicipalityInput struct {
		MunicipalityID string `json:"municipality_id" validate:"required"`
	}

	var input AddMunicipalityInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	// Check if the relationship already exists
	var existing database.OrganizationGeolockMunicipality
	result = db.Where("organization_id = ? AND municipality_id = ?", org.ID, input.MunicipalityID).First(&existing)
	if result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Municipality is already associated with this organization"})
	}

	// Create the relationship
	geolockMunicipality := database.OrganizationGeolockMunicipality{
		OrganizationID: org.ID,
		MunicipalityID: input.MunicipalityID,
	}

	result = db.Create(&geolockMunicipality)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(geolockMunicipality)
}

func RemoveMunicipalityFromOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type RemoveMunicipalityInput struct {
		MunicipalityID string `json:"municipality_id" validate:"required"`
	}

	var input RemoveMunicipalityInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	result = db.Where("organization_id = ? AND municipality_id = ?", org.ID, input.MunicipalityID).
		Delete(&database.OrganizationGeolockMunicipality{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Municipality association not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Geolock Neighborhood Handlers

func GetOrganizationGeolockNeighborhoods(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	var neighborhoods []database.OrganizationGeolockNeighborhood
	result = db.Where("organization_id = ?", org.ID).Find(&neighborhoods)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(neighborhoods)
}

func AddNeighborhoodToOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type AddNeighborhoodInput struct {
		NeighborhoodID string `json:"neighborhood_id" validate:"required"`
	}

	var input AddNeighborhoodInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	// Check if the relationship already exists
	var existing database.OrganizationGeolockNeighborhood
	result = db.Where("organization_id = ? AND neighborhood_id = ?", org.ID, input.NeighborhoodID).First(&existing)
	if result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Neighborhood is already associated with this organization"})
	}

	// Create the relationship
	geolockNeighborhood := database.OrganizationGeolockNeighborhood{
		OrganizationID: org.ID,
		NeighborhoodID: input.NeighborhoodID,
	}

	result = db.Create(&geolockNeighborhood)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(geolockNeighborhood)
}

func RemoveNeighborhoodFromOrganization(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	organizationID := c.Params("org_id")

	type RemoveNeighborhoodInput struct {
		NeighborhoodID string `json:"neighborhood_id" validate:"required"`
	}

	var input RemoveNeighborhoodInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var org database.Organization
	result := db.First(&org, "id = ?", organizationID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Organization not found"})
	}

	result = db.Where("organization_id = ? AND neighborhood_id = ?", org.ID, input.NeighborhoodID).
		Delete(&database.OrganizationGeolockNeighborhood{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Neighborhood association not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
