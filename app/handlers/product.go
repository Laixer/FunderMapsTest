package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/database"
	"fundermaps/app/platform/geocoder"
)

func GetAnalysis(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	buildingID := c.Params("building_id")

	var analysis database.Analysis
	result := db.Select("external_building_id AS building_id, neighborhood_id, construction_year, construction_year_reliability, foundation_type, foundation_type_reliability, restoration_costs, drystand, drystand_risk, drystand_risk_reliability, bio_infection_risk, bio_infection_risk_reliability, dewatering_depth, dewatering_depth_risk, dewatering_depth_risk_reliability, unclassified_risk, height, velocity, ground_water_level, ground_level, soil, surface_area, owner, inquiry_id, inquiry_type, damage_cause, enforcement_term, overall_quality, recovery_type").
		First(&analysis, "external_building_id = ?", buildingID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Analysis not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	c.Locals("tracker", database.ProductTracker{
		Name:       "analysis3",
		BuildingID: analysis.BuildingID,
		Identifier: buildingID,
	})

	return c.JSON(analysis)
}

// FoundationTypeDistributionItem holds data for foundation type distribution.
type FoundationTypeDistributionItem struct {
	FoundationType string  `json:"foundation_type"`
	Percentage     float64 `json:"percentage"`
}

// ConstructionYearDistributionItem holds data for construction year distribution.
type ConstructionYearDistributionItem struct {
	YearFrom int `json:"year_from"`
	Count    int `json:"count"`
}

// FoundationRiskDistributionItem holds data for foundation risk distribution.
type FoundationRiskDistributionItem struct {
	FoundationRisk string  `json:"foundation_risk"`
	Percentage     float64 `json:"percentage"`
}

// IncidentCountItem holds data for incident counts per year.
type IncidentCountItem struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

// NeighborhoodStatisticsResponse is the combined response for neighborhood statistics.
type NeighborhoodStatisticsResponse struct {
	FoundationTypeDistribution   []FoundationTypeDistributionItem   `json:"foundation_type_distribution"`
	ConstructionYearDistribution []ConstructionYearDistributionItem `json:"construction_year_distribution"`
	DataCollectedPercentage      float64                            `json:"data_collected_percentage"`
	FoundationRiskDistribution   []FoundationRiskDistributionItem   `json:"foundation_risk_distribution"`
	BuildingRestoredCount        int                                `json:"building_restored_count"`
	IncidentCounts               []IncidentCountItem                `json:"incident_counts"`
	NeighborhoodReportCounts     []IncidentCountItem                `json:"neighborhood_report_counts"`
	MunicipalityIncidentCounts   []IncidentCountItem                `json:"municipality_incident_counts"`
	MunicipalityReportCounts     []IncidentCountItem                `json:"municipality_report_counts"`
}

func GetStatistics(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	buildingID := c.Params("building_id")

	geocoderService := geocoder.NewService(db)

	// 0. Validate buildingID
	building, err := geocoderService.GetBuildingByGeocoderID(buildingID)
	if err != nil {
		if err.Error() == "building not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Building not found"})
		} else if err.Error() == "unknown geocoder identifier" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Unknown geocoder identifier"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	response := NeighborhoodStatisticsResponse{}

	// 2. Fetch FoundationTypeDistribution
	sqlFoundation := `
		SELECT  spft.foundation_type,
				round(spft.percentage::numeric, 2) as percentage
		FROM    data.statistics_product_foundation_type AS spft
		WHERE   spft.neighborhood_id = ?`
	if err := db.Raw(sqlFoundation, building.NeighborhoodID).Scan(&response.FoundationTypeDistribution).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error fetching foundation type distribution"})
	}

	// 3. Fetch ConstructionYearDistribution
	sqlConstruction := `
		SELECT  spcy.year_from,
				spcy.count
		FROM    data.statistics_product_construction_years AS spcy
		WHERE   spcy.neighborhood_id = ?`
	if err := db.Raw(sqlConstruction, building.NeighborhoodID).Scan(&response.ConstructionYearDistribution).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error fetching construction year distribution"})
	}

	// 4. Fetch DataCollectedPercentage
	sqlDataCollected := `
		SELECT  round(spdc.percentage::numeric, 2)
		FROM    data.statistics_product_data_collected AS spdc
		WHERE   spdc.neighborhood_id = ?
		LIMIT   1`

	// Use a temporary variable to scan, as Scan expects a pointer.
	// If no record is found, Raw().Scan() might return an error or leave the variable as zero-value.
	// We'll default to 0.0 if there's an error or no record.
	var dataCollectedPercentage float64
	if err := db.Raw(sqlDataCollected, building.NeighborhoodID).Scan(&dataCollectedPercentage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no record found, it's okay, percentage is 0.
			response.DataCollectedPercentage = 0.0
		} else {
			// Log error: e.g., log.Printf("Error fetching data collected percentage: %v", err)
			// For other errors, we might still want to return 0 or an error response.
			// Depending on requirements, you might choose to return an error here.
			// For now, we'll set to 0 and continue, or you could return:
			// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error fetching data collected percentage"})
			response.DataCollectedPercentage = 0.0 // Default to 0 on error
		}
	} else {
		response.DataCollectedPercentage = dataCollectedPercentage
	}

	// 5. Fetch FoundationRiskDistribution
	sqlFoundationRisk := `
		SELECT  spfr.foundation_risk,
				round(spfr.percentage::numeric, 2) as percentage
		FROM    data.statistics_product_foundation_risk AS spfr
		WHERE   spfr.neighborhood_id = ?`
	if err := db.Raw(sqlFoundationRisk, building.NeighborhoodID).Scan(&response.FoundationRiskDistribution).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error fetching foundation risk distribution"})
	}

	// 6. Fetch BuildingRestoredCount
	sqlBuildingRestored := `
		SELECT  spbr.count
		FROM    data.statistics_product_buildings_restored AS spbr
		WHERE   spbr.neighborhood_id = ?
		LIMIT   1`
	var buildingRestoredCount int
	if err := db.Raw(sqlBuildingRestored, building.NeighborhoodID).Scan(&buildingRestoredCount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.BuildingRestoredCount = 0
		} else {
			response.BuildingRestoredCount = 0 // Default to 0 on error
		}
	} else {
		response.BuildingRestoredCount = buildingRestoredCount
	}

	// 7. Fetch IncidentCounts (Neighborhood)
	sqlIncidentCounts := `
		SELECT  spi.year,
				spi.count
		FROM    data.statistics_product_incidents AS spi
		WHERE   spi.neighborhood_id = ?`
	if err := db.Raw(sqlIncidentCounts, building.NeighborhoodID).Scan(&response.IncidentCounts).Error; err != nil {
		// Log error
		// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error fetching neighborhood incident counts"})
	}

	// 8. Fetch MunicipalityIncidentCounts
	sqlMunicipalityIncidentCounts := `
		SELECT  spim.year,
				spim.count
		FROM    data.statistics_product_incident_municipality spim
		WHERE   spim.municipality_id = ?`
	if err := db.Raw(sqlMunicipalityIncidentCounts, building.MunicipalityID).Scan(&response.MunicipalityIncidentCounts).Error; err != nil {
		// Log error
	}

	// 9. Fetch NeighborhoodReportCounts (from statistics_product_inquiries)
	sqlNeighborhoodReportCounts := `
		SELECT  spi.year,
				spi.count
		FROM    data.statistics_product_inquiries AS spi
		WHERE   spi.neighborhood_id = ?`
	if err := db.Raw(sqlNeighborhoodReportCounts, building.NeighborhoodID).Scan(&response.NeighborhoodReportCounts).Error; err != nil {
		// Log error
	}

	// 10. Fetch MunicipalityReportCounts (from statistics_product_inquiry_municipality)
	sqlMunicipalityReportCounts := `
		SELECT  spim.year,
				spim.count
		FROM    data.statistics_product_inquiry_municipality spim
		WHERE   spim.municipality_id = ?`
	if err := db.Raw(sqlMunicipalityReportCounts, building.MunicipalityID).Scan(&response.MunicipalityReportCounts).Error; err != nil {
		// Log error: e.g., log.Printf("Error fetching municipality report counts: %v", err)
	}

	return c.JSON(response)
}
