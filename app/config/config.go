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
	MailgunAPIKey  string `mapstructure:"MAILGUN_API_KEY"`
	MailgunDomain  string `mapstructure:"MAILGUN_DOMAIN"`
	MailgunAPIBase string `mapstructure:"MAILGUN_API_BASE"`
	S3Endpoint     string `mapstructure:"S3_ENDPOINT"`
	S3Region       string `mapstructure:"S3_REGION"`
	S3Bucket       string `mapstructure:"S3_BUCKET"`
	S3AccessKey    string `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey    string `mapstructure:"S3_SECRET_KEY"`
}

// # TODO: Use FM_ prefix for environment variables
func Load() (*Config, error) {
	viper.SetDefault("SERVER_PORT", 3_000)
	viper.SetDefault("DATABASE_URL", "postgres://postgres:password@localhost:5432/fundermaps")
	viper.SetDefault("JWT_SECRET", utils.GenerateRandomString(32))

	viper.AutomaticEnv()

	viper.BindEnv("APP_ID")

	viper.BindEnv("S3_ENDPOINT")
	viper.BindEnv("S3_REGION")
	viper.BindEnv("S3_BUCKET")
	viper.BindEnv("S3_ACCESS_KEY")
	viper.BindEnv("S3_SECRET_KEY")

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
