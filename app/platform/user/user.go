package user

import (
	"errors"
	"fundermaps/app/database"
	"fundermaps/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(user *database.User) error {
	result := s.db.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *UserService) GetUserByID(userID uuid.UUID) (*database.User, error) {
	var user database.User

	result := s.db.First(&user, "id = ?", userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (s *UserService) GetUserByEmail(email string) (*database.User, error) {
	var user database.User

	result := s.db.First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (s *UserService) Update(user *database.User) error {
	result := s.db.Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *UserService) UpdatePassword(user *database.User, password string) error {
	user.PasswordHash = utils.HashLegacyPassword(password)

	result := s.db.Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *UserService) IsLocked(user *database.User) bool {
	return user.AccessFailedCount >= 5
}

func (s *UserService) IncrementAccessFailedCount(user *database.User) error {
	user.AccessFailedCount++

	result := s.db.Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
