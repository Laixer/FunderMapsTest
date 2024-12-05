package models

import "time"

type User struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	GivenName         string    `json:"given_name"`
	LastName          string    `json:"last_name"`
	Email             string    `json:"email"`
	NormalizedEmail   string    `json:"normalized_email"`
	Avatar            string    `json:"avatar"`
	JobTitle          string    `json:"job_title"`
	PasswordHash      string    `json:"password_hash"`
	PhoneNumber       string    `json:"phone_number"`
	AccessFailedCount int       `json:"access_failed_count"`
	Role              string    `json:"role"`
	LastLogin         time.Time `json:"last_login"`
	LoginCount        int       `json:"login_count"`
}
