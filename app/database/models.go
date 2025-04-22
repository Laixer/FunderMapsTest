package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	BioInfectionReliability     string   `json:"bioInfection_reliability" gorm:"->"`
	BioInfectionRisk            *string  `json:"bioInfection_risk" gorm:"->"`
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

// User represents a user account in the system
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

// TableName specifies the database table name for the User model
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

// Application represents an OAuth application that can access the API
type Application struct {
	ApplicationID string     `json:"id" gorm:"primaryKey"`
	Name          string     `json:"name" validate:"required"`
	Data          JSONObject `json:"data" gorm:"type:jsonb"`
	Secret        string     `json:"-"` // TODO Rename to SecretHash
	RedirectURL   string     `json:"-" validate:"url"`
	Public        bool       `json:"-"`
	UserID        uuid.UUID  `json:"-" gorm:"type:uuid"`
}

// TableName specifies the database table name for the Application model
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

type AuthLog struct {
	LogID      int        `json:"log_id" gorm:"primaryKey"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid"`
	ActionType string     `json:"action_type"`
	IPAddress  string     `json:"ip_address"`
	UserAgent  string     `json:"user_agent"`
	Timestamp  time.Time  `json:"timestamp" gorm:"default:now()"`
	Metadata   JSONObject `json:"metadata" gorm:"type:jsonb"`
}

func (al *AuthLog) TableName() string {
	return "application.auth_logs"
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
	Key              string     `json:"key" gorm:"unique;not null" validate:"required"`
	OriginalFilename string     `json:"original_filename" gorm:"not null" validate:"required"`
	Status           string     `json:"status" gorm:"default:'uploaded'"`
	SizeBytes        int64      `json:"size_bytes"`
	MimeType         string     `json:"mime_type"`
	Metadata         JSONObject `json:"metadata" gorm:"type:jsonb"`
	CreatedAt        time.Time  `json:"created_at" gorm:"default:now()"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"default:now()"`
}

// TableName specifies the database table name for the FileResource model
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

// Incident represents a foundation-related incident report
type Incident struct {
	ID                               string      `json:"id" gorm:"primaryKey;<-:create"`
	ClientID                         int         `json:"client_id" gorm:"-:all" validate:"required"`
	FoundationType                   *string     `json:"foundation_type"`
	ChainedBuilding                  bool        `json:"chained_building"`
	Owner                            bool        `json:"owner"`
	FoundationRecovery               bool        `json:"foundation_recovery"`
	NeightborRecovery                bool        `json:"neightbor_recovery"` // TODO: Fix typo
	FoundationDamageCause            *string     `json:"foundation_damage_cause"`
	FileResourceKey                  *string     `json:"file_resource_key"`
	DocumentFile                     StringArray `json:"document_file" gorm:"type:text[]"`
	Note                             *string     `json:"note"`
	Contact                          string      `json:"contact" validate:"required,email"`
	ContactName                      *string     `json:"contact_name" validate:"required"`
	ContactPhoneNumber               *string     `json:"contact_phone_number"`
	EnvironmentDamageCharacteristics StringArray `json:"environment_damage_characteristics" gorm:"type:text[]"`
	FoundationDamageCharacteristics  StringArray `json:"foundation_damage_characteristics" gorm:"type:text[]"`
	Building                         string      `json:"building" validate:"required"` // TODO: Rename to BuildingID
	Metadata                         JSONObject  `json:"metadata" gorm:"type:jsonb"`
}

// BeforeCreate generates an ID for the incident before creation
func (i *Incident) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Raw("SELECT report.fir_generate_id(?)", i.ClientID).Scan(&i.ID)
	return nil
}

// TableName specifies the database table name for the Incident model
func (i *Incident) TableName() string {
	return "report.incident"
}

