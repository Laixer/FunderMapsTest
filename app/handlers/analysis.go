package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/database"
)

func GetAnalysis(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	buildingID := c.Params("building_id")

	var analysis database.Analysis
	result := db.First(&analysis, "external_building_id = ?", buildingID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Analysis not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	c.Locals("tracker", database.ProductTracker{
		Name:       "analysis3",
		BuildingID: analysis.BuildingID,
		Identifier: buildingID,
	})

	return c.JSON(analysis)
}
