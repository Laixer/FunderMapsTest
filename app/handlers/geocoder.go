package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/database"
	"fundermaps/app/platform/geocoder"
)

func GetGeocoder(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	geocoderService := geocoder.NewService(db)

	geocoderID := c.Params("geocoder_id") // TODO: Validate the geocoder identifier

	building, err := geocoderService.GetBuildingByGeocoderID(geocoderID)
	if err != nil {
		if err.Error() == "building not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Building not found"})
		} else if err.Error() == "unknown geocoder identifier" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Unknown geocoder identifier"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(building)
}

func GetAllAddresses(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	geocoderID := c.Params("geocoder_id")

	// TODO: Implement the other geocoder identifiers

	var addresses []database.Address
	result := db.Joins("JOIN geocoder.building ON geocoder.building.id = geocoder.address.building_id").
		Where("geocoder.building.external_id = ?", geocoderID).
		Find(&addresses)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Address not found"})
		} else if result.Error.Error() == "unknown geocoder identifier" { // TODO: This is ugly
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Unknown geocoder identifier"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(addresses)
}
