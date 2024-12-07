package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Building struct {
	ID             string `json:"-" gorm:"primaryKey"`
	BuiltYear      string `json:"built_year"`
	IsActive       bool   `json:"is_active"`
	ExternalID     string `json:"external_id"`
	NeighborhoodID string `json:"neighborhood_id"`
}

func (b *Building) TableName() string {
	return "geocoder.building"
}

func GetGeocoder(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	// TODO: Not always a building_id
	BuildingID := c.Params("building_id")

	var building Building

	// TODO: There are other types of building IDs
	if strings.HasPrefix(BuildingID, "NL.IMBAG.NUMMERAANDUIDING") {
		result := db.Joins("JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id").
			Where("geocoder.address.external_id = ?", BuildingID).
			First(&building)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "Building not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
			})
		}
	} else {
		result := db.First(&building, "external_id = ?", BuildingID)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "Building not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
			})
		}
	}

	return c.JSON(building)
}
