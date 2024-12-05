package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/database"
)

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
