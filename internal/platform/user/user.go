package user

import (
	"errors"
	"fundermaps/internal/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// func (s *UserService) CreateProduct(product *Product) error {
// 	// ... any validation or business logic ...

// 	result := s.db.Create(product)
// 	if result.Error != nil {
// 		return fmt.Errorf("failed to create product: %w", result.Error)
// 	}
// 	return nil
// }

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

func (s *UserService) IsLocked(user *database.User) bool {
	return user.AccessFailedCount >= 5
}
