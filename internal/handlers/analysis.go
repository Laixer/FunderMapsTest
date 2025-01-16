package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/database"
	"fundermaps/internal/middleware"
)

func GetAnalysis(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	// user := c.Locals("user").(database.User)

	buildingID := c.Params("building_id")

	// if len(user.Organizations) == 0 {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
	// 		"message": "Forbidden",
	// 	})
	// }

	var analysis database.Analysis
	result := db.First(&analysis, "external_building_id = ?", buildingID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Analysis not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	c.Locals("tracker", middleware.ProductTracker{
		Name:       "analysis3",
		BuildingID: analysis.BuildingID,
		Identifier: buildingID,
	})

	// firstOrganization := user.Organizations[0]

	// TODO: Move this into middleware so we can use it in other handlers
	// var isRegistered bool
	// db.Raw(`
	// 	WITH register_product_request AS (
	// 		INSERT INTO application.product_tracker(organization_id, product, building_id, identifier)
	// 		SELECT ?, ?, ?, ?
	// 		WHERE NOT EXISTS (
	// 			SELECT  1
	// 			FROM    application.product_tracker pt
	// 			WHERE   pt.organization_id = ?
	// 			AND     pt.product = ?
	// 			AND     pt.identifier = ?
	// 			AND     pt.create_date > CURRENT_TIMESTAMP - interval '24 hours'
	// 		)
	// 		RETURNING 1
	// 	)
	// 	SELECT EXISTS (SELECT 1 FROM register_product_request) AS is_registered
	// `, firstOrganization.ID, "analysis3", analysis.BuildingID, buildingID, firstOrganization.ID, "analysis3", buildingID).Scan(&isRegistered)

	// c.Set("X-Product-Registered", fmt.Sprintf("%t", isRegistered))

	return c.JSON(analysis)
}
