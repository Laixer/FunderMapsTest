package mngmt

import (
	"fundermaps/app/config"
	"fundermaps/app/database"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetAllApplications(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var apps []database.Application
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)
	result := db.Limit(limit).Offset(offset).Order("name ASC").Find(&apps)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(apps)
}

func CreateApplication(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	type ApplicationInput struct {
		Name string `json:"name" validate:"required"`
	}

	var input ApplicationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	app := database.Application{
		Name: input.Name,
	}

	result := db.Create(&app)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(app)
}
