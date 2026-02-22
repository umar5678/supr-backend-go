# Profile Module Documentation

## Overview

The Profile Module manages user profile data, preferences, settings, privacy controls, and personal information management. It provides a unified interface for users to view and update their profile across all user types (riders, drivers, service providers).

## Key Responsibilities

- User profile management (personal information)
- Avatar and profile photo management
- Privacy settings and data sharing preferences
- Notification preferences and settings
- Account security settings
- Profile visibility controls
- User preferences (language, theme, units)
- Account deletion and data export

## Architecture

### Handler Layer (`profile/handler.go`)

Handles HTTP requests related to profile operations.

**Key Endpoints:**

```go
GET /api/v1/profile                      // Get current user profile
PUT /api/v1/profile                      // Update profile information
POST /api/v1/profile/avatar              // Upload profile avatar
DELETE /api/v1/profile/avatar            // Delete avatar
GET /api/v1/profile/preferences          // Get user preferences
PUT /api/v1/profile/preferences          // Update preferences
GET /api/v1/profile/privacy              // Get privacy settings
PUT /api/v1/profile/privacy              // Update privacy settings
POST /api/v1/profile/change-email        // Change email address
POST /api/v1/profile/change-phone        // Change phone number
POST /api/v1/profile/change-password     // Change password
GET /api/v1/profile/data-export          // Request data export
DELETE /api/v1/profile                   // Delete account
GET /api/v1/profile/activity             // Get account activity log
```

### Service Layer (`profile/service.go`)

Contains business logic for profile operations.

**Key Methods:**

```go
func (s *ProfileService) GetProfile(ctx context.Context, userID string) (*ProfileResponse, error)
func (s *ProfileService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*ProfileResponse, error)
func (s *ProfileService) UploadAvatar(ctx context.Context, userID string, file *multipart.FileHeader) (string, error)
func (s *ProfileService) GetPreferences(ctx context.Context, userID string) (*PreferencesResponse, error)
func (s *ProfileService) UpdatePreferences(ctx context.Context, userID string, req *UpdatePreferencesRequest) (*PreferencesResponse, error)
func (s *ProfileService) GetPrivacySettings(ctx context.Context, userID string) (*PrivacySettingsResponse, error)
func (s *ProfileService) UpdatePrivacySettings(ctx context.Context, userID string, req *UpdatePrivacyRequest) (*PrivacySettingsResponse, error)
func (s *ProfileService) ChangeEmail(ctx context.Context, userID string, newEmail string) error
func (s *ProfileService) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error
func (s *ProfileService) DeleteAccount(ctx context.Context, userID string) error
func (s *ProfileService) ExportData(ctx context.Context, userID string) ([]byte, error)
func (s *ProfileService) GetActivityLog(ctx context.Context, userID string) ([]*ActivityLog, error)
```

### Repository Layer (`profile/repository.go`)

Manages database operations.

**Key Methods:**

```go
func (r *ProfileRepository) GetProfile(ctx context.Context, userID string) (*Profile, error)
func (r *ProfileRepository) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error
func (r *ProfileRepository) GetPreferences(ctx context.Context, userID string) (*Preferences, error)
func (r *ProfileRepository) UpdatePreferences(ctx context.Context, userID string, updates map[string]interface{}) error
func (r *ProfileRepository) GetPrivacySettings(ctx context.Context, userID string) (*PrivacySettings, error)
func (r *ProfileRepository) UpdatePrivacySettings(ctx context.Context, userID string, updates map[string]interface{}) error
func (r *ProfileRepository) LogActivity(ctx context.Context, activity *ActivityLog) error
func (r *ProfileRepository) GetActivityLog(ctx context.Context, userID string) ([]*ActivityLog, error)
```

## Data Models

### Profile

