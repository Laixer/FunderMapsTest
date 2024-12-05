package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/database"
)

// TODO: Needs 'service' role
func GetAnalysis(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	buildingID := c.Params("building_id")

	var analysis database.Analysis
	result := db.First(&analysis, "external_building_id = ?", buildingID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Analysis not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(analysis)
}
