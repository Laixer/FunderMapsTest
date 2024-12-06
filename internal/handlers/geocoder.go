package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// var building = await geocoderTranslation.GetBuildingIdAsync(id);
// var address = await geocoderTranslation.GetAddressIdAsync(building.ExternalId);
// var residence = await geocoderTranslation.GetResidenceIdAsync(building.ExternalId);
// var neighborhood = building.NeighborhoodId is not null
// 	? await geocoderTranslation.GetNeighborhoodIdAsync(building.NeighborhoodId)
// 	: null;
// var district = neighborhood!.DistrictId is not null
// 	? await geocoderTranslation.GetDistrictIdAsync(neighborhood.DistrictId)
// 	: null;
// var municipality = district!.MunicipalityId is not null
// 	? await geocoderTranslation.GetMunicipalityIdAsync(district.MunicipalityId)
// 	: null;
// var state = municipality!.StateId is not null
// 	? await geocoderTranslation.GetStateIdAsync(municipality.StateId)
// 	: null;

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

// type Residence struct {
// 	ID string `json:"id" gorm:"primaryKey"`
// 	// AddressID  string  `json:"address_id"`
// 	BuildingID string  `json:"building_id"`
// 	Longitude  float64 `json:"longitude"`
// 	Latitude   float64 `json:"latitude"`
// }

// NL.IMBAG.NUMMERAANDUIDING.0202200000386458

func GetGeocoder(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	BuildingID := c.Params("building_id")

	var building Building

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
