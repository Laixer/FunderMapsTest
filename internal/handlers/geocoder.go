package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// type Contractor struct {
// 	ID   int    `json:"id" gorm:"primaryKey"`
// 	Name string `json:"name"`
// }

// func (u *Contractor) TableName() string {
// 	return "application.contractor"
// }s

func GetGeocoder(c *fiber.Ctx) error {
	// db := c.Locals("db").(*gorm.DB)

	// var contractors []Contractor
	// db.Order("id").Find(&contractors)

	// return c.JSON(contractors)

	return c.SendStatus(fiber.StatusNotImplemented)
}
