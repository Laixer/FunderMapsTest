package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
)

func GetApplication(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	applicationID := c.Params("application_id", cfg.ApplicationID)

	if !strings.HasPrefix(applicationID, "app-") {
		applicationID = fmt.Sprintf("app-%s", applicationID)
	}

	var application database.Application
	result := db.First(&application, "application_id = ?", applicationID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Application not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(application)
}
