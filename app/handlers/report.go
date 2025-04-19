package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func GetReport(c *fiber.Ctx) error {
	buildingID := c.Params("building_id")
	if buildingID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Building ID is required",
		})
	}

	return c.JSON(map[string]any{
		"building_id": buildingID,
	})
}
