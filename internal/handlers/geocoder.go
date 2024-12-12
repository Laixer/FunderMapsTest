package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/pkg/utils"
)

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

func (b *BuildingGeocoder) TableName() string {
	return "geocoder.building_geocoder"
}

func GetGeocoder(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	geocoderID := c.Params("geocoder_id")

	// TODO: Move into a platform service
	getBuilding := func(geocoderID string) (BuildingGeocoder, error) {
		// TODO: Normalize the geocoder identifier

		switch utils.FromIdentifier(geocoderID) {
		case utils.NlBagBuilding:
			var buildingGeocoder BuildingGeocoder

			result := db.First(&buildingGeocoder, "building_id = ?", geocoderID)
			return buildingGeocoder, result.Error

		case utils.NlBagLegacyBuilding:
			var buildingGeocoder BuildingGeocoder

			result := db.First(&buildingGeocoder, "building_id = 'NL.IMBAG.PAND.' || ?", geocoderID)
			return buildingGeocoder, result.Error

		case utils.NlBagAddress:
			var buildingGeocoder BuildingGeocoder

			result := db.Joins("join geocoder.building on geocoder.building.external_id = geocoder.building_geocoder.building_id").
				Joins("JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id").
				Where("geocoder.address.external_id = ?", geocoderID).
				First(&buildingGeocoder)
			return buildingGeocoder, result.Error

		case utils.NlBagLegacyAddress:
			var buildingGeocoder BuildingGeocoder

			result := db.Joins("join geocoder.building on geocoder.building.external_id = geocoder.building_geocoder.building_id").
				Joins("JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id").
				Where("geocoder.address.external_id = 'NL.IMBAG.NUMMERAANDUIDING.' || ?", geocoderID).
				First(&buildingGeocoder)
			return buildingGeocoder, result.Error
		}

		return BuildingGeocoder{}, errors.New("unknown geocoder identifier")
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

type Address struct {
	ID             string `json:"-" gorm:"primaryKey"`
	ExternalID     string `json:"id"`
	BuildingID     string `json:"-"`
	BuildingNumber string `json:"building_number"`
	PostalCode     string `json:"postal_code"`
	Street         string `json:"street"`
	City           string `json:"city"`
}

func (a *Address) TableName() string {
	return "geocoder.address"
}

func GetAllAddresses(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	geocoderID := c.Params("geocoder_id")

	// TODO: Implement the other geocoder identifiers

	var addresses []Address
	result := db.Joins("JOIN geocoder.building ON geocoder.building.id = geocoder.address.building_id").
		Where("geocoder.building.external_id = ?", geocoderID).
		Find(&addresses)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(addresses)
}
