package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/database"
)

func GetReport(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	buildingExternalID := c.Params("building_id")
	if buildingExternalID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Building ID is required",
		})
	}

	var incidents []database.Incident
	result := db.Joins("JOIN geocoder.building ON geocoder.building.id = report.incident.building").
		Where("geocoder.building.external_id = ?", buildingExternalID).
		Find(&incidents)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	var inquirySamples []database.InquirySample
	result = db.Joins("JOIN geocoder.building ON geocoder.building.id = report.inquiry_sample.building").
		Where("geocoder.building.external_id = ?", buildingExternalID).
		Find(&inquirySamples)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	var recoverySamples []database.RecoverySample
	result = db.Find(&recoverySamples, "building_id = ?", buildingExternalID)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"incidents":        incidents,
		"inquiries":        []any{},
		"inquiry_samples":  inquirySamples,
		"recoveries":       []any{},
		"recovery_samples": recoverySamples,
	})
}
