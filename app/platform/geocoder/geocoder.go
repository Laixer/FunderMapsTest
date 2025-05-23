package geocoder

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"fundermaps/pkg/utils"
)

type BuildingGeocoder struct {
	BuildingBuiltYear time.Time `json:"building_built_year"` // TODO: Change to int
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

type GeocoderService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *GeocoderService {
	return &GeocoderService{db: db}
}

// GetBuildingByGeocoderID retrieves building information based on the provided geocoder identifier.
// It supports multiple identifier formats including NlBagBuilding, NlBagLegacyBuilding, NlBagAddress, and NlBagLegacyAddress.
func (s *GeocoderService) GetBuildingByGeocoderID(geocoderID string) (*BuildingGeocoder, error) {
	getBuilding := func(geocoderID string) (BuildingGeocoder, error) {
		idType := utils.FromIdentifier(geocoderID)

		var buildingGeocoder BuildingGeocoder
		var result *gorm.DB

		switch idType {
		case utils.NlBagBuilding:
			result = s.db.First(&buildingGeocoder, "building_id = ?", geocoderID)

		case utils.NlBagLegacyBuilding:
			result = s.db.First(&buildingGeocoder, "building_id = 'NL.IMBAG.PAND.' || ?", geocoderID)

		case utils.NlBagAddress, utils.NlBagLegacyAddress:
			query := "JOIN geocoder.building ON geocoder.building.external_id = geocoder.building_geocoder.building_id " +
				"JOIN geocoder.address ON geocoder.address.building_id = geocoder.building.id"

			whereClause := "geocoder.address.external_id = ?"
			params := []any{geocoderID}

			if idType == utils.NlBagLegacyAddress {
				whereClause = "geocoder.address.external_id = 'NL.IMBAG.NUMMERAANDUIDING.' || ?"
			}

			result = s.db.Joins(query).Where(whereClause, params...).First(&buildingGeocoder)

		default:
			return BuildingGeocoder{}, errors.New("unknown geocoder identifier")
		}

		return buildingGeocoder, result.Error
	}

	building, err := getBuilding(geocoderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("building not found")
		} else if err.Error() == "unknown geocoder identifier" {
			return nil, err
		}
		return nil, err
	}
	return &building, nil
}

func (s *GeocoderService) GetOldBuildingID(buildingID string) (string, error) {
	var oldBuildingID string
	result := s.db.Raw("SELECT id FROM geocoder.building WHERE external_id = ? LIMIT 1", buildingID).Scan(&oldBuildingID)
	return oldBuildingID, result.Error
}
