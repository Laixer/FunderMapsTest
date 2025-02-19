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

type Address struct {
	ID             string `json:"-" gorm:"primaryKey"`
	ExternalID     string `json:"id"`
	BuildingID     string `json:"-"`
	BuildingNumber string `json:"building_number"`
	PostalCode     string `json:"postal_code"`
	Street         string `json:"street"`
	City           string `json:"city"`
}

func (a *Address) TableName() string {
	return "geocoder.address"
}

// TODO: Add custom types for the database data types
type Analysis struct {
	BuildingID                  string   `json:"building_id" gorm:"->"`
	NeighborhoodID              string   `json:"neighborhood_id" gorm:"->"`
	ConstructionYear            *int     `json:"construction_year" gorm:"->"`
	ConstructionYearReliability string   `json:"construction_year_reliability" gorm:"->"`
	RecoveryType                *string  `json:"recovery_type" gorm:"->"`
	RestorationCosts            *float64 `json:"restoration_costs" gorm:"->"`
	Height                      *float64 `json:"height" gorm:"->"`
	Velocity                    *float64 `json:"velocity" gorm:"->"`
	GroundWaterLevel            *float64 `json:"ground_water_level" gorm:"->"`
	GroundLevel                 *float64 `json:"ground_level" gorm:"->"`
	Soil                        *string  `json:"soil" gorm:"->"`
	SurfaceArea                 *float64 `json:"surface_area" gorm:"->"`
	DamageCause                 *string  `json:"damage_cause" gorm:"->"`
	EnforcementTerm             *string  `json:"enforcement_term" gorm:"->"`
	OverallQuality              *string  `json:"overall_quality" gorm:"->"`
	InquiryType                 *string  `json:"inquiry_type" gorm:"->"`
	FoundationType              *string  `json:"foundation_type" gorm:"->"`
	FoundationTypeReliability   string   `json:"foundation_type_reliability" gorm:"->"`
	Drystand                    *float64 `json:"drystand" gorm:"->"`
	DrystandReliability         string   `json:"drystand_reliability" gorm:"->"`
	DrystandRisk                *string  `json:"drystand_risk" gorm:"->"`
	DewateringDepth             *float64 `json:"dewatering_depth" gorm:"->"`
	DewateringDepthReliability  string   `json:"dewatering_depth_reliability" gorm:"->"`
	DewateringDepthRisk         *string  `json:"dewatering_depth_risk" gorm:"->"`
	BioInfectionReliability     string   `json:"bio_infection_reliability" gorm:"->"`
	BioInfectionRisk            *string  `json:"bio_infection_risk" gorm:"->"`
	UnclassifiedRisk            *string  `json:"unclassified_risk" gorm:"->"`
}

func (a *Analysis) TableName() string {
	return "data.model_risk_static"
}

type AuthKey struct {
	Key    string    `json:"key" gorm:"default:concat('fmsk.', application.random_string(32));primaryKey"` // TODO: Wrap this into a database function
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid"`
}

func (ak *AuthKey) TableName() string {
	return "application.auth_key"
}

type User struct {
	// NormalizedEmail   string    `json:"normalized_email"` // TODO: Drop from database
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GivenName         *string        `json:"given_name"`
	LastName          *string        `json:"family_name"` // TODO: Update database column name
	Email             string         `json:"email"`
	Avatar            *string        `json:"picture"` // TODO: Update database column name
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
	ApplicationID string     `json:"id" gorm:"primaryKey"`
	Name          string     `json:"name"`
	Data          JSONObject `json:"data" gorm:"type:jsonb"`
	Secret        string     `json:"-"` // TODO Rename to SecretHash
	RedirectURL   string     `json:"-"`
}

func (a *Application) TableName() string {
	return "application.application"
}

type ApplicationUser struct {
	UserID        string     `json:"-" gorm:"primaryKey"`
	ApplicationID string     `json:"-" gorm:"primaryKey"`
	Metadata      JSONObject `json:"metadata" gorm:"type:jsonb"`
	UpdateDate    time.Time  `json:"update_date"`
}

func (a *ApplicationUser) TableName() string {
	return "application.application_user"
}

type AuthCode struct {
	Code          string      `json:"code" gorm:"primaryKey"`
	Application   Application `json:"application" gorm:"foreignKey:ApplicationID;references:ApplicationID"`
	ApplicationID string      `json:"-" gorm:"type:uuid"`
	User          User        `json:"user" gorm:"foreignKey:UserID;references:ID"`
	UserID        uuid.UUID   `json:"-" gorm:"type:uuid"`
	CreatedAt     time.Time   `json:"created_at" gorm:"default:now()"`
	ExpiredAt     time.Time   `json:"expired_at"`
}

func (ac *AuthCode) TableName() string {
	return "application.auth_code"
}

type AuthAccessToken struct {
	AccessToken   string      `json:"access_token" gorm:"primaryKey"`
	IPAddress     string      `json:"ip_address"`
	Application   Application `json:"application" gorm:"foreignKey:ApplicationID;references:ApplicationID"`
	ApplicationID string      `json:"-" gorm:"type:uuid"`
	User          User        `json:"user" gorm:"foreignKey:UserID;references:ID"`
	UserID        uuid.UUID   `json:"-" gorm:"type:uuid"`
	CreatedAt     time.Time   `json:"created_at" gorm:"default:now()"`
	UpdatedAt     time.Time   `json:"updated_at"`
	ExpiredAt     time.Time   `json:"expired_at"`
}

func (aat *AuthAccessToken) TableName() string {
	return "application.auth_access_token"
}

type AuthRefreshToken struct {
	Token         string      `json:"token" gorm:"primaryKey"`
	Application   Application `json:"application" gorm:"foreignKey:ApplicationID;references:ApplicationID"`
	ApplicationID string      `json:"-" gorm:"type:uuid"`
	User          User        `json:"user" gorm:"foreignKey:UserID;references:ID"`
	UserID        uuid.UUID   `json:"-" gorm:"type:uuid"`
	CreatedAt     time.Time   `json:"created_at" gorm:"default:now()"`
	ExpiredAt     time.Time   `json:"expired_at"`
}

func (art *AuthRefreshToken) TableName() string {
	return "application.auth_refresh_token"
}

type Mapset struct {
	ID       string      `json:"id" gorm:"primaryKey"`
	Name     string      `json:"name"`
	Slug     string      `json:"slug"`
	Style    string      `json:"style"`
	Layers   StringArray `json:"layers" gorm:"type:text[]"`
	Options  JSONObject  `json:"options" gorm:"type:jsonb"`
	Public   bool        `json:"public"`
	Consent  *string     `json:"consent"`
	Note     string      `json:"note"`
	Icon     *string     `json:"icon"`
	Order    int         `json:"order"`
	Layerset JSONArray   `json:"layerset" gorm:"type:jsonb"`
}

func (u *Mapset) TableName() string {
	return "maplayer.mapset_collection"
}
