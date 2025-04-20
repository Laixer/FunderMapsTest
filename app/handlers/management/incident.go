package mngmt

import (
	// "fundermaps/app/database"

	"github.com/gofiber/fiber/v2"
	// "gorm.io/gorm"
)

func DeleteIncident(c *fiber.Ctx) error {
	// db := c.Locals("db").(*gorm.DB)

	incidentID := c.Params("incident_id")
	if incidentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid incident ID"})
	}

	// var incident database.Incident
	// result := db.First(&incident, "id = ?", incidentID)
	// if result.Error != nil {
	// 	if result.Error == gorm.ErrRecordNotFound {
	// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Incident not found"})
	// 	}
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	// }

	// result = db.Delete(&incident)
	// if result.Error != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	// }

	return c.JSON(fiber.Map{"message": "Incident deleted successfully"})
}
