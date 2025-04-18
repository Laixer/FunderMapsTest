package mngmt

import (
	"fundermaps/app/config"
	"fundermaps/app/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetAllApplications(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var apps []database.Application
	limit := min(c.QueryInt("limit", 100), 100)
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

// TODO: Not tested
func UpdateApplication(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	appID := c.Params("app_id")
	if appID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Application ID is required"})
	}

	var app database.Application
	result := db.First(&app, "id = ?", appID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Application not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	type ApplicationInput struct {
		Name        string              `json:"name" validate:"required"`
		Data        database.JSONObject `json:"data"`
		RedirectURL string              `json:"redirect_url"`
		Public      bool                `json:"public"`
		UserID      uuid.UUID           `json:"user_id"`
	}

	var input ApplicationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	err := config.Validate.Struct(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	app.Name = input.Name
	app.Data = input.Data
	app.RedirectURL = input.RedirectURL
	app.Public = input.Public
	app.UserID = input.UserID

	result = db.Save(&app)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(app)
}

func GetApplication(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	appID := c.Params("app_id")
	if appID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Application ID is required"})
	}

	var app database.Application
	result := db.First(&app, "id = ?", appID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Application not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(app)
}
