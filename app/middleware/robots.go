package middleware

import "github.com/gofiber/fiber/v2"

// RobotsMiddleware handles requests for robots.txt file
func RobotsMiddleware(robotsPath string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/robots.txt" {
			c.Set(fiber.HeaderCacheControl, "public, max-age=86400")
			return c.SendFile(robotsPath)
		}
		return c.Next()
	}
}
