package handlers

import (
	"errors"
	"time"

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

// -- geocoder.building_geocoder source

// CREATE OR REPLACE VIEW geocoder.building_geocoder
// AS SELECT b.built_year AS building_built_year,
//     b.external_id AS building_id,
//     b.building_type,
//     b.zone_function AS building_zone_function,
//     r.id AS residence_id,
//     st_y(r.geom) AS residence_lat,
//     st_x(r.geom) AS residence_lon,
//     n.external_id AS neighborhood_id,
//     n.name AS neighborhood_name,
//     d.external_id AS district_id,
//     d.name AS district_name,
//     m.external_id AS municipality_id,
//     m.name AS municipality_name,
//     s.external_id AS state_id,
//     s.name AS state_name

type BuildingGeocoder struct {
	BuildingBuiltYear time.Time `json:"building_built_year"`
	BuildingID        string    `json:"building_id"`
	BuildingType      string    `json:"building_type"`
	// BuildingZoneFunction string `json:"building_zone_function"` // array
	ResidenceID      string  `json:"residence_id"`
	ResidenceLat     float32 `json:"residence_lat"`
	ResidenceLon     float32 `json:"residence_lon"`
	NeighborhoodID   string  `json:"neighborhood_id"`
	NeighborhoodName string  `json:"neighborhood_name"`
	DistrictID       string  `json:"district_id"`
	DistrictName     string  `json:"district_name"`
	MunicipalityID   string  `json:"municipality_id"`
	MunicipalityName string  `json:"municipality_name"`
	StateID          string  `json:"state_id"`
	StateName        string  `json:"state_name"`
}

func GetGeocoder(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	geocoderID := c.Params("geocoder_id")

	// TODO: Move into a platform service
	getBuilding := func(geocoderID string) (Building, error) {
		// TODO: Normalize the geocoder identifier

		switch utils.FromIdentifier(geocoderID) {
		case utils.NlBagBuilding:
			var building Building

			result := db.First(&building, "external_id = ?", geocoderID)
			return building, result.Error

		case utils.NlBagLegacyBuilding:
			var building Building

			result := db.First(&building, "external_id = 'NL.IMBAG.PAND.' || ?", geocoderID)
			return building, result.Error

		case utils.NlBagAddress:
			var building Building

			result := db.Joins("JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id").
				Where("geocoder.address.external_id = ?", geocoderID).
				First(&building)
			return building, result.Error

		case utils.NlBagLegacyAddress:
			var building Building

			result := db.Joins("JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id").
				Where("geocoder.address.external_id = 'NL.IMBAG.NUMMERAANDUIDING.' || ?", geocoderID).
				First(&building)
			return building, result.Error
		}

		return Building{}, errors.New("unknown geocoder identifier")
	}

	building, err := getBuilding(geocoderID)
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
