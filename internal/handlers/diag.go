package handlers

import "github.com/gofiber/fiber/v2"

func GetIP(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ip": c.IP()})
}

func GetHeaders(c *fiber.Ctx) error {
	return c.JSON(c.GetReqHeaders())
}
