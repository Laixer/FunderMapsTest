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
	result := db.Select("external_building_id AS building_id, neighborhood_id, construction_year, construction_year_reliability, foundation_type, foundation_type_reliability, restoration_costs, drystand, drystand_risk, drystand_risk_reliability, bio_infection_risk, bio_infection_risk_reliability, dewatering_depth, dewatering_depth_risk, dewatering_depth_risk_reliability, unclassified_risk, height, velocity, ground_water_level, ground_level, soil, surface_area, owner, inquiry_id, inquiry_type, damage_cause, enforcement_term, overall_quality, recovery_type").
		First(&analysis, "external_building_id = ?", buildingID)
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
