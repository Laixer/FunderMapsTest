package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/internal/database"
)

func GetAllContractors(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var contractors []database.Contractor
	db.Order("id").Find(&contractors)

	return c.JSON(contractors)
}
