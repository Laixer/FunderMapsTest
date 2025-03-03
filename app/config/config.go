package config

import (
	"fmt"

	"github.com/go-playground/validator"
	"github.com/gofiber/storage/s3/v2"
	"github.com/spf13/viper"
)

// TODO: Move into a separate package
var Validate *validator.Validate

type Config struct {
	ServerPort     int      `mapstructure:"SERVER_PORT" validate:"required,min=1,max=65535"`
	DatabaseURL    string   `mapstructure:"DATABASE_URL" validate:"required,url"`
	ApplicationID  string   `mapstructure:"APP_ID" validate:"required"`
	AuthExpiration int      `mapstructure:"AUTH_EXPIRATION" validate:"required,min=1"`
	AuthDomain     string   `mapstructure:"AUTH_DOMAIN" validate:"required"`
	AuthSecure     bool     `mapstructure:"AUTH_SECURE"`
	MailgunAPIKey  string   `mapstructure:"MAILGUN_API_KEY" validate:"required_with=MailgunDomain"`
	MailgunDomain  string   `mapstructure:"MAILGUN_DOMAIN" validate:"required_with=MailgunAPIKey"`
	MailgunAPIBase string   `mapstructure:"MAILGUN_API_BASE"`
	S3Endpoint     string   `mapstructure:"S3_ENDPOINT" validate:"required_with=S3Bucket"`
	S3Region       string   `mapstructure:"S3_REGION" validate:"required_with=S3Bucket"`
	S3Bucket       string   `mapstructure:"S3_BUCKET"`
	S3AccessKey    string   `mapstructure:"S3_ACCESS_KEY" validate:"required_with=S3Bucket"`
	S3SecretKey    string   `mapstructure:"S3_SECRET_KEY" validate:"required_with=S3Bucket,min=8"`
	ProxyEnabled   bool     `mapstructure:"PROXY_ENABLED"`
	ProxyNetworks  []string `mapstructure:"PROXY_NETWORKS"` // validate:"dive,cidr,required_if=ProxyEnabled true"`
	ProxyHeader    string   `mapstructure:"PROXY_HEADER"`   // validate:"required_if=ProxyEnabled true"`
}

func Load() (*Config, error) {
	// Set environment variable prefix for FunderMaps
	viper.SetEnvPrefix("FM")

	// Default values
	viper.SetDefault("SERVER_PORT", 3_000)
	viper.SetDefault("DATABASE_URL", "postgres://postgres@localhost:5432/fundermaps")
	viper.SetDefault("AUTH_EXPIRATION", 24)
	viper.SetDefault("AUTH_DOMAIN", "localhost")
	viper.SetDefault("AUTH_SECURE", false)

	// Enable automatic environment variable binding with the FM_ prefix
	viper.AutomaticEnv()

	// Explicitly bind both prefixed and non-prefixed environment variables
	// for backward compatibility
	viper.BindEnv("APP_ID", "FM_APP_ID", "APP_ID")

	viper.BindEnv("SERVER_PORT", "FM_SERVER_PORT", "SERVER_PORT")
	viper.BindEnv("DATABASE_URL", "FM_DATABASE_URL", "DATABASE_URL")

	// Bind authentication environment variables
	viper.BindEnv("AUTH_EXPIRATION", "FM_AUTH_EXPIRATION", "AUTH_EXPIRATION")
	viper.BindEnv("AUTH_DOMAIN", "FM_AUTH_DOMAIN", "AUTH_DOMAIN")
	viper.BindEnv("AUTH_SECURE", "FM_AUTH_SECURE", "AUTH_SECURE")

	// Bind Mailgun environment variables
	viper.BindEnv("MAILGUN_API_KEY", "FM_MAILGUN_API_KEY", "MAILGUN_API_KEY")
	viper.BindEnv("MAILGUN_DOMAIN", "FM_MAILGUN_DOMAIN", "MAILGUN_DOMAIN")
	viper.BindEnv("MAILGUN_API_BASE", "FM_MAILGUN_API_BASE", "MAILGUN_API_BASE")

	// Bind S3 storage environment variables
	viper.BindEnv("S3_ENDPOINT", "FM_S3_ENDPOINT", "S3_ENDPOINT")
	viper.BindEnv("S3_REGION", "FM_S3_REGION", "S3_REGION")
	viper.BindEnv("S3_BUCKET", "FM_S3_BUCKET", "S3_BUCKET")
	viper.BindEnv("S3_ACCESS_KEY", "FM_S3_ACCESS_KEY", "S3_ACCESS_KEY")
	viper.BindEnv("S3_SECRET_KEY", "FM_S3_SECRET_KEY", "S3_SECRET_KEY")

	// Bind proxy config environment variables
	viper.BindEnv("PROXY_ENABLED", "FM_PROXY_ENABLED", "PROXY_ENABLED")
	viper.BindEnv("PROXY_NETWORKS", "FM_PROXY_NETWORKS", "PROXY_NETWORKS")
	viper.BindEnv("PROXY_HEADER", "FM_PROXY_HEADER", "PROXY_HEADER")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/fundermaps/")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Initialize validator
	Validate = validator.New()

	// Validate the config
	if err := Validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func (cfg *Config) Storage() *s3.Storage {
	return s3.New(s3.Config{
		Bucket:   cfg.S3Bucket,
		Endpoint: cfg.S3Endpoint,
		Region:   cfg.S3Region,
		Reset:    false,
		Credentials: s3.Credentials{
			AccessKey:       cfg.S3AccessKey,
			SecretAccessKey: cfg.S3SecretKey,
		},
	})
}
