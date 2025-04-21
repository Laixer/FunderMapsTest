package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
)

func GetCurrentUser(c *fiber.Ctx) error {
	user := c.Locals("user").(database.User)

	return c.JSON(user)
}

func UpdateCurrentUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	var input database.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	updateNullableString := func(target **string, value *string) {
		if value != nil {
			if *value != "" {
				*target = value
			} else {
				*target = nil
			}
		}
	}

	updateNullableString(&user.GivenName, input.GivenName)
	updateNullableString(&user.LastName, input.LastName)
	updateNullableString(&user.Avatar, input.Avatar)
	updateNullableString(&user.JobTitle, input.JobTitle)
	updateNullableString(&user.PhoneNumber, input.PhoneNumber)

	result := db.Save(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(user)
}

func GetCurrentUserMetadata(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	// Get application ID from query string if available, otherwise use config value
	applicationID := cfg.ApplicationID
	if queryAppID := c.Query("app_id"); queryAppID != "" {
		applicationID = queryAppID
	}

	var applicationUser database.ApplicationUser
	result := db.First(&applicationUser, "user_id = ? AND application_id = ?", user.ID, applicationID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return c.JSON(database.ApplicationUser{})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(applicationUser)
}

func UpdateCurrentUserMetadata(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	// Get application ID from query string if available, otherwise use config value
	applicationID := cfg.ApplicationID
	if queryAppID := c.Query("app_id"); queryAppID != "" {
		applicationID = queryAppID
	}

	var input struct {
		Metadata map[string]any `json:"metadata"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	var applicationUser database.ApplicationUser
	result := db.Where("user_id = ? AND application_id = ?", user.ID.String(), applicationID).FirstOrCreate(&applicationUser, database.ApplicationUser{
		UserID:        user.ID.String(),
		ApplicationID: applicationID,
	})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	applicationUser.Metadata = input.Metadata
	applicationUser.UpdateDate = time.Now()

	if err := db.Save(&applicationUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
