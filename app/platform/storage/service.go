package storage

import (
	"fmt"
	"fundermaps/app/database"
	"fundermaps/pkg/utils"
	"path/filepath"
	"strings"

	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/s3/v2"
	"gorm.io/gorm"
)

// Constants for file paths and statuses
const (
	DefaultUploadPath = "user-data"
	DefaultFormField  = "files"

	// File statuses
	StatusUploaded   = "uploaded"
	StatusProcessing = "processing"
	StatusActive     = "active"
	StatusArchived   = "archived"

	// Random key length
	KeyLength = 16
)

// List of allowed file extensions
var AllowedExtensions = []string{
	"jpg", "jpeg", "png", "pdf",
	"doc", "docx", "xls", "xlsx",
	"csv", "txt", "zip", "ppt", "pptx",
}

// FileUploadResult contains information about uploaded files
type FileUploadResult struct {
	Files      []string `json:"files"`
	Key        string   `json:"key"`
	TotalSize  int64    `json:"total_size"`
	TotalFiles int      `json:"total_files"`
}

// StorageService defines methods for file storage operations
type StorageService interface {
	// IsFileExtensionAllowed checks if file extension is allowed
	IsFileExtensionAllowed(filename string) bool

	// UploadFile handles the complete file upload process
	UploadFile(c *fiber.Ctx, db *gorm.DB, formFieldName string) (*FileUploadResult, error)

	// UpdateFileStatus updates the status of files associated with a key
	UpdateFileStatus(db *gorm.DB, key string, status string) error
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
func (s *storageService) generateKeyName() string {
	return strings.ToLower(utils.GenerateRandomString(KeyLength))
}

// UploadFile handles the complete file upload process
func (s *storageService) UploadFile(c *fiber.Ctx, db *gorm.DB, formFieldName string) (*FileUploadResult, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	// Use default form field name if empty
	if formFieldName == "" {
		formFieldName = DefaultFormField
	}

	files := form.File[formFieldName]
	if len(files) == 0 {
		return nil, fmt.Errorf("no files uploaded")
	}

	keyName := s.generateKeyName()

	uploadedFiles := make([]string, 0)
	resourceFiles := []database.FileResource{}
	var totalSize int64

	for _, file := range files {
		if !s.IsFileExtensionAllowed(file.Filename) {
			continue
		}

		filePath := fmt.Sprintf("%s/%s/%s", DefaultUploadPath, keyName, file.Filename)
		if err := c.SaveFileToStorage(file, filePath, s.storage); err != nil {
			return nil, fmt.Errorf("failed to save file: %w", err)
		}

		uploadedFiles = append(uploadedFiles, file.Filename)
		totalSize += file.Size

		fileResource := database.FileResource{
			Key:              keyName,
			OriginalFilename: file.Filename,
			SizeBytes:        file.Size,
			MimeType:         file.Header.Get("Content-Type"),
			Status:           StatusUploaded,
		}
		resourceFiles = append(resourceFiles, fileResource)
	}

	if len(uploadedFiles) == 0 {
		return nil, fmt.Errorf("no valid files were uploaded")
	}

	if err := db.Create(&resourceFiles).Error; err != nil {
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return &FileUploadResult{
		Files:      uploadedFiles,
		Key:        keyName,
		TotalSize:  totalSize,
		TotalFiles: len(uploadedFiles),
	}, nil
}

// UpdateFileStatus updates the status of files associated with a key
func (s *storageService) UpdateFileStatus(db *gorm.DB, key string, status string) error {
	result := db.Model(&database.FileResource{}).
		Where("key = ?", key).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update file status: %w", result.Error)
	}

	return nil
}
