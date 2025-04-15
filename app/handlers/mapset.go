package handlers

import (
	"errors"
	"fundermaps/app/database"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetMapset(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	mapsetID := c.Params("mapset_id")

	if mapsetID == "" {
		var mapsets []database.Mapset

		var organizationIDs []uuid.UUID
		for _, org := range user.Organizations {
			organizationIDs = append(organizationIDs, org.ID)
		}

		// TODO: Move this in part to a database view
		result := db.Select("DISTINCT ON (maplayer.mapset_collection.id) maplayer.mapset_collection.*").
			Joins("JOIN maplayer.map_organization ON maplayer.map_organization.map_id = maplayer.mapset_collection.id").
			Where("maplayer.map_organization.organization_id IN (?)", organizationIDs).
			Find(&mapsets)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Mapset not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}

		return c.JSON(mapsets)
	}

	if strings.HasPrefix(mapsetID, "cl") || strings.HasPrefix(mapsetID, "ck") {
		var mapset database.Mapset
		result := db.Where("public = true").First(&mapset, "id = ?", mapsetID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Mapset not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}

		return c.JSON(mapset)
	}

	var mapset database.Mapset
	result := db.Where("public = true").First(&mapset, "slug = ?", mapsetID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Mapset not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(mapset)
}