```go
type Profile struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    FirstName           string     `db:"first_name" json:"first_name"`
    LastName            string     `db:"last_name" json:"last_name"`
    Email               string     `db:"email" json:"email"`
    Phone               string     `db:"phone" json:"phone"`
    Avatar              *string    `db:"avatar" json:"avatar"`                       // S3/Cloud URL
    DateOfBirth         *time.Time `db:"date_of_birth" json:"date_of_birth"`
    Gender              *string    `db:"gender" json:"gender"`                       // Male, Female, Other, Prefer not to say
    Bio                 *string    `db:"bio" json:"bio"`
    Website             *string    `db:"website" json:"website"`
    
    // Location
    Country             *string    `db:"country" json:"country"`
    City                *string    `db:"city" json:"city"`
    
    // Contact Verification
    EmailVerified       bool       `db:"email_verified" json:"email_verified"`
    EmailVerifiedAt     *time.Time `db:"email_verified_at" json:"email_verified_at"`
    PhoneVerified       bool       `db:"phone_verified" json:"phone_verified"`
    PhoneVerifiedAt     *time.Time `db:"phone_verified_at" json:"phone_verified_at"`
    
    // Status
    IsCompleted         bool       `db:"is_completed" json:"is_completed"`           // Full profile setup completed
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### Preferences

```go
type Preferences struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    
    // Language and Localization
    Language            string     `db:"language" json:"language"`                   // en, es, fr, de, etc.
    Timezone            string     `db:"timezone" json:"timezone"`                   // America/New_York, etc.
    DistanceUnit        string     `db:"distance_unit" json:"distance_unit"`         // km, mi
    TemperatureUnit     string     `db:"temperature_unit" json:"temperature_unit"`   // C, F
    
    // Appearance
    ThemeMode           string     `db:"theme_mode" json:"theme_mode"`               // light, dark, auto
    DateFormat          string     `db:"date_format" json:"date_format"`             // MM/DD/YYYY, DD/MM/YYYY
    TimeFormat          string     `db:"time_format" json:"time_format"`             // 12h, 24h
    
    // Notifications
    EmailNotifications  bool       `db:"email_notifications" json:"email_notifications"`
    SMSNotifications    bool       `db:"sms_notifications" json:"sms_notifications"`
    PushNotifications   bool       `db:"push_notifications" json:"push_notifications"`
    InAppNotifications  bool       `db:"in_app_notifications" json:"in_app_notifications"`
    
    // Content Preferences
    ReceivePromotions   bool       `db:"receive_promotions" json:"receive_promotions"`
    ReceiveNewsletter   bool       `db:"receive_newsletter" json:"receive_newsletter"`
    
    // Data
    AllowAnalytics      bool       `db:"allow_analytics" json:"allow_analytics"`
    AllowCrashReports   bool       `db:"allow_crash_reports" json:"allow_crash_reports"`
    
    // Accessibility
    HighContrast        bool       `db:"high_contrast" json:"high_contrast"`
    LargeText           bool       `db:"large_text" json:"large_text"`
    ScreenReaderOptimized bool     `db:"screen_reader_optimized" json:"screen_reader_optimized"`
    
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### PrivacySettings