type InquirySample struct {
	ID                              int         `json:"id" gorm:"primaryKey"`
	Inquiry                         int         `json:"inquiry" gorm:"index"`
	CreateDate                      time.Time   `json:"create_date" gorm:"default:now()"`
	UpdateDate                      time.Time   `json:"update_date"`
	DeleteDate                      *time.Time  `json:"delete_date"`
	Note                            *string     `json:"note"`
	BuiltYear                       *time.Time  `json:"built_year"`
	Substructure                    *string     `json:"substructure"`
	OverallQuality                  *string     `json:"overall_quality"`
	WoodQuality                     *string     `json:"wood_quality"`
	ConstructionQuality             *string     `json:"construction_quality"`
	WoodCapacityHorizontalQuality   *string     `json:"wood_capacity_horizontal_quality"`
	PileWoodCapacityVerticalQuality *string     `json:"pile_wood_capacity_vertical_quality"`
	CarryingCapacityQuality         *string     `json:"carrying_capacity_quality"`
	MasonQuality                    *string     `json:"mason_quality"`
	WoodQualityNecessity            *bool       `json:"wood_quality_necessity"`
	ConstructionLevel               *float64    `json:"construction_level"`
	WoodLevel                       *float64    `json:"wood_level"`
	PileDiameterTop                 *float64    `json:"pile_diameter_top"`
	PileDiameterBottom              *float64    `json:"pile_diameter_bottom"`
	PileHeadLevel                   *float64    `json:"pile_head_level"`
	PileTipLevel                    *float64    `json:"pile_tip_level"`
	FoundationDepth                 *float64    `json:"foundation_depth"`
	MasonLevel                      *float64    `json:"mason_level"`
	ConcreteChargerLength           *float64    `json:"concrete_charger_length"`
	PileDistanceLength              *float64    `json:"pile_distance_length"`
	WoodPenetrationDepth            *float64    `json:"wood_penetration_depth"`
	CPT                             *string     `json:"cpt"`
	MonitoringWell                  *string     `json:"monitoring_well"`
	GroundwaterLevelTemp            *float64    `json:"groundwater_level_temp"`
	GroundLevel                     *float64    `json:"ground_level" gorm:"column:groundlevel"` // TODO: Fix coulumn name in database
	GroundwaterLevelNet             *float64    `json:"groundwater_level_net"`
	FoundationType                  *string     `json:"foundation_type"`
	EnforcementTerm                 *string     `json:"enforcement_term"`
	RecoveryAdvised                 *bool       `json:"recovery_advised"`
	DamageCause                     *string     `json:"damage_cause"`
	DamageCharacteristics           StringArray `json:"damage_characteristics" gorm:"type:text[]"`
	ConstructionPile                *string     `json:"construction_pile"`
	WoodType                        *string     `json:"wood_type"`
	WoodEncroachment                *string     `json:"wood_encroachment" gorm:"column:wood_encroachement"` // TODO: Fix typo in database
	CrackIndoorRestored             *bool       `json:"crack_indoor_restored"`
	CrackIndoorType                 *string     `json:"crack_indoor_type"`
	CrackIndoorSize                 *string     `json:"crack_indoor_size"`
	CrackFacadeFrontRestored        *bool       `json:"crack_facade_front_restored"`
	CrackFacadeFrontType            *string     `json:"crack_facade_front_type"`
	CrackFacadeFrontSize            *string     `json:"crack_facade_front_size"`
	CrackFacadeBackRestored         *bool       `json:"crack_facade_back_restored"`
	CrackFacadeBackType             *string     `json:"crack_facade_back_type"`
	CrackFacadeBackSize             *string     `json:"crack_facade_back_size"`
	CrackFacadeLeftRestored         *bool       `json:"crack_facade_left_restored"`
	CrackFacadeLeftType             *string     `json:"crack_facade_left_type"`
	CrackFacadeLeftSize             *string     `json:"crack_facade_left_size"`
	CrackFacadeRightRestored        *bool       `json:"crack_facade_right_restored"`
	CrackFacadeRightType            *string     `json:"crack_facade_right_type"`
	CrackFacadeRightSize            *string     `json:"crack_facade_right_size"`
	DeformedFacade                  *bool       `json:"deformed_facade"`
	ThresholdUpdownSkewed           *bool       `json:"threshold_updown_skewed"`
	ThresholdFrontLevel             *float64    `json:"threshold_front_level"`
	ThresholdBackLevel              *float64    `json:"threshold_back_level"`
	SkewedParallel                  *float64    `json:"skewed_parallel"`
	SkewedPerpendicular             *float64    `json:"skewed_perpendicular"`
	SkewedParallelFacade            *string     `json:"skewed_parallel_facade"`
	SettlementSpeed                 *float64    `json:"settlement_speed"`
	SkewedWindowFrame               *bool       `json:"skewed_window_frame"`
	SkewedPerpendicularFacade       *string     `json:"skewed_perpendicular_facade"`
	Building                        string      `json:"building" gorm:"index"`
	FacadeScanRisk                  *string     `json:"facade_scan_risk"`
	Metadata                        JSONObject  `json:"metadata" gorm:"type:jsonb"`
}

// TableName specifies the database table name for the Incident model
func (is *InquirySample) TableName() string {
	return "report.inquiry_sample"
}

type RecoverySample struct {
	ID           int         `json:"id" gorm:"primaryKey"`
	Recovery     int         `json:"recovery" gorm:"index"`
	CreateDate   time.Time   `json:"create_date" gorm:"default:now()"`
	UpdateDate   time.Time   `json:"update_date"`
	DeleteDate   *time.Time  `json:"delete_date"`
	Note         *string     `json:"note"`
	Status       *string     `json:"status"`
	Type         *string     `json:"type"`
	PileType     *string     `json:"pile_type"`
	Facade       StringArray `json:"facade" gorm:"type:text[]"`
	Permit       *string     `json:"permit"`
	PermitDate   *time.Time  `json:"permit_date"`
	RecoveryDate *time.Time  `json:"recovery_date"`
	Contractor   *int        `json:"contractor"`
	BuildingID   string      `json:"building_id" gorm:"index"`
	Metadata     JSONObject  `json:"metadata" gorm:"type:jsonb"`
}

// TableName specifies the database table name for the RecoverySample model
func (rs *RecoverySample) TableName() string {
	return "report.recovery_sample"
}
