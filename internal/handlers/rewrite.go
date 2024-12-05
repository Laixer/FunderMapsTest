package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const domain string = "cld.sh"

type UrlRewrite struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code" gorm:"unique"`
	Domain      string    `json:"domain"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	HitCount    int       `json:"hit_count"`
	Metadata    string    `json:"metadata"`
	IsActive    bool      `json:"is_active"`
}

func (u *UrlRewrite) TableName() string {
	return "application.url_rewrite"
}

func GetRewriteUrl(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	shortCode := c.Params("short_code")

	var urlRewrite UrlRewrite
	result := db.Where("domain = ? AND short_code = ?", domain, shortCode).First(&urlRewrite)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).SendString("URL not found")
	}

	return c.Redirect(urlRewrite.OriginalURL)
}
