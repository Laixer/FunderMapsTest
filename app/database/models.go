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

// TODO: Rename to APIKey
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
	LastName          *string        `json:"family_name"` // TODO: Update database column name to family_name
	Email             string         `json:"email"`
	Avatar            *string        `json:"picture"` // TODO: Update database column name to picture
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
	ID                uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name              string      `json:"name"`
	Email             string      `json:"-"`                                     // TODO: Remove from db
	FenceMunicipality StringArray `json:"fence_municipality" gorm:"type:text[]"` // TODO: Move this out of the organization table
	FenceDistrict     StringArray `json:"fence_district" gorm:"type:text[]"`     // TODO: Move this out of the organization table
	FenceNeighborhood StringArray `json:"fence_neighborhood" gorm:"type:text[]"` // TODO: Move this out of the organization table
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
	Public        bool       `json:"-"`
	UserID        uuid.UUID  `json:"-" gorm:"type:uuid"`
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
	Code                string      `json:"code" gorm:"primaryKey"`
	Application         Application `json:"application" gorm:"foreignKey:ApplicationID;references:ApplicationID"`
	ApplicationID       string      `json:"-" gorm:"type:uuid;index"`
	User                User        `json:"user" gorm:"foreignKey:UserID;references:ID"`
	UserID              uuid.UUID   `json:"-" gorm:"type:uuid;index"`
	CreatedAt           time.Time   `json:"created_at" gorm:"default:now()"`
	ExpiredAt           time.Time   `json:"expired_at"`
	CodeChallenge       string      `json:"code_challenge"`
	CodeChallengeMethod string      `json:"code_challenge_method"`
}

func (ac *AuthCode) TableName() string {
	return "application.auth_code"
}

type ResetKey struct {
	Key        uuid.UUID `json:"key" gorm:"primaryKey"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid"`
	CreateDate time.Time `json:"create_date" gorm:"default:now()"`
}

func (rk *ResetKey) TableName() string {
	return "application.reset_key"
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

type FileResource struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Key              string     `json:"key" gorm:"unique;not null"`
	OriginalFilename string     `json:"original_filename" gorm:"not null"`
	Status           string     `json:"status" gorm:"default:'uploaded'"`
	SizeBytes        int64      `json:"size_bytes"`
	MimeType         string     `json:"mime_type"`
	Metadata         JSONObject `json:"metadata" gorm:"type:jsonb"`
	CreatedAt        time.Time  `json:"created_at" gorm:"default:now()"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"default:now()"`
}

func (fr *FileResource) TableName() string {
	return "application.file_resources"
}

type ProductTracker struct {
	Name       string `json:"product"`
	BuildingID string `json:"building_id"`
	Identifier string `json:"identifier"`
}

func (pt *ProductTracker) TableName() string {
	return "application.product_tracker"
}
