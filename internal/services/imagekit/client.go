package imagekit

import (
	"fmt"
	"strings"

	"github.com/umar5678/go-backend/internal/config"
)

// DocumentUploadChecks returns ImageKit validation checks for document uploads
// Restricts uploads to documents/ folder with 10MB max size
func DocumentUploadChecks() string {
	return `"request.folder" : "documents/" AND "file.size" <= "10mb"`
}

// BannerUploadChecks returns ImageKit validation checks for banner uploads
// Restricts uploads to offers/ folder with 5MB max size and image mime types
func BannerUploadChecks() string {
	return `"request.folder" : "offers/" AND "file.size" <= "5mb" AND "file.mime" : "image"`
}

// GenerateAuthenticationToken creates an authentication token for ImageKit client-side uploads
// This uses the PrivateKey to create a signed token
func GenerateAuthenticationToken(cfg *config.Config) (string, error) {
	if cfg.Upload.ImageKit.PrivateKey == "" {
		return "", fmt.Errorf("ImageKit private key not configured")
	}

	// For client-side uploads, you typically use imagekit.io SDK which handles token generation
	// This is a placeholder for the token generation logic
	// In production, you would use the imagekit-go SDK to properly generate the token

	return "", fmt.Errorf("token generation requires imagekit-go SDK implementation")
}

// GetDocumentFolder returns the documents folder path for the given document type
// Document types: "license", "aadhaar", "registration", "insurance", "trade-license", "profile-photo"
func GetDocumentFolder(documentType string) string {
	documentType = strings.ToLower(documentType)

	switch documentType {
	case "license", "driving-license":
		return "documents/licenses/"
	case "aadhaar", "aadhaar-card":
		return "documents/aadhaar/"
	case "registration", "vehicle-registration":
		return "documents/registration/"
	case "insurance", "vehicle-insurance":
		return "documents/insurance/"
	case "trade-license", "trade_license":
		return "documents/trade-license/"
	case "profile-photo", "profile_photo":
		return "documents/profile-photos/"
	default:
		return "documents/misc/"
	}
}

// ValidateDocumentType checks if the document type is valid
func ValidateDocumentType(docType string) bool {
	validTypes := map[string]bool{
		"license":              true,
		"driving-license":      true,
		"aadhaar":              true,
		"aadhaar-card":         true,
		"registration":         true,
		"vehicle-registration": true,
		"insurance":            true,
		"vehicle-insurance":    true,
		"trade-license":        true,
		"trade_license":        true,
		"profile-photo":        true,
		"profile_photo":        true,
	}
	return validTypes[strings.ToLower(docType)]
}

// AllowedDocumentMimeTypes returns the mime types allowed for document uploads
func AllowedDocumentMimeTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"application/pdf",
	}
}

// AllowedBannerMimeTypes returns the mime types allowed for banner uploads
func AllowedBannerMimeTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"image/gif",
	}
}
