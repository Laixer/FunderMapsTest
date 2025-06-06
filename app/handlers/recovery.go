package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
)

type CreateRecoveryInput struct {
	// Fields for Recovery
	Note         *string   `json:"note"`
	AccessPolicy string    `json:"access_policy" validate:"required,oneof=public private"`
	Type         string    `json:"type" validate:"required"`
	DocumentDate time.Time `json:"document_date" validate:"required"` // TODO: This is a date, not a datetime
	DocumentFile string    `json:"document_file" validate:"required"`
	DocumentName string    `json:"document_name" validate:"required"`

	// Fields for Attribution
	AttributionReviewer   uuid.UUID `json:"attribution_reviewer" validate:"required"`
	AttributionContractor int       `json:"attribution_contractor"`
}

func CreateRecovery(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User) // TODO: Create a function to get the user from the context

	var input CreateRecoveryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Cannot parse JSON", "error": err.Error()})
	}

	if err := config.Validate.Struct(&input); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"errors":  errorMessages,
		})
	}

	// TODO: Maybe turn this into a function
	// Handle note: NULLIF(trim(@note), '')
	var finalNote *string
	if input.Note != nil {
		trimmedNote := strings.TrimSpace(*input.Note)
		if trimmedNote != "" {
			finalNote = &trimmedNote
		}
	}

	var createdRecoveryID int

	// Use a transaction to ensure atomicity
	err := db.Transaction(func(tx *gorm.DB) error {
		attribution := database.Attribution{
			Reviewer:   input.AttributionReviewer,
			Creator:    user.ID,
			Owner:      user.Organizations[0].ID, // TODO: handle multiple organizations
			Contractor: input.AttributionContractor,
		}
		if err := tx.Create(&attribution).Error; err != nil {
			return fmt.Errorf("failed to create attribution: %w", err)
		}

		// 2. Create Recovery
		recovery := database.Recovery{
			Note:         finalNote,
			Attribution:  attribution.ID,
			AccessPolicy: input.AccessPolicy,
			Type:         input.Type,
			DocumentDate: input.DocumentDate,
			DocumentFile: input.DocumentFile,
			DocumentName: input.DocumentName,
		}
		if err := tx.Create(&recovery).Error; err != nil {
			return fmt.Errorf("failed to create recovery: %w", err)
		}

		createdRecoveryID = recovery.ID
		return nil // Commit transaction
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create recovery record", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": createdRecoveryID})
}

func CreateRecoverySample(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	recoveryIdStr := c.Params("recovery_id")
	recoveryId, err := strconv.Atoi(recoveryIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid recovery_id format", "error": err.Error()})
	}

	var input database.RecoverySample
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Cannot parse JSON", "error": err.Error()})
	}

	// Set the recovery ID from the path parameter
	input.Recovery = recoveryId

	// Note: If database.RecoverySample struct has 'validate' tags, they will be checked here.
	// For example, if BuildingID is required, it should have `validate:"required"` tag in models.go.
	if err := config.Validate.Struct(&input); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"errors":  errorMessages,
		})
	}

	// Create the recovery sample record in the database
	if err := db.Create(&input).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create recovery sample record", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": input.ID})
}
