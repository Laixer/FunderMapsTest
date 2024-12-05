package handlers

import (
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

func GetGeocoder(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	BuildingID := c.Params("building_id")

	var building Building
	result := db.First(&building, "external_id = ?", BuildingID)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(building)
}
