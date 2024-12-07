package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/pkg/utils"
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

	// TODO: Move into a platform service
	getBuilding := func(BuildingID string) (Building, error) {
		switch utils.FromIdentifier(BuildingID) {
		case utils.NlBagBuilding:
			var building Building

			result := db.First(&building, "external_id = ?", BuildingID)
			return building, result.Error

		case utils.NlBagLegacyBuilding:
			var building Building

			result := db.First(&building, "external_id = 'NL.IMBAG.PAND.' || ?", BuildingID)
			return building, result.Error

		case utils.NlBagAddress:
			var building Building

			result := db.Joins("JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id").
				Where("geocoder.address.external_id = ?", BuildingID).
				First(&building)
			return building, result.Error

		case utils.NlBagLegacyAddress:
			var building Building

			result := db.Joins("JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id").
				Where("geocoder.address.external_id = 'NL.IMBAG.NUMMERAANDUIDING.' || ?", BuildingID).
				First(&building)
			return building, result.Error
		}

		return Building{}, errors.New("unknown geocoder identifier")
	}

	building, err := getBuilding(BuildingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Building not found",
			})
			// TODO: This is ugly
		} else if err.Error() == "unknown geocoder identifier" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unknown geocoder identifier",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(building)
}
