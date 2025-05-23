package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"fundermaps/app/config"
)

func Open(c *config.Config) (*gorm.DB, error) {
	gromLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	db, err := gorm.Open(postgres.Open(c.DatabaseURL), &gorm.Config{
		Logger: gromLogger,
	})
	if err != nil {
		log.Fatal(err)
	}

	return db, err
}
