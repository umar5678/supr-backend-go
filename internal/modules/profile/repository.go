package profile

import (
    "context"
    "github.com/umar5678/go-backend/internal/models"
    "gorm.io/gorm"
)

type Repository interface {
    // Emergency contacts
    UpdateEmergencyContact(ctx context.Context, userID, name, phone string) error
    
    // Referrals
    GenerateReferralCode(ctx context.Context, userID, code string) error
    FindUserByReferralCode(ctx context.Context, code string) (*models.User, error)
    FindUserByID(ctx context.Context, userID string) (*models.User, error)
    ApplyReferralCode(ctx context.Context, userID, referredBy string) error
    GetReferralStats(ctx context.Context, userID string) (count int64, bonus float64, err error)
    
    // KYC
    CreateKYC(ctx context.Context, kyc *models.UserKYC) error
    FindKYCByUserID(ctx context.Context, userID string) (*models.UserKYC, error)
    UpdateKYCStatus(ctx context.Context, kycID, status, reason string) error
    
    // Saved locations
    CreateLocation(ctx context.Context, location *models.SavedLocation) error
    FindLocationsByUserID(ctx context.Context, userID string) ([]*models.SavedLocation, error)
    FindLocationByID(ctx context.Context, id string) (*models.SavedLocation, error)
    UpdateLocation(ctx context.Context, location *models.SavedLocation) error
    DeleteLocation(ctx context.Context, id string) error
    SetDefaultLocation(ctx context.Context, userID, locationID string) error
    GetRecentLocations(ctx context.Context, userID string, limit int) ([]*models.SavedLocation, error)
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) UpdateEmergencyContact(ctx context.Context, userID, name, phone string) error {
    return r.db.WithContext(ctx).
        Model(&models.User{}).
        Where("id = ?", userID).
        Updates(map[string]interface{}{
            "emergency_contact_name":  name,
            "emergency_contact_phone": phone,
        }).Error
}

func (r *repository) GenerateReferralCode(ctx context.Context, userID, code string) error {
    return r.db.WithContext(ctx).
        Model(&models.User{}).
        Where("id = ?", userID).
        Update("referral_code", code).Error
}

func (r *repository) FindUserByReferralCode(ctx context.Context, code string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).
        Where("referral_code = ?", code).
        First(&user).Error
    return &user, err
}

func (r *repository) FindUserByID(ctx context.Context, userID string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).
        Where("id = ?", userID).
        First(&user).Error
    return &user, err
}

func (r *repository) ApplyReferralCode(ctx context.Context, userID, referredBy string) error {
    return r.db.WithContext(ctx).
        Model(&models.User{}).
        Where("id = ?", userID).
        Update("referred_by", referredBy).Error
}

func (r *repository) GetReferralStats(ctx context.Context, userID string) (count int64, bonus float64, err error) {
    var user models.User
    if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
        return 0, 0, err
    }
    
    // Count referrals
    r.db.WithContext(ctx).
        Model(&models.User{}).
        Where("referred_by = ?", user.ReferralCode).
        Count(&count)
    
    // Calculate bonus ($5 per referral)
    bonus = float64(count) * 5.0
    
    return count, bonus, nil
}

func (r *repository) CreateKYC(ctx context.Context, kyc *models.UserKYC) error {
    return r.db.WithContext(ctx).Create(kyc).Error
}

func (r *repository) FindKYCByUserID(ctx context.Context, userID string) (*models.UserKYC, error) {
    var kyc models.UserKYC
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        First(&kyc).Error
    return &kyc, err
}

func (r *repository) UpdateKYCStatus(ctx context.Context, kycID, status, reason string) error {
    updates := map[string]interface{}{
        "status": status,
    }
    if reason != "" {
        updates["rejection_reason"] = reason
    }
    if status == "approved" {
        updates["verified_at"] = gorm.Expr("NOW()")
    }
    
    return r.db.WithContext(ctx).
        Model(&models.UserKYC{}).
        Where("id = ?", kycID).
        Updates(updates).Error
}

func (r *repository) CreateLocation(ctx context.Context, location *models.SavedLocation) error {
    return r.db.WithContext(ctx).Create(location).Error
}

func (r *repository) FindLocationsByUserID(ctx context.Context, userID string) ([]*models.SavedLocation, error) {
    var locations []*models.SavedLocation
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Order("is_default DESC, created_at DESC").
        Find(&locations).Error
    return locations, err
}

func (r *repository) FindLocationByID(ctx context.Context, id string) (*models.SavedLocation, error) {
    var location models.SavedLocation
    err := r.db.WithContext(ctx).
        Where("id = ?", id).
        First(&location).Error
    return &location, err
}

func (r *repository) UpdateLocation(ctx context.Context, location *models.SavedLocation) error {
    return r.db.WithContext(ctx).Save(location).Error
}

func (r *repository) DeleteLocation(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).
        Delete(&models.SavedLocation{}, "id = ?", id).Error
}

func (r *repository) SetDefaultLocation(ctx context.Context, userID, locationID string) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Remove default from all locations
        if err := tx.Model(&models.SavedLocation{}).
            Where("user_id = ?", userID).
            Update("is_default", false).Error; err != nil {
            return err
        }
        
        // Set new default
        return tx.Model(&models.SavedLocation{}).
            Where("id = ? AND user_id = ?", locationID, userID).
            Update("is_default", true).Error
    })
}

func (r *repository) GetRecentLocations(ctx context.Context, userID string, limit int) ([]*models.SavedLocation, error) {
    var locations []*models.SavedLocation
    err := r.db.WithContext(ctx).Raw(`
        SELECT DISTINCT ON (pickup_address) 
            uuid_generate_v4() as id,
            ? as user_id,
            'recent' as label,
            pickup_address as address,
            pickup_lat as latitude,
            pickup_lon as longitude,
            false as is_default,
            MAX(requested_at) as created_at,
            MAX(requested_at) as updated_at
        FROM rides
        WHERE rider_id = ?
        GROUP BY pickup_address, pickup_lat, pickup_lon
        ORDER BY created_at DESC
        LIMIT ?
    `, userID, userID, limit).Scan(&locations).Error
    return locations, err
}
