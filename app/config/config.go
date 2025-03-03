package config

import (
	"fmt"

	"github.com/go-playground/validator"
	"github.com/gofiber/storage/s3/v2"
	"github.com/spf13/viper"

	"fundermaps/pkg/utils"
)

// TODO: Move into a separate package
var Validate *validator.Validate

type Config struct {
	ServerPort     int    `mapstructure:"SERVER_PORT"`
	DatabaseURL    string `mapstructure:"DATABASE_URL"`
	ApplicationID  string `mapstructure:"APP_ID"`
	JWTSecret      string `mapstructure:"JWT_SECRET"`
	AuthExpiration int    `mapstructure:"AUTH_EXPIRATION"`
	AuthDomain     string `mapstructure:"AUTH_DOMAIN"`
	AuthSecure     bool   `mapstructure:"AUTH_SECURE"`
	MailgunAPIKey  string `mapstructure:"MAILGUN_API_KEY"`
	MailgunDomain  string `mapstructure:"MAILGUN_DOMAIN"`
	MailgunAPIBase string `mapstructure:"MAILGUN_API_BASE"`
	S3Endpoint     string `mapstructure:"S3_ENDPOINT"`
	S3Region       string `mapstructure:"S3_REGION"`
	S3Bucket       string `mapstructure:"S3_BUCKET"`
	S3AccessKey    string `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey    string `mapstructure:"S3_SECRET_KEY"`
}

func Load() (*Config, error) {
	// Set environment variable prefix for FunderMaps
	viper.SetEnvPrefix("FM")

	// Default values
	viper.SetDefault("SERVER_PORT", 3_000)
	viper.SetDefault("DATABASE_URL", "postgres://postgres:password@localhost:5432/fundermaps")
	viper.SetDefault("JWT_SECRET", utils.GenerateRandomString(32))
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
	viper.BindEnv("JWT_SECRET", "FM_JWT_SECRET", "JWT_SECRET")

	viper.BindEnv("MAILGUN_API_KEY", "FM_MAILGUN_API_KEY", "MAILGUN_API_KEY")
	viper.BindEnv("MAILGUN_DOMAIN", "FM_MAILGUN_DOMAIN", "MAILGUN_DOMAIN")
	viper.BindEnv("MAILGUN_API_BASE", "FM_MAILGUN_API_BASE", "MAILGUN_API_BASE")

	viper.BindEnv("S3_ENDPOINT", "FM_S3_ENDPOINT", "S3_ENDPOINT")
	viper.BindEnv("S3_REGION", "FM_S3_REGION", "S3_REGION")
	viper.BindEnv("S3_BUCKET", "FM_S3_BUCKET", "S3_BUCKET")
	viper.BindEnv("S3_ACCESS_KEY", "FM_S3_ACCESS_KEY", "S3_ACCESS_KEY")
	viper.BindEnv("S3_SECRET_KEY", "FM_S3_SECRET_KEY", "S3_SECRET_KEY")

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

	if cfg.ApplicationID == "" {
		return nil, fmt.Errorf("missing application ID")
	}

	// TODO: Move this to somewhere else
	Validate = validator.New()

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
