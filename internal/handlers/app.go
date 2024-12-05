package handlers

import (
	"errors"
	"fmt"
	"fundermaps/internal/database"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetApplication(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	applicationID := c.Params("application_id")

	if !strings.HasPrefix(applicationID, "app-") {
		applicationID = fmt.Sprintf("app-%s", applicationID)
	}

	var application database.Application
	result := db.First(&application, "application_id = ?", applicationID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Application not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(application)
}
