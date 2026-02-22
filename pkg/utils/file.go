package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileValidationError represents file validation errors
type FileValidationError struct {
	Message string
}

func (e *FileValidationError) Error() string {
	return e.Message
}

// ValidateImageFile validates an uploaded image file
func ValidateImageFile(file *multipart.FileHeader, maxSize int64, allowedTypes []string) error {
	// Validate file size
	if file.Size > maxSize {
		return &FileValidationError{
			Message: fmt.Sprintf("file size exceeds maximum allowed size of %d bytes", maxSize),
		}
	}

	// Open file to check MIME type
	src, err := file.Open()
	if err != nil {
		return &FileValidationError{Message: "failed to open uploaded file"}
	}
	defer src.Close()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return &FileValidationError{Message: "failed to read file content"}
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer[:n])

	// Check if MIME type is allowed
	allowed := false
	for _, allowedType := range allowedTypes {
		if mimeType == allowedType {
			allowed = true
			break
		}
	}

	if !allowed {
		return &FileValidationError{
			Message: fmt.Sprintf("file type %s is not allowed. Allowed types: %s", mimeType, strings.Join(allowedTypes, ", ")),
		}
	}

	return nil
}

// SaveUploadedFile saves an uploaded file to the specified directory with the given filename
func SaveUploadedFile(file *multipart.FileHeader, directory string, filename string) (string, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Build full file path
	fullPath := filepath.Join(directory, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	return fullPath, nil
}

// DeleteFile deletes a file if it exists
func DeleteFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		// File doesn't exist, nothing to delete
		return nil
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileExtension returns the file extension from a filename
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	// Return extension without the dot
	return ext[1:]
}

// GenerateUniqueFilename generates a unique filename from a base name and extension
func GenerateUniqueFilename(baseName string, extension string) string {
	slug := Slugify(baseName)
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return fmt.Sprintf("logo-%s%s", slug, extension)
}
