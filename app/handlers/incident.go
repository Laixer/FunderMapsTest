package handlers

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
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
	ClientID                         int         `json:"client_id" gorm:"-:all"`
	FoundationType                   *string     `json:"foundation_type"`
	ChainedBuilding                  bool        `json:"chained_building"`
	Owner                            bool        `json:"owner"`
	FoundationRecovery               bool        `json:"foundation_recovery"`
	NeightborRecovery                bool        `json:"neightbor_recovery"` // TODO: Fix typo
	FoundationDamageCause            *string     `json:"foundation_damage_cause"`
	DocumentFile                     StringArray `json:"document_file" gorm:"type:text[]"`
	Note                             *string     `json:"note"`
	Contact                          string      `json:"contact"`
	ContactName                      *string     `json:"contact_name"`
	ContactPhoneNumber               *string     `json:"contact_phone_number"`
	EnvironmentDamageCharacteristics StringArray `json:"environment_damage_characteristics" gorm:"type:text[]"`
	FoundationDamageCharacteristics  StringArray `json:"foundation_damage_characteristics" gorm:"type:text[]"`
	Building                         string      `json:"building"` // TODO: Rename to BuildingID
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

	// TODO: Add data validation
	// TODO: Check email is valid (regex)
	// TODO: Replace empty strings with db null

	if input.ClientID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Client ID is required"})
	}

	if input.Building == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Building is required"})
	}

	if input.Contact == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Contact is required"})
	}

	if input.ContactName == nil || *input.ContactName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Contact name is required"})
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

	message := mail.Email{
		Subject: "New Incident Report",
		Body:    fmt.Sprintf("A new incident report has been created:\n\nID: %s\nClient ID: %d\nBuilding: %s\nContact: %s\nContact Name: %s\nContact Phone Number: %s\nNote: %s", incident.ID, incident.ClientID, incident.Building, incident.Contact, *incident.ContactName, *incident.ContactPhoneNumber, *incident.Note),
		From:    fmt.Sprintf("Fundermaps <no-reply@%s>", cfg.MailgunDomain),
		To:      []string{fmt.Sprintf("Fundermaps <info@%s>", cfg.MailgunDomain)},
	}

	mailer := mail.NewMailer(cfg.MailgunDomain, cfg.MailgunAPIKey, cfg.MailgunAPIBase)
	mailer.SendMail(&message)

	return c.JSON(incident)
}

func UploadFiles(c *fiber.Ctx) error {
	cfg := c.Locals("config").(*config.Config)
	db := c.Locals("db").(*gorm.DB)

	storageService := storage.NewStorageService(cfg.Storage())

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Failed to parse form"})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "No files uploaded"})
	}

	keyName := storageService.GenerateKeyName()

	for _, file := range files {
		if storageService.IsFileExtensionAllowed(file.Filename) {
			if err := storageService.SaveFile(file, fmt.Sprintf("user-data/%s/%s", keyName, file.Filename), c); err != nil {
				return err
			}
		}
	}

	uploadedFiles := []string{}
	for _, file := range files {
		uploadedFiles = append(uploadedFiles, file.Filename)
	}

	// TODO: Move to service
	resourceFiles := []database.FileResource{}
	for _, file := range files {
		fileResource := database.FileResource{
			Key:              keyName,
			OriginalFilename: file.Filename,
			SizeBytes:        file.Size,
			MimeType:         file.Header.Get("Content-Type"),
		}
		resourceFiles = append(resourceFiles, fileResource)
	}
	if err := db.Create(&resourceFiles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to save file metadata"})
	}

	return c.JSON(fiber.Map{"files": uploadedFiles, "key": keyName})
}
