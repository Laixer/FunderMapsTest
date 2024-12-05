package database

import (
	"time"

	"github.com/google/uuid"
)

type Contractor struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func (u *Contractor) TableName() string {
	return "application.contractor"
}

// TODO: Incomplete model
type Analysis struct {
	NeighborhoodID   string  `json:"neighborhood_id"`
	RestorationCosts float64 `json:"restoration_costs"`
	Height           float64 `json:"height"`
	Velocity         float64 `json:"velocity"`
}

func (a *Analysis) TableName() string {
	return "data.model_risk_static"
}

type AuthKey struct {
	Key    string    `json:"key" gorm:"default:concat('fmsk.', application.random_string(32));primaryKey"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid"`
}

func (ak *AuthKey) TableName() string {
	return "application.auth_key"
}

type User struct {
	// NormalizedEmail   string    `json:"normalized_email"` // TODO: Do we need this?
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GivenName         *string        `json:"given_name"`
	LastName          *string        `json:"last_name"`
	Email             string         `json:"email"`
	Avatar            *string        `json:"avatar"`
	JobTitle          *string        `json:"job_title"`
	PasswordHash      string         `json:"-"`
	PhoneNumber       *string        `json:"phone_number"`
	AccessFailedCount int            `json:"-" gorm:"default:0"`
	Role              string         `json:"role" gorm:"default:'user'"`
	LastLogin         time.Time      `json:"-" gorm:"default:now()"`
	LoginCount        int            `json:"-" gorm:"default:0"`
	Organizations     []Organization `json:"organizations" gorm:"many2many:application.organization_user;foreignKey:ID;joinForeignKey:user_id;References:ID;joinReferences:organization_id;jointable_columns:role"`
}

func (u *User) TableName() string {
	return "application.user"
}

type Organization struct {
	ID    uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name  string    `json:"name"`
	Email string    `json:"-"` // TODO: Remove from db
	// FenceMunicipality string `json:"fence_municipality"`
	// FenceDistrict     string `json:"fence_district"`
	// FenceNeighborhood string `json:"fence_neighborhood"`
	// Users []User `gorm:"many2many:application.organization_user;foreignKey:ID;joinForeignKey:OrganizationID;References:ID;joinReferences:UserID"`
}

func (o *Organization) TableName() string {
	return "application.organization"
}

// type OrganizationUser struct {
// 	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
// 	OrganizationID uuid.UUID `gorm:"type:uuid;primaryKey"`
// 	Role           string    `gorm:"not null"`
// }

// func (ou *OrganizationUser) TableName() string {
// 	return "application.organization_user"
// }

type Application struct {
	ApplicationID string `json:"id" gorm:"primaryKey"`
	Name          string `json:"name"`
	Data          string `json:"data" gorm:"type:jsonb"`
	Secret        string `json:"-"`
}

func (a *Application) TableName() string {
	return "application.application"
}