```go
type PrivacySettings struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    
    // Profile Visibility
    ProfileVisibility   string     `db:"profile_visibility" json:"profile_visibility"`   // public, friends, private
    ShowEmail           bool       `db:"show_email" json:"show_email"`
    ShowPhone           bool       `db:"show_phone" json:"show_phone"`
    ShowLocation        bool       `db:"show_location" json:"show_location"`
    ShowAvatar          bool       `db:"show_avatar" json:"show_avatar"`
    
    // Activity Privacy
    ShowOnlineStatus    bool       `db:"show_online_status" json:"show_online_status"`
    ShowLastSeen        bool       `db:"show_last_seen" json:"show_last_seen"`
    ShowActivityStatus  bool       `db:"show_activity_status" json:"show_activity_status"`
    
    // Data Sharing
    AllowThirdParty     bool       `db:"allow_third_party" json:"allow_third_party"`
    AllowLocationSharing bool      `db:"allow_location_sharing" json:"allow_location_sharing"`
    AllowContactImport  bool       `db:"allow_contact_import" json:"allow_contact_import"`
    
    // Search and Discovery
    SearchEngineIndexing bool      `db:"search_engine_indexing" json:"search_engine_indexing"`
    AllowInSearchResults bool      `db:"allow_in_search_results" json:"allow_in_search_results"`
    
    // Marketing
    AllowEmailMarketing bool       `db:"allow_email_marketing" json:"allow_email_marketing"`
    AllowSMSMarketing   bool       `db:"allow_sms_marketing" json:"allow_sms_marketing"`
    AllowPushMarketing  bool       `db:"allow_push_marketing" json:"allow_push_marketing"`
    
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### ActivityLog

```go
type ActivityLog struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    ActivityType        string     `db:"activity_type" json:"activity_type"`         // LOGIN, LOGOUT, UPDATE_PROFILE, CHANGE_PASSWORD, etc.
    Description         string     `db:"description" json:"description"`
    IpAddress           string     `db:"ip_address" json:"ip_address"`
    UserAgent           string     `db:"user_agent" json:"user_agent"`
    Status              string     `db:"status" json:"status"`                       // SUCCESS, FAILED
    Details             string     `db:"details" json:"details"`                     // JSON data
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}
```

## DTOs (Data Transfer Objects)

### ProfileResponse

```go
type ProfileResponse struct {
    ID                  string     `json:"id"`
    FirstName           string     `json:"first_name"`
    LastName            string     `json:"last_name"`
    Email               string     `json:"email"`
    Phone               string     `json:"phone"`
    Avatar              *string    `json:"avatar"`
    DateOfBirth         *time.Time `json:"date_of_birth"`
    Gender              *string    `json:"gender"`
    EmailVerified       bool       `json:"email_verified"`
    PhoneVerified       bool       `json:"phone_verified"`
    IsCompleted         bool       `json:"is_completed"`
    CreatedAt           time.Time  `json:"created_at"`
    UpdatedAt           time.Time  `json:"updated_at"`
}
```

### UpdateProfileRequest

```go
type UpdateProfileRequest struct {
    FirstName           *string    `json:"first_name"`
    LastName            *string    `json:"last_name"`
    DateOfBirth         *time.Time `json:"date_of_birth"`
    Gender              *string    `json:"gender"`
    Bio                 *string    `json:"bio"`
    Website             *string    `json:"website"`
    Country             *string    `json:"country"`
    City                *string    `json:"city"`
}
```

### PreferencesResponse

```go
type PreferencesResponse struct {
    Language            string     `json:"language"`
    Timezone            string     `json:"timezone"`
    DistanceUnit        string     `json:"distance_unit"`
    ThemeMode           string     `json:"theme_mode"`
    EmailNotifications  bool       `json:"email_notifications"`
    SMSNotifications    bool       `json:"sms_notifications"`
    PushNotifications   bool       `json:"push_notifications"`
    ReceivePromotions   bool       `json:"receive_promotions"`
    AllowAnalytics      bool       `json:"allow_analytics"`
}
```

### PrivacySettingsResponse

```go
type PrivacySettingsResponse struct {
    ProfileVisibility   string     `json:"profile_visibility"`
    ShowEmail           bool       `json:"show_email"`
    ShowPhone           bool       `json:"show_phone"`
    ShowLocation        bool       `json:"show_location"`
    ShowOnlineStatus    bool       `json:"show_online_status"`
    AllowThirdParty     bool       `json:"allow_third_party"`
    AllowLocationSharing bool      `json:"allow_location_sharing"`
    SearchEngineIndexing bool      `json:"search_engine_indexing"`
}
```

### ChangePasswordRequest

```go
type ChangePasswordRequest struct {
    CurrentPassword     string `json:"current_password" binding:"required"`
    NewPassword         string `json:"new_password" binding:"required,min=8"`
    ConfirmPassword     string `json:"confirm_password" binding:"required"`
}
```

## Database Schema

### profiles table

```sql
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    avatar TEXT,
    date_of_birth DATE,
    gender VARCHAR(20),
    bio TEXT,
    website TEXT,
    country VARCHAR(100),
    city VARCHAR(100),
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP,
    phone_verified BOOLEAN DEFAULT FALSE,
    phone_verified_at TIMESTAMP,
    is_completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);
