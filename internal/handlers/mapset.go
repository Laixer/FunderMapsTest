package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Mapset struct {
	ID    string `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Style string `json:"style"`
	// Layers  pq.StringArray `json:"layers" gorm:"type:text[]"`
	Options string  `json:"options" gorm:"type:jsonb"`
	Public  bool    `json:"public"`
	Consent *string `json:"consent"`
	Note    string  `json:"note"`
	Icon    *string `json:"icon"`
	// FenceNeighborhood []string    `json:"fence_neighborhood"`
	// FenceDistrict     []string    `json:"fence_district"`
	// FenceMunicipality []string    `json:"fence_municipality"`
	Layerset string `json:"layerset" gorm:"type:jsonb"`
}

func (u *Mapset) TableName() string {
	return "maplayer.mapset_collection"
}

func GetMapset(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	mapsetID := c.Params("mapset_id")

	if strings.HasPrefix(mapsetID, "cl") || strings.HasPrefix(mapsetID, "ck") {
		var mapset Mapset
		result := db.Where("public = true").First(&mapset, "id = ?", mapsetID)

		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
			})
		}

		return c.JSON(mapset)
	}

	var mapset Mapset
	result := db.Where("public = true").First(&mapset, "slug = ?", mapsetID)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(mapset)
}
