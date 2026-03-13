package documents

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"gorm.io/gorm"
)

type Repository interface {
	CreateDocument(ctx context.Context, doc *models.Document) error
	GetDocumentByID(ctx context.Context, docID string) (*models.Document, error)
	GetDocumentsByDriverID(ctx context.Context, driverID string) ([]*models.Document, error)
	GetDocumentsByServiceProviderID(ctx context.Context, providerID string) ([]*models.Document, error)
	GetDocumentsByUserID(ctx context.Context, userID string) ([]*models.Document, error)
	GetDocumentsByType(ctx context.Context, userID, docType string) ([]*models.Document, error)
	GetDocumentsPaginated(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Document, int64, error)
	UpdateDocumentStatus(ctx context.Context, docID, status string, verifiedBy *string, verifiedAt *time.Time, rejectionReason string) error
	DeleteDocument(ctx context.Context, docID string) error
	GetPendingDocuments(ctx context.Context, limit int) ([]*models.Document, error)
	GetDocumentsByStatus(ctx context.Context, status string) ([]*models.Document, error)
	CreateVerificationLog(ctx context.Context, log *models.DocumentVerificationLog) error
	GetVerificationLogs(ctx context.Context, docID string) ([]*models.DocumentVerificationLog, error)
	// Profile verification methods
	UpdateDriverProfileVerification(ctx context.Context, driverID string, isVerified bool) error
	UpdateServiceProviderProfileVerification(ctx context.Context, providerID string, isVerified bool) error
	CountVerifiedDocumentsByDriverID(ctx context.Context, driverID string) (int, error)
	CountVerifiedDocumentsByServiceProviderID(ctx context.Context, providerID string) (int, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateDocument(ctx context.Context, doc *models.Document) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		logger.Error("failed to create document", "error", err, "userID", doc.UserID)
		return err
	}
	logger.Info("document created successfully", "docID", doc.ID, "userID", doc.UserID, "docType", doc.DocumentType)
	return nil
}

func (r *repository) GetDocumentByID(ctx context.Context, docID string) (*models.Document, error) {
	var doc models.Document
	if err := r.db.WithContext(ctx).Where("id = ?", docID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not found")
		}
		logger.Error("failed to fetch document", "error", err, "docID", docID)
		return nil, err
	}
	return &doc, nil
}

