package mngmt

import (
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
