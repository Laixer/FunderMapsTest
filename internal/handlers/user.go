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

func GetUserInfo(c *fiber.Ctx) error {
	user := c.Locals("user").(database.User)

	var name string
	if user.GivenName != nil {
		name = *user.GivenName
	}
	if user.LastName != nil {
		if name != "" {
			name += " "
		}
		name += *user.LastName
	}

	userInfo := fiber.Map{
		"sub":          user.ID,
		"name":         name, // TODO: Should be null if both given_name and family_name are null
		"given_name":   user.GivenName,
		"family_name":  user.LastName,
		"email":        user.Email,
		"picture":      user.Avatar,
		"phone_number": user.PhoneNumber,
	}
	return c.JSON(userInfo)
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

func UpdateCurrentUserMetadata(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	var input struct {
		Metadata interface{} `json:"metadata"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	result := db.Exec(`
		INSERT INTO application.application_user (user_id, application_id, metadata, update_date)
		VALUES (?, ?, ?, now())
		ON CONFLICT (user_id, application_id)
		DO UPDATE SET metadata = excluded.metadata, update_date = excluded.update_date;`,
		user.ID, ApplicationID, input.Metadata)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
