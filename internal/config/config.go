package config

import (
	"fmt"

	"github.com/spf13/viper"

	"fundermaps/pkg/utils"
)

type Config struct {
	ServerPort     int    `mapstructure:"SERVER_PORT"`
	DatabaseURL    string `mapstructure:"DATABASE_URL"`
	JWTSecret      string `mapstructure:"JWT_SECRET"`
	MailgunAPIKey  string `mapstructure:"MAILGUN_API_KEY"`
	MailgunDomain  string `mapstructure:"MAILGUN_DOMAIN"`
	MailgunAPIBase string `mapstructure:"MAILGUN_API_BASE"`
}

func Load() (*Config, error) {
	viper.SetDefault("SERVER_PORT", 3000)
	viper.SetDefault("DATABASE_URL", "postgres://postgres:password@localhost:5432/fundermaps")
	viper.SetDefault("JWT_SECRET", utils.GenerateRandomString(32))

	viper.SetEnvPrefix("FM")
	viper.AutomaticEnv()

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

	return &cfg, nil
}