func (r *repository) GetDocumentsByDriverID(ctx context.Context, driverID string) ([]*models.Document, error) {
	var docs []*models.Document
	if err := r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		logger.Error("failed to fetch driver documents", "error", err, "driverID", driverID)
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetDocumentsByServiceProviderID(ctx context.Context, providerID string) ([]*models.Document, error) {
	var docs []*models.Document
	if err := r.db.WithContext(ctx).
		Where("service_provider_id = ?", providerID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		logger.Error("failed to fetch service provider documents", "error", err, "providerID", providerID)
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetDocumentsByUserID(ctx context.Context, userID string) ([]*models.Document, error) {
	var docs []*models.Document
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		logger.Error("failed to fetch user documents", "error", err, "userID", userID)
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetDocumentsByType(ctx context.Context, userID, docType string) ([]*models.Document, error) {
	var docs []*models.Document
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND document_type = ?", userID, docType).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		logger.Error("failed to fetch documents by type", "error", err, "userID", userID, "docType", docType)
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetDocumentsPaginated(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Document, int64, error) {
	var docs []*models.Document
	var total int64

	query := r.db.WithContext(ctx)

	// Apply filters
	if userID, ok := filters["user_id"]; ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if driverID, ok := filters["driver_id"]; ok && driverID != "" {
		query = query.Where("driver_id = ?", driverID)
	}
	if providerID, ok := filters["service_provider_id"]; ok && providerID != "" {
		query = query.Where("service_provider_id = ?", providerID)
	}
	if docType, ok := filters["document_type"]; ok && docType != "" {
		query = query.Where("document_type = ?", docType)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	query = query.Where("deleted_at IS NULL")

	// Get total count
	if err := query.Model(&models.Document{}).Count(&total).Error; err != nil {
		logger.Error("failed to count documents", "error", err)
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&docs).Error; err != nil {
		logger.Error("failed to fetch paginated documents", "error", err)
		return nil, 0, err
	}

	return docs, total, nil
}

func (r *repository) UpdateDocumentStatus(ctx context.Context, docID, status string, verifiedBy *string, verifiedAt *time.Time, rejectionReason string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if verifiedBy != nil {
		updates["verified_by"] = *verifiedBy
	}
	if verifiedAt != nil {
		updates["verified_at"] = *verifiedAt
	}
	if rejectionReason != "" {
		updates["rejection_reason"] = rejectionReason
	}

	if err := r.db.WithContext(ctx).Model(&models.Document{}).Where("id = ?", docID).Updates(updates).Error; err != nil {
		logger.Error("failed to update document status", "error", err, "docID", docID)
		return err
	}
	logger.Info("document status updated", "docID", docID, "status", status)
	return nil
}

func (r *repository) DeleteDocument(ctx context.Context, docID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", docID).Delete(&models.Document{}).Error; err != nil {
		logger.Error("failed to delete document", "error", err, "docID", docID)
		return err
	}
	logger.Info("document deleted", "docID", docID)
	return nil
}

func (r *repository) GetPendingDocuments(ctx context.Context, limit int) ([]*models.Document, error) {
	var docs []*models.Document
	if err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Where("deleted_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&docs).Error; err != nil {
		logger.Error("failed to fetch pending documents", "error", err)
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetDocumentsByStatus(ctx context.Context, status string) ([]*models.Document, error) {
	var docs []*models.Document
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		logger.Error("failed to fetch documents by status", "error", err, "status", status)
		return nil, err
	}
	return docs, nil
}

func (r *repository) CreateVerificationLog(ctx context.Context, log *models.DocumentVerificationLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		logger.Error("failed to create verification log", "error", err, "docID", log.DocumentID)
		return err
	}
	logger.Info("verification log created", "docID", log.DocumentID, "action", log.Action)
	return nil
}

func (r *repository) GetVerificationLogs(ctx context.Context, docID string) ([]*models.DocumentVerificationLog, error) {
	var logs []*models.DocumentVerificationLog
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", docID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		logger.Error("failed to fetch verification logs", "error", err, "docID", docID)
		return nil, err
	}
	return logs, nil
}

// UpdateDriverProfileVerification updates the verification status of a driver profile
func (r *repository) UpdateDriverProfileVerification(ctx context.Context, driverID string, isVerified bool) error {
	if err := r.db.WithContext(ctx).
		Model(&models.DriverProfile{}).
		Where("id = ?", driverID).
		Update("is_verified", isVerified).Error; err != nil {
		logger.Error("failed to update driver profile verification", "error", err, "driverID", driverID)
		return err
	}
	logger.Info("driver profile verification updated", "driverID", driverID, "isVerified", isVerified)
	return nil
}

// UpdateServiceProviderProfileVerification updates the verification status of a service provider profile
func (r *repository) UpdateServiceProviderProfileVerification(ctx context.Context, providerID string, isVerified bool) error {
	if err := r.db.WithContext(ctx).
		Model(&models.ServiceProviderProfile{}).
		Where("id = ?", providerID).
		Update("is_verified", isVerified).Error; err != nil {
		logger.Error("failed to update service provider profile verification", "error", err, "providerID", providerID)
		return err
	}
	logger.Info("service provider profile verification updated", "providerID", providerID, "isVerified", isVerified)
	return nil
}

// CountVerifiedDocumentsByDriverID counts verified documents for a driver
func (r *repository) CountVerifiedDocumentsByDriverID(ctx context.Context, driverID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("driver_id = ? AND status = ?", driverID, "verified").
		Count(&count).Error; err != nil {
		logger.Error("failed to count verified documents", "error", err, "driverID", driverID)
		return 0, err
	}
	return int(count), nil
}

// CountVerifiedDocumentsByServiceProviderID counts verified documents for a service provider
func (r *repository) CountVerifiedDocumentsByServiceProviderID(ctx context.Context, providerID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("service_provider_id = ? AND status = ?", providerID, "verified").
		Count(&count).Error; err != nil {
		logger.Error("failed to count verified documents", "error", err, "providerID", providerID)
		return 0, err
	}
	return int(count), nil
}

