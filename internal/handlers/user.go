package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/database"
)

const ApplicationID = "app-0blu4s39"

func GetCurrentUser(c *fiber.Ctx) error {
	user := c.Locals("user").(database.User)

	return c.JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	var input database.User
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(user)
}

func GetCurrentUserMetadata(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	type Metadata struct {
		Metadata   string    `json:"metadata" gorm:"type:jsonb"`
		UpdateDate time.Time `json:"update_date"`
	}

	// TODO: Use gorm instead of raw query
	var metadata Metadata
	db.Raw(`
		SELECT metadata, update_date
		FROM application.application_user
		WHERE user_id = ?
		AND application_id = ?
		LIMIT 1`, user.ID, ApplicationID).Scan(&metadata)

	return c.JSON(metadata)
}
