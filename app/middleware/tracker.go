package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"fundermaps/app/database"
)

type ProductTracker struct {
	Name       string `json:"product"`
	BuildingID string `json:"building_id"`
	Identifier string `json:"identifier"`
}

func TrackerMiddleware(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	user := c.Locals("user").(database.User)

	if len(user.Organizations) == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden"})
	}

	firstOrganization := user.Organizations[0]

	n := c.Next()

	tracker := c.Locals("tracker").(ProductTracker)

	// FUTURE:
	// - drop building_id
	// - save request_id
	// - save user_id
	// - save status
	// - get the product_id from the endpoint name
	// - save the product_id to memcache
	var isRegistered bool
	db.Raw(`
		WITH register_product_request AS (
			INSERT INTO application.product_tracker(organization_id, product, building_id, identifier)
			SELECT ?, ?, ?, ?
			WHERE NOT EXISTS (
				SELECT  1
				FROM    application.product_tracker pt
				WHERE   pt.organization_id = ?
				AND     pt.product = ?
				AND     pt.identifier = ?
				AND     pt.create_date > CURRENT_TIMESTAMP - interval '24 hours'
			)
			RETURNING 1
		)
		SELECT EXISTS (SELECT 1 FROM register_product_request) AS is_registered
	`, firstOrganization.ID, tracker.Name, tracker.BuildingID, tracker.Identifier, firstOrganization.ID, tracker.Name, tracker.Identifier).Scan(&isRegistered)

	c.Set("X-Product-Registered", fmt.Sprintf("%t", isRegistered))

	return n
}
