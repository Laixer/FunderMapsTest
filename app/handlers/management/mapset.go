package mngmt

import (
	"fundermaps/app/config"
	"fundermaps/app/database"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetAllMapsets(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var mapsets []database.Mapset
	limit := min(c.QueryInt("limit", 100), 100)
	offset := c.QueryInt("offset", 0)
	result := db.Limit(limit).Offset(offset).Order("name ASC").Find(&mapsets)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(mapsets)
}

func GetMapsetByID(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var mapset database.Mapset
	result := db.First(&mapset, "id = ?", c.Params("mapset_id"))
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Mapset not found"})
	}

	return c.JSON(mapset)
}

func GetAllMapsetLayers(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var layers []database.MapsetLayer
	result := db.Order(`"order" ASC`).Find(&layers)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(layers)
}

func UpdateMapsetLayers(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	mapsetID := c.Params("mapset_id")
	if mapsetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Mapset ID is required"})
	}

	type LayersInput struct {
		Layers []string `json:"layers" validate:"required"`
	}

	var input LayersInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	if err := config.Validate.Struct(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var mapset database.Mapset
	if err := db.First(&mapset, "id = ?", mapsetID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Mapset not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if len(input.Layers) > 0 {
		var validIDs []string
		if err := db.Model(&database.MapsetLayer{}).Where("id IN ?", input.Layers).Pluck("id", &validIDs).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}
		known := make(map[string]struct{}, len(validIDs))
		for _, id := range validIDs {
			known[id] = struct{}{}
		}
		var unknown []string
		for _, id := range input.Layers {
			if _, ok := known[id]; !ok {
				unknown = append(unknown, id)
			}
		}
		if len(unknown) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unknown layer IDs",
				"unknown": unknown,
			})
		}
	}

	if err := db.Table("application.mapset").Where("id = ?", mapsetID).Update("layers", database.StringArray(input.Layers)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	mapset.Layers = database.StringArray(input.Layers)
	return c.JSON(mapset)
}
