package storage

import (
	"fundermaps/pkg/utils"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/s3/v2"
)

// StorageService defines methods for file storage operations
type StorageService interface {
	// SaveFile saves a file to the storage
	SaveFile(file *multipart.FileHeader, path string, c *fiber.Ctx) error

	// IsFileExtensionAllowed checks if file extension is allowed
	IsFileExtensionAllowed(filename string) bool

	// GenerateKeyName generates a random key name for file storage
	GenerateKeyName() string
}

// storageService implements StorageService interface
type storageService struct {
	storage *s3.Storage
}

// NewStorageService creates a new StorageService
func NewStorageService(storage *s3.Storage) StorageService {
	return &storageService{
		storage: storage,
	}
}

// SaveFile saves a file to the storage
func (s *storageService) SaveFile(file *multipart.FileHeader, path string, c *fiber.Ctx) error {
	return c.SaveFileToStorage(file, path, s.storage)
}

// IsFileExtensionAllowed checks if file extension is allowed
func (s *storageService) IsFileExtensionAllowed(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	// Remove the dot from extension
	ext = ext[1:]

	return slices.Contains(AllowedExtensions, ext)
}

// GenerateKeyName generates a random key name for file storage
func (s *storageService) GenerateKeyName() string {
	return strings.ToLower(utils.GenerateRandomString(16))
}
