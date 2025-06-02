package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
)

// CreateRecoveryInput defines the expected request body for creating a recovery record.
type CreateRecoveryInput struct {
	// Fields for Recovery
	Note         *string   `json:"note"`
	AccessPolicy string    `json:"access_policy" validate:"required,oneof=public private"`
	Type         string    `json:"type" validate:"required"`
	DocumentDate time.Time `json:"document_date" validate:"required"`
	DocumentFile string    `json:"document_file" validate:"required"`
	DocumentName string    `json:"document_name" validate:"required"`

	// Fields for Attribution
	AttributionReviewer   uuid.UUID `json:"attribution_reviewer" validate:"required"`
	AttributionContractor int       `json:"attribution_contractor"`
}

func CreateRecovery(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

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
		// 1. Create Attribution
		// Assumes database.Attribution struct exists and matches table structure:
		// ID (int, pk), Reviewer (uuid), Creator (uuid), Owner (uuid),
		// Contractor (uuid), Contractor2 (int, nullable)
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
		// Log the internal error for debugging
		// log.Printf("Error creating recovery record: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create recovery record", "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": createdRecoveryID})
}
