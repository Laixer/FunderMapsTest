package handlers

import (
	"fmt"
	"log"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
	"fundermaps/app/mail"
	"fundermaps/app/platform/geocoder"
	"fundermaps/app/platform/storage"
)

// TODO: Move into a separate package
func localizeBoolean(value *bool) string {
	localizer := i18n.NewLocalizer(config.Bundle, "nl", "en")

	if value == nil {
		return "N/A"
	}
	if *value {
		return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "yes"})
	}
	return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "no"})
}

func CreateIncident(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	localizer := i18n.NewLocalizer(config.Bundle, "nl", "en")

	geocoderService := geocoder.NewService(db)

	var input database.Incident
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
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

	if input.ContactPhoneNumber != nil && *input.ContactPhoneNumber == "" {
		input.ContactPhoneNumber = nil
	}

	building, err := geocoderService.GetBuildingByGeocoderID(input.Building)
	if err != nil {
		if err.Error() == "building not found" || err.Error() == "unknown geocoder identifier" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Building not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	legacyBuildingID, err := geocoderService.GetOldBuildingID(building.BuildingID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	input.Building = legacyBuildingID

	incident := database.Incident{
		ClientID:                         input.ClientID,
		FoundationType:                   input.FoundationType,
		ChainedBuilding:                  input.ChainedBuilding,
		Owner:                            input.Owner,
		FoundationRecovery:               input.FoundationRecovery,
		NeightborRecovery:                input.NeightborRecovery,
		FoundationDamageCause:            input.FoundationDamageCause,
		FileResourceKey:                  input.FileResourceKey,
		DocumentFile:                     input.DocumentFile,
		Note:                             input.Note,
		Contact:                          input.Contact,
		ContactName:                      input.ContactName,
		ContactPhoneNumber:               input.ContactPhoneNumber,
		EnvironmentDamageCharacteristics: input.EnvironmentDamageCharacteristics,
		FoundationDamageCharacteristics:  input.FoundationDamageCharacteristics,
		Building:                         input.Building,
		Metadata:                         input.Metadata,
	}

	result := db.Create(&incident)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// Update file resources if document files are provided
	if input.FileResourceKey != nil && len(incident.DocumentFile) > 0 {
		storageService := storage.NewStorageService(cfg.Storage())
		if err := storageService.UpdateFileStatus(db, *input.FileResourceKey, storage.StatusActive); err != nil {
			log.Printf("Failed to update file status: %v", err)
		}
	}

	// Use a nil check before dereferencing ContactPhoneNumber for the email body
	contactPhoneStr := "N/A"
	if incident.ContactPhoneNumber != nil {
		contactPhoneStr = *incident.ContactPhoneNumber
	}
	noteStr := "N/A"
	if incident.Note != nil {
		noteStr = *incident.Note
	}

	addressStr := incident.Building
	if incident.Metadata != nil {
		if addressName, ok := incident.Metadata["address_name"].(string); ok && addressName != "" {
			addressStr = addressName
		}
	}

	var foundationType string = "N/A"
	if incident.FoundationType != nil {
		foundationType, _ = localizer.Localize(&i18n.LocalizeConfig{MessageID: *incident.FoundationType, DefaultMessage: &i18n.Message{
			ID:    *incident.FoundationType,
			Other: *incident.FoundationType,
		}})
	}

	var foundationDamageCause string = "N/A"
	if incident.FoundationDamageCause != nil {
		foundationDamageCause, _ = localizer.Localize(&i18n.LocalizeConfig{MessageID: *incident.FoundationDamageCause, DefaultMessage: &i18n.Message{
			ID:    *incident.FoundationDamageCause,
			Other: *incident.FoundationDamageCause,
		}})
	}

	message := mail.Email{
		Subject:  "New Incident Report",
		From:     fmt.Sprintf("Fundermaps <no-reply@%s>", cfg.MailgunDomain),
		To:       cfg.EmailReceivers,
		Template: "incident-customer",
		TemplateVars: map[string]any{
			"id":                               incident.ID,
			"address":                          addressStr,
			"name":                             incident.ContactName,
			"phone":                            contactPhoneStr,
			"email":                            incident.Contact,
			"foundationType":                   foundationType,
			"chainedBuilding":                  localizeBoolean(&incident.ChainedBuilding),
			"owner":                            localizeBoolean(&incident.Owner),
			"neighborRecovery":                 localizeBoolean(&incident.NeightborRecovery),
			"foundationDamageCause":            foundationDamageCause,
			"foundationDamageCharacteristics":  incident.FoundationDamageCharacteristics,
			"environmentDamageCharacteristics": incident.EnvironmentDamageCharacteristics,
			"note":                             noteStr,
		},
	}

	mailer := mail.NewMailer(cfg.MailgunDomain, cfg.MailgunAPIKey, cfg.MailgunAPIBase)
	if err := mailer.SendTemplatedMail(&message); err != nil {
		log.Printf("Failed to send email notification: %v\n", err)
	}

	return c.JSON(incident)
}
