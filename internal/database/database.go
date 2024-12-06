package database

import (
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"fundermaps/internal/config"
)

func Connect(c *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(c.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("GORM connected to database")

	return db, err
}
