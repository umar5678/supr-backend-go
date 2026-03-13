package imagekit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// UploadResponse represents the response from ImageKit upload API
type UploadResponse struct {
	FileID        string `json:"fileId"`
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	FilePath      string `json:"filePath"`
	URL           string `json:"url"`
	FileType      string `json:"fileType"`
	Mime          string `json:"mime"`
	Height        int    `json:"height,omitempty"`
	Width         int    `json:"width,omitempty"`
	CreatedAt     string `json:"createdAt"`
	IsPrivateFile bool   `json:"isPrivateFile"`
}

// UploadError represents an error response from ImageKit API
type UploadError struct {
	Message string `json:"message"`
	Help    string `json:"help"`
}

// UploadDocumentToImageKit uploads a document to ImageKit
// It handles the server-side upload using ImageKit's API
func UploadDocumentToImageKit(
	cfg *config.Config,
	file *multipart.FileHeader,
	documentType string,
	username string,
) (*UploadResponse, error) {
	if cfg.Upload.ImageKit.PrivateKey == "" || cfg.Upload.ImageKit.URLEndpoint == "" {
		return nil, fmt.Errorf("ImageKit configuration incomplete: missing private key or URL endpoint")
	}

	// Get the folder based on document type
	folder := GetDocumentFolder(documentType)

	// Read file content
	src, err := file.Open()
	if err != nil {
		logger.Error("failed to open file", "error", err, "filename", file.Filename)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	defer src.Close()

	fileContent, err := io.ReadAll(src)
	if err != nil {
		logger.Error("failed to read file content", "error", err, "filename", file.Filename)
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Create multipart form data for upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file content
	fileWriter, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		logger.Error("failed to create form file", "error", err)
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := fileWriter.Write(fileContent); err != nil {
		logger.Error("failed to write file content", "error", err)
		return nil, fmt.Errorf("failed to write file to form: %w", err)
	}

	// Add folder
	if err := writer.WriteField("folder", folder); err != nil {
		logger.Error("failed to write folder field", "error", err)
		return nil, fmt.Errorf("failed to write folder field: %w", err)
	}

	// Build new filename with username and document type
	// Format: {username}_{documenttype}_{originalfilename}
	originalFileName := sanitizeFileName(file.Filename)
	sanitizedUsername := strings.ReplaceAll(strings.ToLower(username), " ", "_")
	fileName := fmt.Sprintf("%s_%s_%s", sanitizedUsername, documentType, originalFileName)

	if err := writer.WriteField("fileName", fileName); err != nil {
		logger.Error("failed to write fileName field", "error", err)
		return nil, fmt.Errorf("failed to write fileName field: %w", err)
	}

	// Close the writer to finalize multipart data
	if err := writer.Close(); err != nil {
		logger.Error("failed to close multipart writer", "error", err)
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://upload.imagekit.io/api/v1/files/upload", body)
	if err != nil {
		logger.Error("failed to create HTTP request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add basic auth with private key
	req.SetBasicAuth(cfg.Upload.ImageKit.PrivateKey, "")

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to upload to ImageKit", "error", err, "filename", file.Filename)
		return nil, fmt.Errorf("failed to upload to ImageKit: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read ImageKit response", "error", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		var uploadErr UploadError
		if err := json.Unmarshal(respBody, &uploadErr); err != nil {
			logger.Error("failed to parse ImageKit error response",
				"error", err,
				"statusCode", resp.StatusCode,
				"responseBody", string(respBody))
			return nil, fmt.Errorf("ImageKit upload failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		logger.Error("ImageKit upload failed",
			"message", uploadErr.Message,
			"help", uploadErr.Help,
			"statusCode", resp.StatusCode)
		return nil, fmt.Errorf("ImageKit upload failed: %s", uploadErr.Message)
	}

	// Parse successful response
	var uploadResp UploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		logger.Error("failed to parse ImageKit response", "error", err, "responseBody", string(respBody))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.Info("document uploaded to ImageKit successfully",
		"fileID", uploadResp.FileID,
		"filename", uploadResp.Name,
		"path", uploadResp.FilePath,
		"size", uploadResp.Size)

	return &uploadResp, nil
}

// GenerateAuthenticationToken generates a client-side upload token for ImageKit
// Token generation for client-side uploads without SDK
func GenerateAuthenticationTokenManual(cfg *config.Config) (map[string]string, error) {
	if cfg.Upload.ImageKit.PrivateKey == "" {
		return nil, fmt.Errorf("ImageKit private key not configured")
	}

	// Generate token
	token := base64.StdEncoding.EncodeToString([]byte(cfg.Upload.ImageKit.PrivateKey))

	return map[string]string{
		"token":       token,
		"publicKey":   cfg.Upload.ImageKit.PublicKey,
		"urlEndpoint": cfg.Upload.ImageKit.URLEndpoint,
	}, nil
}

// SignRequestForUpload creates a signature for client-side upload requests
// This is used when uploading files directly from client to ImageKit
func SignRequestForUpload(cfg *config.Config, expireIn int) (map[string]string, error) {
	if cfg.Upload.ImageKit.PrivateKey == "" {
		return nil, fmt.Errorf("ImageKit private key not configured")
	}

	// Create timestamp
	timestamp := strconv.FormatInt(time.Now().Unix()+int64(expireIn), 10)

	// Create string to sign
	auth := cfg.Upload.ImageKit.PrivateKey + timestamp
	hash := hmac.New(sha1.New, []byte(cfg.Upload.ImageKit.PrivateKey))
	hash.Write([]byte(timestamp))
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return map[string]string{
		"signature":   signature,
		"token":       base64.StdEncoding.EncodeToString([]byte(auth)),
		"expire":      timestamp,
		"publicKey":   cfg.Upload.ImageKit.PublicKey,
		"urlEndpoint": cfg.Upload.ImageKit.URLEndpoint,
	}, nil
}

// sanitizeFileName removes special characters from filename
// ImageKit might have restrictions on filenames
func sanitizeFileName(filename string) string {
	// Remove any path components
	if idx := strings.LastIndexAny(filename, "/\\"); idx >= 0 {
		filename = filename[idx+1:]
	}

	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")

	// Remove any characters that are not alphanumeric, hyphen, underscore, or dot
	var result strings.Builder
	for _, char := range filename {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.' {
			result.WriteRune(char)
		}
	}

	sanitized := result.String()
	if sanitized == "" {
		return "file_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return sanitized
}

// GetImageKitURL constructs the full ImageKit URL for a file
func GetImageKitURL(cfg *config.Config, filePath string) string {
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}
	return cfg.Upload.ImageKit.URLEndpoint + filePath
}

// DeleteFileFromImageKit deletes a file from ImageKit using its file ID
func DeleteFileFromImageKit(cfg *config.Config, fileID string) error {
	if cfg.Upload.ImageKit.PrivateKey == "" {
		return fmt.Errorf("ImageKit configuration incomplete: missing private key")
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://api.imagekit.io/v1/files/%s", fileID), nil)
	if err != nil {
		logger.Error("failed to create delete request", "error", err, "fileID", fileID)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(cfg.Upload.ImageKit.PrivateKey, "")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to delete from ImageKit", "error", err, "fileID", fileID)
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("ImageKit delete failed",
			"statusCode", resp.StatusCode,
			"responseBody", string(body),
			"fileID", fileID)
		return fmt.Errorf("ImageKit delete failed with status %d", resp.StatusCode)
	}

	logger.Info("file deleted from ImageKit successfully", "fileID", fileID)
	return nil
}
