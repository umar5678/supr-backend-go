package documents

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/models"
	documentdto "github.com/umar5678/go-backend/internal/modules/documents/dto"
	imagekit "github.com/umar5678/go-backend/internal/services/imagekit"
	"github.com/umar5678/go-backend/internal/modules/notifications"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	UploadDocument(ctx context.Context, userID string, documentType string, file *multipart.FileHeader) (*documentdto.DocumentResponse, error)
	GetDocuments(ctx context.Context, userID string) ([]*documentdto.DocumentResponse, error)
	GetDocumentsByDriver(ctx context.Context, driverID string) ([]*documentdto.DocumentResponse, error)
	GetDocumentsByServiceProvider(ctx context.Context, providerID string) ([]*documentdto.DocumentResponse, error)
	GetDocumentsPaginated(ctx context.Context, filters map[string]interface{}, page, limit int) (*documentdto.DocumentListResponse, error)
	VerifyDocument(ctx context.Context, adminID, docID, status, rejectionReason string) (*documentdto.VerifyDocumentResponse, error)
	DeleteDocument(ctx context.Context, docID string) error
	GetPendingDocuments(ctx context.Context) ([]*documentdto.DocumentResponse, error)
	InitializeDriverProfile(ctx context.Context, driverID string) error
	InitializeServiceProviderProfile(ctx context.Context, providerID string) error
}

type service struct {
	repo Repository
	cfg  *config.Config
	eventProducer notifications.EventProducer
}

func NewService(repo Repository, cfg *config.Config, eventProducer notifications.EventProducer) Service {
	return &service{
		repo:         repo,
		cfg:          cfg,
		eventProducer: eventProducer,
	}
}

