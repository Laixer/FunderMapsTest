package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/database"
)

func GetAllContractors(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var contractors []database.Contractor
	db.Order("id").Find(&contractors)

	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(contractors)
}
