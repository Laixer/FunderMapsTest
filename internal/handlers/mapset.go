package handlers

import (
	"errors"
	"fundermaps/internal/database"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetMapset(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	mapsetID := c.Params("mapset_id")

	if mapsetID == "" {
		var mapsets []database.Mapset
		// TODO: Select mapsets based on organization
		result := db.Where("public = true").Find(&mapsets)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Mapset not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}
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
