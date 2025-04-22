package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/database"
)

func GetReport(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	buildingExternalID := c.Params("building_id")
	if buildingExternalID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Building ID is required",
		})
	}

	var incidents []database.Incident
	result := db.Joins("JOIN geocoder.building ON geocoder.building.id = report.incident.building").
		Where("geocoder.building.external_id = ?", buildingExternalID).
		Find(&incidents)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	var inquirySamples []database.InquirySample
	result = db.Joins("JOIN geocoder.building ON geocoder.building.id = report.inquiry_sample.building").
		Where("geocoder.building.external_id = ?", buildingExternalID).
		Find(&inquirySamples)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	var recoverySamples []database.RecoverySample
	result = db.Find(&recoverySamples, "building_id = ?", buildingExternalID)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"incidents": incidents,
		"inquiries": []fiber.Map{
			{
				"id":               145339,
				"documentName":     "Dossiernummer 1966-7049 Schiedam.pdf",
				"inspection":       false,
				"jointMeasurement": false,
				"floorMeasurement": false,
				"note":             nil,
				"documentDate":     "1966-11-21T00:00:00",
				"documentFile":     "39eb036b-5a41-498c-83b3-5937873a99e8.pdf",
				"type":             7,
				"standardF3o":      false,
				"attribution": fiber.Map{
					"id":             0,
					"reviewer":       "6c6b646e-3ad7-4373-9f79-c061abe2ab48",
					"reviewerName":   "schiedam@kcaf.nl",
					"creator":        "9fa1b0a9-4b04-4cdd-94e7-66fa5ee187ad",
					"creatorName":    "don@schiedam.nl",
					"owner":          "7ecb4f7a-75ce-4b2f-b9c2-68ddd502a5ae",
					"ownerName":      "Gemeente Schiedam",
					"contractor":     10,
					"contractorName": "FunderMaps B.V.",
				},
				"state": fiber.Map{
					"auditStatus": 4,
					"allowWrite":  false,
				},
				"access": fiber.Map{
					"accessPolicy": 1,
					"isPublic":     false,
					"isPrivate":    true,
				},
				"record": fiber.Map{
					"createDate": "2025-02-10T10:43:12.520961Z",
					"updateDate": "2025-02-10T10:44:48.067023Z",
					"deleteDate": nil,
				},
			},
		},
		"inquiry_samples":  inquirySamples,
		"recoveries":       []any{},
		"recovery_samples": recoverySamples,
	})
}
