package handlers

import (
	"database/sql/driver"
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/mail"
	"fundermaps/app/platform/geocoder"
	"fundermaps/app/platform/storage"
)

// TODO: Move into a models package
type StringArray []string

// Implement the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		str := string(v)
		str = strings.Trim(str, "{}")
		if str == "" {
			*a = []string{}
		} else {
			*a = strings.Split(str, ",")
		}
		return nil
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type StringArray", value)
	}
}

// Convert the StringArray to a valid string for the database
func (a StringArray) Value() (driver.Value, error) {
	var cleanedArray []string
	for _, str := range a {
		if strings.TrimSpace(str) != "" {
			cleanedArray = append(cleanedArray, str)
		}
	}
	if len(cleanedArray) == 0 {
		return nil, nil
	}
	return fmt.Sprintf("{%s}", strings.Join(cleanedArray, ",")), nil
}

type Incident struct {
	ID                               string      `json:"id" gorm:"primaryKey;<-:create"`
	ClientID                         int         `json:"client_id" gorm:"-:all" validate:"required"`
	FoundationType                   *string     `json:"foundation_type"`
	ChainedBuilding                  bool        `json:"chained_building"`
	Owner                            bool        `json:"owner"`
	FoundationRecovery               bool        `json:"foundation_recovery"`
	NeightborRecovery                bool        `json:"neightbor_recovery"` // TODO: Fix typo
	FoundationDamageCause            *string     `json:"foundation_damage_cause"`
	DocumentFile                     StringArray `json:"document_file" gorm:"type:text[]"`
	Note                             *string     `json:"note"`
	Contact                          string      `json:"contact" validate:"required,email"`
	ContactName                      *string     `json:"contact_name" validate:"required"`
	ContactPhoneNumber               *string     `json:"contact_phone_number"`
	EnvironmentDamageCharacteristics StringArray `json:"environment_damage_characteristics" gorm:"type:text[]"`
	FoundationDamageCharacteristics  StringArray `json:"foundation_damage_characteristics" gorm:"type:text[]"`
	Building                         string      `json:"building" validate:"required"` // TODO: Rename to BuildingID
	// Meta							 *string  `json:"meta"`
}

func (i *Incident) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Raw("SELECT report.fir_generate_id(?)", i.ClientID).Scan(&i.ID)
	return nil
}

func (i *Incident) TableName() string {
	return "report.incident"
}

func CreateIncident(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	geocoderService := geocoder.NewService(db)

	var input Incident
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

	// If ContactPhoneNumber is provided but is an empty string, treat it as NULL
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

	incident := Incident{
		ClientID:                         input.ClientID,
		FoundationType:                   input.FoundationType,
		ChainedBuilding:                  input.ChainedBuilding,
		Owner:                            input.Owner,
		FoundationRecovery:               input.FoundationRecovery,
		NeightborRecovery:                input.NeightborRecovery,
		FoundationDamageCause:            input.FoundationDamageCause,
		DocumentFile:                     input.DocumentFile,
		Note:                             input.Note,
		Contact:                          input.Contact,
		ContactName:                      input.ContactName,
		ContactPhoneNumber:               input.ContactPhoneNumber,
		EnvironmentDamageCharacteristics: input.EnvironmentDamageCharacteristics,
		FoundationDamageCharacteristics:  input.FoundationDamageCharacteristics,
		Building:                         input.Building,
	}

	result := db.Create(&incident)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// Update file resources if document files are provided
	// if len(incident.DocumentFile) > 0 {
	// 	storageService := storage.NewStorageService(cfg.Storage())
	// 	if err := storageService.UpdateFileStatus(db, strings.Join(incident.DocumentFile, ","), storage.StatusActive); err != nil {
	// 		log.Printf("Failed to update file status: %v", err)
	// 	}
	// }

	// Use a nil check before dereferencing ContactPhoneNumber for the email body
	contactPhoneStr := "N/A"
	if incident.ContactPhoneNumber != nil {
		contactPhoneStr = *incident.ContactPhoneNumber
	}
	noteStr := "N/A"
	if incident.Note != nil {
		noteStr = *incident.Note
	}

	message := mail.Email{
		Subject:  "New Incident Report",
		From:     fmt.Sprintf("Fundermaps <no-reply@%s>", cfg.MailgunDomain),
		To:       cfg.EmailReceivers,
		Template: "incident-customer",
		TemplateVars: map[string]any{
			"id":                               incident.ID,
			"address":                          incident.Building,
			"name":                             incident.ContactName,
			"phone":                            contactPhoneStr,
			"email":                            incident.Contact,
			"foundationType":                   incident.FoundationType,
			"chainedBuilding":                  incident.ChainedBuilding,
			"owner":                            incident.Owner,
			"neighborRecovery":                 incident.NeightborRecovery,
			"foundationDamageCause":            incident.FoundationDamageCause,
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

func UploadFiles(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	storageService := storage.NewStorageService(cfg.Storage())

	formField := c.Query("field")

	result, err := storageService.UploadFile(c, db, formField)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(result)
}
