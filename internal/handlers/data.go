package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type BuildingSubsidence struct {
	BuildingID string  `json:"building_id" gorm:"primaryKey"`
	Velocity   float64 `json:"velocity"`
}

func (a *BuildingSubsidence) TableName() string {
	return "data.building_subsidence"
}

func GetDataSubsidence(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	buildingID := c.Params("building_id")

	var subsidence BuildingSubsidence
	result := db.First(&subsidence, "building_id = ?", buildingID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Subsidence not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(subsidence)
}

type BuildingSubsidenceHistory struct {
	BuildingID string  `json:"building_id" gorm:"primaryKey"`
	Velocity   float64 `json:"velocity" gorm:"primaryKey"`
	MarkAt     string  `json:"mark_at" gorm:"primaryKey"`
}

func (a *BuildingSubsidenceHistory) TableName() string {
	return "data.subsidence_history"
}

func GetDataSubsidenceHistoric(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	buildingID := c.Params("building_id")

	var subsidenceHistory []BuildingSubsidenceHistory
	result := db.Find(&subsidenceHistory, "building_id = ?", buildingID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Subsidence history not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(subsidenceHistory)
}