func (s *service) UploadDocument(ctx context.Context, userID string, documentType string, file *multipart.FileHeader) (*documentdto.DocumentResponse, error) {
	if !imagekit.ValidateDocumentType(documentType) {
		return nil, response.BadRequest(fmt.Sprintf("Invalid document type: %s", documentType))
	}

	maxSize := s.cfg.Upload.ImageKit.DocumentsMaxSize
	if file.Size > maxSize {
		return nil, response.BadRequest(fmt.Sprintf("File size exceeds maximum allowed (%d bytes)", maxSize))
	}

	mimeType := file.Header.Get("Content-Type")
	allowedMimes := imagekit.AllowedDocumentMimeTypes()
	if !isValidMimeType(mimeType, allowedMimes) {
		return nil, response.BadRequest(fmt.Sprintf("Invalid file type: %s. Allowed types: %v", mimeType, allowedMimes))
	}

	logger.Info("document upload initiated",
		"userID", userID,
		"documentType", documentType,
		"fileName", file.Filename,
		"fileSize", file.Size,
		"mimeType", mimeType)

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("failed to fetch user for document upload", "error", err, "userID", userID)
		return nil, response.NotFoundError("User not found")
	}

	uploadResp, err := imagekit.UploadDocumentToImageKit(s.cfg, file, documentType, user.Name)
	if err != nil {
		logger.Error("failed to upload document to ImageKit", "error", err, "userID", userID, "fileName", file.Filename)
		return nil, response.InternalServerError("Failed to upload document to storage", err)
	}

	fileExt := getFileExt(file.Filename)

	doc := &models.Document{
		UserID:           userID,
		DocumentType:     documentType,
		FileName:         file.Filename,
		FileSize:         file.Size,
		MimeType:         mimeType,
		Status:           "pending",
		FileURL:          uploadResp.URL,
		ImageKitFileID:   uploadResp.FileID,
		ImageKitFilePath: uploadResp.FilePath,
		Metadata: map[string]interface{}{
			"uploadedAt":    time.Now().Format(time.RFC3339),
			"uploadSize":    uploadResp.Size,
			"fileExtension": fileExt,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if driverProfile, err := s.repo.GetDriverByUserID(ctx, userID); err == nil {
		doc.DriverID = &driverProfile.ID
		logger.Info("document linked to driver profile", "docID", doc.ID, "driverID", driverProfile.ID, "userID", userID)
	}

	if providerProfile, err := s.repo.GetServiceProviderByUserID(ctx, userID); err == nil {
		doc.ServiceProviderID = &providerProfile.ID
		logger.Info("document linked to service provider profile", "docID", doc.ID, "providerID", providerProfile.ID, "userID", userID)
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		logger.Error("failed to create document record", "error", err, "userID", userID)
		if delErr := imagekit.DeleteFileFromImageKit(s.cfg, uploadResp.FileID); delErr != nil {
			logger.Error("failed to cleanup ImageKit file after DB error", "error", delErr, "fileID", uploadResp.FileID)
		}
		return nil, response.InternalServerError("Failed to store document record", err)
	}

	logger.Info("document uploaded successfully",
		"docID", doc.ID,
		"userID", userID,
		"documentType", documentType,
		"imageKitFileID", uploadResp.FileID,
		"fileURL", uploadResp.URL)

	return toDocumentResponse(doc), nil
}

func (s *service) GetDocuments(ctx context.Context, userID string) ([]*documentdto.DocumentResponse, error) {
	docs, err := s.repo.GetDocumentsByUserID(ctx, userID)
	if err != nil {
		logger.Error("failed to fetch documents", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to fetch documents", err)
	}

	responses := make([]*documentdto.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toDocumentResponse(doc)
	}

	return responses, nil
}

func (s *service) GetDocumentsByDriver(ctx context.Context, driverID string) ([]*documentdto.DocumentResponse, error) {
	docs, err := s.repo.GetDocumentsByDriverID(ctx, driverID)
	if err != nil {
		logger.Error("failed to fetch driver documents", "error", err, "driverID", driverID)
		return nil, response.InternalServerError("Failed to fetch documents", err)
	}

	responses := make([]*documentdto.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toDocumentResponse(doc)
	}

	return responses, nil
}

func (s *service) GetDocumentsByServiceProvider(ctx context.Context, providerID string) ([]*documentdto.DocumentResponse, error) {
	docs, err := s.repo.GetDocumentsByServiceProviderID(ctx, providerID)
	if err != nil {
		logger.Error("failed to fetch service provider documents", "error", err, "providerID", providerID)
		return nil, response.InternalServerError("Failed to fetch documents", err)
	}

	responses := make([]*documentdto.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toDocumentResponse(doc)
	}

	return responses, nil
}

func (s *service) GetDocumentsPaginated(ctx context.Context, filters map[string]interface{}, page, limit int) (*documentdto.DocumentListResponse, error) {
	docs, total, err := s.repo.GetDocumentsPaginated(ctx, filters, page, limit)
	if err != nil {
		logger.Error("failed to fetch paginated documents", "error", err)
		return nil, response.InternalServerError("Failed to fetch documents", err)
	}

	responses := make([]*documentdto.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toDocumentResponse(doc)
	}

	return &documentdto.DocumentListResponse{
		Documents: responses,
		Total:     total,
		Page:      page,
		Limit:     limit,
	}, nil
}

func (s *service) VerifyDocument(ctx context.Context, adminID, docID, status, rejectionReason string) (*documentdto.VerifyDocumentResponse, error) {

	if status != "verified" && status != "rejected" {
		return nil, response.BadRequest("Invalid status. Must be 'verified' or 'rejected'")
	}

	doc, err := s.repo.GetDocumentByID(ctx, docID)
	if err != nil {
		return nil, response.NotFoundError("Document")
	}

	now := time.Now()
	if err := s.repo.UpdateDocumentStatus(ctx, docID, status, &adminID, &now, rejectionReason); err != nil {
		logger.Error("failed to update document status", "error", err, "docID", docID)
		s.publishDocumentsEvent(ctx, notifications.EventDocumentStatusUpdated, adminID, map[string]interface{}{
			"document_id": docID,
			"status":      status,
		})
		return nil, response.InternalServerError("Failed to verify document", err)
	}

	log := &models.DocumentVerificationLog{
		DocumentID: docID,
		AdminID:    adminID,
		Action:     status,
		Status:     status,
		Comments:   rejectionReason,
		CreatedAt:  now,
	}

	if err := s.repo.CreateVerificationLog(ctx, log); err != nil {
		logger.Error("failed to create verification log", "error", err, "docID", docID)
	}

	if doc.DriverID != nil {
		switch status {
		case "rejected":
			if err := s.repo.UpdateDriverProfileVerification(ctx, *doc.DriverID, false); err != nil {
				logger.Error("failed to update driver profile verification", "error", err, "driverID", *doc.DriverID)
			}
		case "verified":
			verifiedCount, err := s.repo.CountVerifiedDocumentsByDriverID(ctx, *doc.DriverID)
			if err != nil {
				logger.Error("failed to count verified documents", "error", err, "driverID", *doc.DriverID)
			} else {
				requiredDocsCount := 4
				if verifiedCount >= requiredDocsCount {
					if err := s.repo.UpdateDriverProfileVerification(ctx, *doc.DriverID, true); err != nil {
						logger.Error("failed to update driver profile verification", "error", err, "driverID", *doc.DriverID)
					}
					logger.Info("driver profile marked as verified",
						"driverID", *doc.DriverID,
						"verifiedDocuments", verifiedCount)
				}
			}
		}
	} else if doc.ServiceProviderID != nil {
		switch status {
		case "rejected":
			if err := s.repo.UpdateServiceProviderProfileVerification(ctx, *doc.ServiceProviderID, false); err != nil {
				logger.Error("failed to update service provider profile verification", "error", err, "providerID", *doc.ServiceProviderID)
			}
		case "verified":
			verifiedCount, err := s.repo.CountVerifiedDocumentsByServiceProviderID(ctx, *doc.ServiceProviderID)
			if err != nil {
				logger.Error("failed to count verified documents", "error", err, "providerID", *doc.ServiceProviderID)
			} else {
				requiredDocsCount := 2
				if verifiedCount >= requiredDocsCount {
					if err := s.repo.UpdateServiceProviderProfileVerification(ctx, *doc.ServiceProviderID, true); err != nil {
						logger.Error("failed to update service provider profile verification", "error", err, "providerID", *doc.ServiceProviderID)
					}
					logger.Info("service provider profile marked as verified",
						"providerID", *doc.ServiceProviderID,
						"verifiedDocuments", verifiedCount)
				}
			}
		}
	}

	logger.Info("document verified",
		"docID", docID,
		"status", status,
		"adminID", adminID)

	s.publishDocumentsEvent(ctx, notifications.EventDocumentStatusUpdated, adminID, map[string]interface{}{
		"document_id": docID,
		"status":      status,
	})
	return &documentdto.VerifyDocumentResponse{
		DocumentID:      docID,
		Status:          status,
		RejectionReason: rejectionReason,
		VerifiedAt:      now.Format(time.RFC3339),
		Message:         fmt.Sprintf("Document %s successfully", status),
	}, nil
}

func (s *service) DeleteDocument(ctx context.Context, docID string) error {
	doc, err := s.repo.GetDocumentByID(ctx, docID)
	if err != nil {
		return response.NotFoundError("Document")
	}

	if err := s.repo.DeleteDocument(ctx, docID); err != nil {
		logger.Error("failed to delete document", "error", err, "docID", docID)
		return response.InternalServerError("Failed to delete document", err)
	}

	logger.Info("document deleted", "docID", docID, "userID", doc.UserID)

	return nil
}

func (s *service) GetPendingDocuments(ctx context.Context) ([]*documentdto.DocumentResponse, error) {
	docs, err := s.repo.GetPendingDocuments(ctx, 100)
	if err != nil {
		logger.Error("failed to fetch pending documents", "error", err)
		return nil, response.InternalServerError("Failed to fetch pending documents", err)
	}

	responses := make([]*documentdto.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toDocumentResponse(doc)
	}

	return responses, nil
}

func (s *service) InitializeDriverProfile(ctx context.Context, driverID string) error {
	logger.Info("initializing driver profile for document verification", "driverID", driverID)

	if err := s.repo.UpdateDriverProfileVerification(ctx, driverID, false); err != nil {
		logger.Error("failed to initialize driver profile verification", "error", err, "driverID", driverID)
		return fmt.Errorf("failed to initialize driver profile: %w", err)
	}

	logger.Info("driver profile initialized for document verification", "driverID", driverID)
	return nil
}

func (s *service) InitializeServiceProviderProfile(ctx context.Context, providerID string) error {
	logger.Info("initializing service provider profile for document verification", "providerID", providerID)

	if err := s.repo.UpdateServiceProviderProfileVerification(ctx, providerID, false); err != nil {
		logger.Error("failed to initialize service provider profile verification", "error", err, "providerID", providerID)
		return fmt.Errorf("failed to initialize service provider profile: %w", err)
	}

	logger.Info("service provider profile initialized for document verification", "providerID", providerID)
	return nil
}

func toDocumentResponse(doc *models.Document) *documentdto.DocumentResponse {
	resp := &documentdto.DocumentResponse{
		ID:              doc.ID,
		UserID:          doc.UserID,
		DocumentType:    doc.DocumentType,
		FileName:        doc.FileName,
		FileURL:         doc.FileURL,
		FileSize:        doc.FileSize,
		MimeType:        doc.MimeType,
		Status:          doc.Status,
		RejectionReason: doc.RejectionReason,
		IsFront:         doc.IsFront,
		UploadedAt:      doc.CreatedAt.Format(time.RFC3339),
	}

	if doc.DriverID != nil {
		resp.DriverID = doc.DriverID
	}
	if doc.ServiceProviderID != nil {
		resp.ServiceProviderID = doc.ServiceProviderID
	}
	if doc.VerifiedBy != nil {
		resp.VerifiedBy = doc.VerifiedBy
	}
	if doc.VerifiedAt != nil {
		resp.VerifiedAt = &[]string{doc.VerifiedAt.Format(time.RFC3339)}[0]
	}
	if doc.ExpiryDate != nil {
		resp.ExpiryDate = &[]string{doc.ExpiryDate.Format(time.RFC3339)}[0]
	}

	return resp
}

func isValidMimeType(mimeType string, allowed []string) bool {
	for _, mime := range allowed {
		if strings.EqualFold(mimeType, mime) {
			return true
		}
	}
	return false
}

func getFileExt(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}


func (s *service) publishDocumentsEvent(ctx context.Context, eventType notifications.EventType, userID string, data map[string]interface{}) {
	if s.eventProducer == nil {
		logger.Debug("event producer not available, skipping event publication", "eventType", eventType, "userID", userID)
		return
	}

	payload := map[string]interface{}{
		"user_id":   userID,
		"timestamp": time.Now().UTC(),
	}

	for k, v := range data {
		payload[k] = v
	}

	if err := s.eventProducer.PublishEventWithKey(ctx, eventType, userID, payload); err != nil {
		logger.Error("failed to publish driver event",
			"error", err,
			"eventType", eventType,
			"userID", userID,
		)
	}
}
