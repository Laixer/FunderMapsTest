package incident

import (
	"fmt"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
	"fundermaps/app/mail"
	"fundermaps/app/platform/geocoder"
	"fundermaps/app/platform/storage"
)

type Service struct {
	db          *gorm.DB
	cfg         *config.Config
	geocoderSvc *geocoder.GeocoderService
	storageSvc  storage.StorageService
	mailer      *mail.Mailgun
	bundle      *i18n.Bundle
}

// NewService creates a new incident service.
func NewService(db *gorm.DB, cfg *config.Config, bundle *i18n.Bundle) *Service {
	return &Service{
		db:          db,
		cfg:         cfg,
		geocoderSvc: geocoder.NewService(db),
		storageSvc:  storage.NewStorageService(cfg.Storage()),
		mailer:      mail.NewMailer(cfg.MailgunDomain, cfg.MailgunAPIKey, cfg.MailgunAPIBase),
		bundle:      bundle,
	}
}

func (s *Service) localizeBoolean(value *bool) string {
	localizer := i18n.NewLocalizer(s.bundle, "nl", "en")

	if value == nil {
		return "N/A"
	}
	if *value {
		return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "yes"})
	}
	return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "no"})
}

// Create handles the business logic of creating an incident.
// The inputData parameter contains the initial incident data.
func (s *Service) Create(inputData database.Incident) (*database.Incident, error) {
	localizer := i18n.NewLocalizer(s.bundle, "nl", "en")

	building, err := s.geocoderSvc.GetBuildingByGeocoderID(inputData.Building)
	if err != nil {
		if err.Error() == "building not found" || err.Error() == "unknown geocoder identifier" {
			return nil, fmt.Errorf("building_not_found: %w", err)
		}
		return nil, fmt.Errorf("geocoder_error: %w", err)
	}

	legacyBuildingID, err := s.geocoderSvc.GetOldBuildingID(building.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("geocoder_old_id_error: %w", err)
	}

	// Prepare the incident struct for database insertion
	incidentToCreate := database.Incident{
		ClientID:                         inputData.ClientID,
		FoundationType:                   inputData.FoundationType,
		ChainedBuilding:                  inputData.ChainedBuilding,
		Owner:                            inputData.Owner,
		FoundationRecovery:               inputData.FoundationRecovery,
		NeightborRecovery:                inputData.NeightborRecovery,
		FoundationDamageCause:            inputData.FoundationDamageCause,
		FileResourceKey:                  inputData.FileResourceKey,
		DocumentFile:                     inputData.DocumentFile,
		Note:                             inputData.Note,
		Contact:                          inputData.Contact,
		ContactName:                      inputData.ContactName,
		ContactPhoneNumber:               inputData.ContactPhoneNumber,
		EnvironmentDamageCharacteristics: inputData.EnvironmentDamageCharacteristics,
		FoundationDamageCharacteristics:  inputData.FoundationDamageCharacteristics,
		Building:                         legacyBuildingID, // Use the resolved legacy building ID
		Metadata:                         inputData.Metadata,
	}

	result := s.db.Create(&incidentToCreate)
	if result.Error != nil {
		return nil, fmt.Errorf("database_error: %w", result.Error)
	}

	// Update file resources if document files are provided
	if incidentToCreate.FileResourceKey != nil && len(incidentToCreate.DocumentFile) > 0 {
		if err := s.storageSvc.UpdateFileStatus(s.db, *incidentToCreate.FileResourceKey, storage.StatusActive); err != nil {
			log.Printf("Failed to update file status: %v", err) // Log and continue
		}
	}

	// Prepare and send email notification
	contactPhoneStr := "N/A"
	if incidentToCreate.ContactPhoneNumber != nil {
		contactPhoneStr = *incidentToCreate.ContactPhoneNumber
	}
	noteStr := "N/A"
	if incidentToCreate.Note != nil {
		noteStr = *incidentToCreate.Note
	}

	addressStr := incidentToCreate.Building
	if incidentToCreate.Metadata != nil {
		if addressName, ok := incidentToCreate.Metadata["address_name"].(string); ok && addressName != "" {
			addressStr = addressName
		}
	}

	var foundationType string = "N/A"
	if incidentToCreate.FoundationType != nil {
		foundationType, _ = localizer.Localize(&i18n.LocalizeConfig{MessageID: *incidentToCreate.FoundationType, DefaultMessage: &i18n.Message{
			ID:    *incidentToCreate.FoundationType,
			Other: *incidentToCreate.FoundationType,
		}})
	}

	var foundationDamageCause string = "N/A"
	if incidentToCreate.FoundationDamageCause != nil {
		foundationDamageCause, _ = localizer.Localize(&i18n.LocalizeConfig{MessageID: *incidentToCreate.FoundationDamageCause, DefaultMessage: &i18n.Message{
			ID:    *incidentToCreate.FoundationDamageCause,
			Other: *incidentToCreate.FoundationDamageCause,
		}})
	}

	message := mail.Email{
		Subject:  "New Incident Report",
		From:     fmt.Sprintf("Fundermaps <no-reply@%s>", s.cfg.MailgunDomain),
		To:       s.cfg.EmailReceivers,
		Template: "incident-customer",
		TemplateVars: map[string]any{
			"id":                               incidentToCreate.ID,
			"address":                          addressStr,
			"name":                             incidentToCreate.ContactName,
			"phone":                            contactPhoneStr,
			"email":                            incidentToCreate.Contact,
			"foundationType":                   foundationType,
			"chainedBuilding":                  s.localizeBoolean(&incidentToCreate.ChainedBuilding),
			"owner":                            s.localizeBoolean(&incidentToCreate.Owner),
			"neighborRecovery":                 s.localizeBoolean(&incidentToCreate.NeightborRecovery),
			"foundationDamageCause":            foundationDamageCause,
			"foundationDamageCharacteristics":  incidentToCreate.FoundationDamageCharacteristics,
			"environmentDamageCharacteristics": incidentToCreate.EnvironmentDamageCharacteristics,
			"note":                             noteStr,
		},
	}

	if err := s.mailer.SendTemplatedMail(&message); err != nil {
		log.Printf("Failed to send email notification: %v\n", err) // Log and continue
	}

	return &incidentToCreate, nil
}
