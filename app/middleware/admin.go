package middleware

import (
	"fundermaps/app/database"

	"github.com/gofiber/fiber/v2"
)

// TODO: Also check if user is active
func AdminMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(database.User)

	if user.Role != "administrator" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	return c.Next()
}