CREATE INDEX idx_profiles_email ON profiles(email);
```

### preferences table

```sql
CREATE TABLE preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(100),
    distance_unit VARCHAR(10) DEFAULT 'km',
    temperature_unit VARCHAR(5) DEFAULT 'C',
    theme_mode VARCHAR(20) DEFAULT 'auto',
    date_format VARCHAR(20),
    time_format VARCHAR(10) DEFAULT '24h',
    email_notifications BOOLEAN DEFAULT TRUE,
    sms_notifications BOOLEAN DEFAULT TRUE,
    push_notifications BOOLEAN DEFAULT TRUE,
    in_app_notifications BOOLEAN DEFAULT TRUE,
    receive_promotions BOOLEAN DEFAULT TRUE,
    receive_newsletter BOOLEAN DEFAULT FALSE,
    allow_analytics BOOLEAN DEFAULT TRUE,
    allow_crash_reports BOOLEAN DEFAULT TRUE,
    high_contrast BOOLEAN DEFAULT FALSE,
    large_text BOOLEAN DEFAULT FALSE,
    screen_reader_optimized BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_preferences_user_id ON preferences(user_id);
```

### privacy_settings table

```sql
CREATE TABLE privacy_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    profile_visibility VARCHAR(20) DEFAULT 'public',
    show_email BOOLEAN DEFAULT FALSE,
    show_phone BOOLEAN DEFAULT FALSE,
    show_location BOOLEAN DEFAULT FALSE,
    show_avatar BOOLEAN DEFAULT TRUE,
    show_online_status BOOLEAN DEFAULT TRUE,
    show_last_seen BOOLEAN DEFAULT FALSE,
    show_activity_status BOOLEAN DEFAULT FALSE,
    allow_third_party BOOLEAN DEFAULT FALSE,
    allow_location_sharing BOOLEAN DEFAULT FALSE,
    allow_contact_import BOOLEAN DEFAULT FALSE,
    search_engine_indexing BOOLEAN DEFAULT TRUE,
    allow_in_search_results BOOLEAN DEFAULT TRUE,
    allow_email_marketing BOOLEAN DEFAULT FALSE,
    allow_sms_marketing BOOLEAN DEFAULT FALSE,
    allow_push_marketing BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_privacy_settings_user_id ON privacy_settings(user_id);
```

### activity_logs table

```sql
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(100),
    description TEXT,
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(50),
    details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_activity_logs_user_id (user_id),
    INDEX idx_activity_logs_created_at (created_at)
);

CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_activity_type ON activity_logs(activity_type);
```

## Use Cases

### Use Case 1: Profile Completion During Onboarding

```
1. User signs up (Auth module)
2. Redirected to complete profile
3. User enters first name, last name, DOB
4. User uploads profile avatar
5. Profile marked as is_completed = true
6. User can proceed to use app
```

### Use Case 2: Privacy and Preferences Management

```
1. User goes to settings
2. Views current privacy settings
3. Adjusts profile visibility to 'friends'
4. Disables showing phone number
5. Opt out of promotional emails
6. Settings saved immediately
7. Changes logged in activity log
```

### Use Case 3: Account Security Management

```
1. User navigates to Account Security
2. Can change password (requires current password)
3. Can change email (requires verification)
4. Can change phone (requires OTP)
5. Views login activity
6. Can see active sessions
7. Can logout from other devices
```

### Use Case 4: Account Deletion with Data Export

```
1. User requests data export
2. System prepares JSON with all user data
3. File sent to user's email
4. User optionally initiates account deletion
5. All user data scheduled for deletion after 30 days
6. Account marked as deleted
7. User cannot login
```

## Common Operations

### Get Profile

```go
handler := func(c *gin.Context) {
    userID := c.GetString("user_id")
    
    profile, err := s.profileService.GetProfile(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, profile)
}
```

### Update Profile

```go
handler := func(c *gin.Context) {
    userID := c.GetString("user_id")
    var req UpdateProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    profile, err := s.profileService.UpdateProfile(c.Request.Context(), userID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Log activity
    s.profileService.LogActivity(c.Request.Context(), userID, "UPDATE_PROFILE", "User updated profile")

    c.JSON(http.StatusOK, profile)
}
```

### Upload Avatar

```go
handler := func(c *gin.Context) {
    userID := c.GetString("user_id")
    
    file, err := c.FormFile("avatar")
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "No file provided"})
        return
    }

    // Validate file size (max 5MB)
    if file.Size > 5*1024*1024 {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "File too large"})
        return
    }

    // Validate file type
    allowedTypes := map[string]bool{"image/jpeg": true, "image/png": true}
    if !allowedTypes[file.Header.Get("Content-Type")] {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid file type"})
        return
    }

    avatarURL, err := s.profileService.UploadAvatar(c.Request.Context(), userID, file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"avatar_url": avatarURL})
}
```

### Change Password

```go
handler := func(c *gin.Context) {
    userID := c.GetString("user_id")
    var req ChangePasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    if req.NewPassword != req.ConfirmPassword {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Passwords don't match"})
        return
    }

    err := s.profileService.ChangePassword(c.Request.Context(), userID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Log activity
    s.profileService.LogActivity(c.Request.Context(), userID, "CHANGE_PASSWORD", "User changed password")

    c.JSON(http.StatusOK, SuccessResponse{Message: "Password changed successfully"})
}
```

## Error Handling

| Error | Status Code | Message |
|-------|------------|---------|
| Profile not found | 404 | "Profile not found" |
| Invalid email format | 400 | "Invalid email format" |
| Email already exists | 409 | "Email already registered" |
| Invalid file type | 400 | "Only JPG and PNG images allowed" |
| File too large | 400 | "File size exceeds 5MB limit" |
| Password mismatch | 400 | "Passwords don't match" |
| Invalid current password | 401 | "Current password is incorrect" |
| Weak password | 400 | "Password must be at least 8 characters" |
| Invalid preferences | 400 | "Invalid preference values" |

## Performance Optimization

### Database Indexes
- Index on user_id for quick lookups
- Index on email for duplicate checking
- Index on activity_type for filtering logs
- Index on created_at in activity logs

### Caching Strategy
- Cache profile data (1-hour TTL)
- Cache preferences (6-hour TTL)
- Cache privacy settings (6-hour TTL)
- Invalidate cache on updates

### File Storage
- Store avatars on S3/Cloud with CDN
- Resize avatars (original + thumbnail)
- Set cache headers for images
- Delete old avatars when updated

## Security Considerations

### Password Security
- Minimum 8 characters
- Require current password for change
- Hash passwords with bcrypt
- Implement rate limiting on attempts

### Email/Phone Changes
- Require verification via OTP/link
- Send notification to old email
- Log the change in activity
- Allow reverting recent changes

### Data Protection
- Encrypt sensitive fields at rest
- Implement data retention policies
- GDPR compliance for data export
- Audit all profile access

### Privacy
- Respect privacy settings
- Don't share data without consent
- Implement privacy by design
- Regular privacy audits

## Testing Strategy

### Unit Tests
- Profile update validation
- Password strength checking
- Email validation
- File upload validation

### Integration Tests
- Complete profile setup workflow
- Privacy setting enforcement
- Activity logging
- Data export functionality

## Integration Points

### With Auth Module
- Profile linked to user account
- Email/phone changes go through auth

### With Wallet Module
- Payment preferences stored
- Default payment method

### With All Modules
- User identification across system
- Display user info in transactions

## Common Pitfalls

1. **Not validating profile completeness** - Mark is_completed only after full setup
2. **Missing privacy enforcement** - Respect privacy settings when returning user data
3. **Not logging activities** - Critical for security and compliance
4. **Avatar file validation** - Check size, type, and scan for malware
5. **Not verifying email/phone changes** - Prevents account takeover
6. **Weak password validation** - Enforce strong requirements

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
