package handlers

import (
	"fundermaps/internal/database"

	"github.com/gofiber/fiber/v2"
)

type UserInfo struct {
	Sub         string  `json:"sub"`
	Name        string  `json:"name"`
	GivenName   *string `json:"given_name,omitempty"`
	FamilyName  *string `json:"family_name,omitempty"`
	Email       string  `json:"email"`
	Picture     *string `json:"picture,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

func GetUserInfo(c *fiber.Ctx) error {
	user := c.Locals("user").(database.User)

	var name string
	if user.GivenName != nil {
		name = *user.GivenName
	}
	if user.LastName != nil {
		if name != "" {
			name += " "
		}
		name += *user.LastName
	}

	userInfo := UserInfo{
		Sub:         user.ID.String(),
		Name:        name,
		GivenName:   user.GivenName,
		FamilyName:  user.LastName,
		Email:       user.Email,
		Picture:     user.Avatar,
		PhoneNumber: user.PhoneNumber,
	}
	return c.JSON(userInfo)
}
