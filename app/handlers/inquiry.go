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

type CreateInquiryInput struct {
	Note                  *string   `json:"note"`
	AttributionReviewer   uuid.UUID `json:"attribution_reviewer" validate:"required"`
	AttributionContractor int       `json:"attribution_contractor" validate:"required"`
	Type                  string    `json:"type" validate:"required"` // TODO: Add validation for specific values from report.inquiry_type
	DocumentDate          time.Time `json:"document_date" validate:"required"`
	DocumentFile          string    `json:"document_file" validate:"required"` // TODO: This should ideally be a key to a file resource
	DocumentName          string    `json:"document_name" validate:"required"`
	Inspection            bool      `json:"inspection"`
	JointMeasurement      bool      `json:"joint_measurement"`
	FloorMeasurement      bool      `json:"floor_measurement"`
	StandardF3O           bool      `json:"standard_f3o"`
}

func CreateInquiry(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User) // TODO: Create a function to get the user from the context

	var input CreateInquiryInput
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

	var createdInquiryID int

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

		// 2. Create Inquiry
		inquiry := database.Inquiry{
			Note:             finalNote,
			Attribution:      attribution.ID,
			AccessPolicy:     "private",
			Type:             input.Type,
			DocumentDate:     input.DocumentDate,
			DocumentFile:     input.DocumentFile,
			DocumentName:     input.DocumentName,
			Inspection:       input.Inspection,
			JointMeasurement: input.JointMeasurement,
			FloorMeasurement: input.FloorMeasurement,
			StandardF3O:      input.StandardF3O,
		}
		if err := tx.Create(&inquiry).Error; err != nil {
			return fmt.Errorf("failed to create inquiry: %w", err)
		}

		createdInquiryID = inquiry.ID
		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create inquiry record", "error": err.Error()}) // Changed message
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": createdInquiryID})
}

func CreateInquirySample(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	inquiryIdStr := c.Params("inquiry_id")
	inquiryId, err := strconv.Atoi(inquiryIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid inquiry_id format", "error": err.Error()})
	}

	var input database.InquirySample
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Cannot parse JSON", "error": err.Error()})
	}

	// Set the inquiry ID from the path parameter
	input.Inquiry = inquiryId

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
