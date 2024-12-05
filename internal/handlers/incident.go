package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Incident struct {
	ID                               string   `json:"id" gorm:"primaryKey;<-:create"`
	ClientID                         int      `json:"client_id" gorm:"-:all"`
	FoundationType                   *string  `json:"foundation_type"`
	ChainedBuilding                  bool     `json:"chained_building"`
	Owner                            bool     `json:"owner"`
	FoundationRecovery               bool     `json:"foundation_recovery"`
	NeightborRecovery                bool     `json:"neightbor_recovery"` // TODO: Fix typo
	FoundationDamageCause            *string  `json:"foundation_damage_cause"`
	DocumentFile                     []string `json:"document_file" gorm:"type:text[]"`
	Note                             *string  `json:"note"`
	Contact                          string   `json:"contact"`
	ContactName                      *string  `json:"contact_name"`
	ContactPhoneNumber               *string  `json:"contact_phone_number"`
	EnvironmentDamageCharacteristics []string `json:"environment_damage_characteristics" gorm:"type:text[]"`
	FoundationDamageCharacteristics  []string `json:"foundation_damage_characteristics" gorm:"type:text[]"`
	Building                         string   `json:"building"` // TODO: Rename to BuildingID
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
	db := c.Locals("db").(*gorm.DB)

	var input Incident
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if input.Building == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Building ID required")
	}

	// TODO: Validate client_id, not all clients can create incidents

	if input.ClientID == 0 {
		input.ClientID = 10
	}

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	return c.JSON(incident)
}
