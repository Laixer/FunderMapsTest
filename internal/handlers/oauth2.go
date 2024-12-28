package handlers

import (
	"fundermaps/internal/database"

	"github.com/gofiber/fiber/v2"
)

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

	// TODO: Create a struct for this
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
