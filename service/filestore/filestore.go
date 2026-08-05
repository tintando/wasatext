package filestore

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid"
)

// FileStore handles file operations for photos
type FileStore struct {
	basePath string
}

// NewFileStore creates a new file store instance
func NewFileStore(basePath string) *FileStore {
	return &FileStore{
		basePath: basePath,
	}
}

// SavePhoto saves an uploaded photo and returns the file path
func (fs *FileStore) SavePhoto(data io.Reader, contentType string) (string, error) {
	// Validate content type
	if !IsValidImageType(contentType) {
		return "", fmt.Errorf("invalid image type: %s", contentType)
	}

	// Generate unique filename
	photoID, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("error generating photo ID: %w", err)
	}

	// Get file extension from content type
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		return "", fmt.Errorf("unable to determine file extension for type: %s", contentType)
	}

	filename := photoID.String() + extensions[0]
	filePath := filepath.Join(fs.basePath, filename)

	// Ensure directory exists
	if err := os.MkdirAll(fs.basePath, 0755); err != nil {
		return "", fmt.Errorf("error creating photo directory: %w", err)
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creating photo file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	_, err = io.Copy(file, data)
	if err != nil {
		// Clean up on error
		os.Remove(filePath)
		return "", fmt.Errorf("error writing photo data: %w", err)
	}

	return filename, nil
}

// GetPhotoPath returns the full file path for a photo
func (fs *FileStore) GetPhotoPath(filename string) string {
	return filepath.Join(fs.basePath, filename)
}

// DeletePhoto removes a photo file
func (fs *FileStore) DeletePhoto(filename string) error {
	filePath := filepath.Join(fs.basePath, filename)
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error deleting photo: %w", err)
	}
	return nil
}

// IsValidImageType reports whether contentType is a supported image format.
func IsValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
	}

	contentType = strings.ToLower(contentType)
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// ValidateImageFile validates that a file is actually an image by reading its header
func ValidateImageFile(data io.Reader) (string, error) {
	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	n, err := data.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading file header: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	if !IsValidImageType(contentType) {
		return "", fmt.Errorf("file is not a valid image: detected type %s", contentType)
	}

	return contentType, nil
}
