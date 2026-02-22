# Riders Module Documentation

## Overview

The Riders Module manages rider-specific functionality including user profiles, saved addresses, ride history, preferences, and rider statistics. It handles rider onboarding, profile management, and rider-specific features like saved locations and ride preferences.

## Key Responsibilities

- Rider profile creation and management
- Saved addresses (home, work, frequent locations)
- Ride history and statistics
- Rider preferences and settings
- Emergency contacts management
- Rider rating and feedback
- Rider verification (phone, email, identity)
- Payment method management
- Referral program management

## Architecture

### Handler Layer (`riders/handler.go`)

Handles HTTP requests related to rider operations.

**Key Endpoints:**

```go
GET /api/v1/riders/profile              // Get rider profile
PUT /api/v1/riders/profile              // Update rider profile
GET /api/v1/riders/addresses            // Get saved addresses
POST /api/v1/riders/addresses           // Add saved address
PUT /api/v1/riders/addresses/:id        // Update saved address
DELETE /api/v1/riders/addresses/:id     // Delete saved address
GET /api/v1/riders/preferences          // Get rider preferences
PUT /api/v1/riders/preferences          // Update preferences
GET /api/v1/riders/history              // Get ride history
GET /api/v1/riders/stats                // Get rider statistics
POST /api/v1/riders/payment-methods     // Add payment method
GET /api/v1/riders/payment-methods      // List payment methods
DELETE /api/v1/riders/payment-methods/:id // Remove payment method
```

### Service Layer (`riders/service.go`)

Contains business logic for rider operations.

**Key Methods:**

```go
func (s *RiderService) GetProfile(ctx context.Context, riderID string) (*RiderProfileResponse, error)
func (s *RiderService) UpdateProfile(ctx context.Context, riderID string, req *UpdateProfileRequest) (*RiderProfileResponse, error)
func (s *RiderService) GetAddresses(ctx context.Context, riderID string) ([]*SavedAddressResponse, error)
func (s *RiderService) AddAddress(ctx context.Context, riderID string, req *AddAddressRequest) (*SavedAddressResponse, error)
func (s *RiderService) UpdateAddress(ctx context.Context, riderID, addressID string, req *UpdateAddressRequest) (*SavedAddressResponse, error)
func (s *RiderService) DeleteAddress(ctx context.Context, riderID, addressID string) error
func (s *RiderService) GetPreferences(ctx context.Context, riderID string) (*RiderPreferencesResponse, error)
func (s *RiderService) UpdatePreferences(ctx context.Context, riderID string, req *UpdatePreferencesRequest) (*RiderPreferencesResponse, error)
func (s *RiderService) GetRideHistory(ctx context.Context, riderID string, filters *RideHistoryFilters) ([]*RideHistoryResponse, error)
func (s *RiderService) GetStatistics(ctx context.Context, riderID string) (*RiderStatsResponse, error)
func (s *RiderService) AddPaymentMethod(ctx context.Context, riderID string, req *AddPaymentMethodRequest) (*PaymentMethodResponse, error)
```

### Repository Layer (`riders/repository.go`)

Manages database operations for riders.

**Key Methods:**

```go
func (r *RiderRepository) Create(ctx context.Context, rider *Rider) error
func (r *RiderRepository) GetByID(ctx context.Context, riderID string) (*Rider, error)
func (r *RiderRepository) Update(ctx context.Context, riderID string, updates map[string]interface{}) error
func (r *RiderRepository) AddAddress(ctx context.Context, address *SavedAddress) error
func (r *RiderRepository) GetAddresses(ctx context.Context, riderID string) ([]*SavedAddress, error)
func (r *RiderRepository) UpdateAddress(ctx context.Context, addressID string, updates map[string]interface{}) error
func (r *RiderRepository) DeleteAddress(ctx context.Context, addressID string) error
func (r *RiderRepository) GetRideHistory(ctx context.Context, riderID string, filters *RideHistoryFilters) ([]*RideHistory, error)
```

## Data Models

### Rider

```go
type Rider struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`                     // Links to user in auth
    FirstName           string     `db:"first_name" json:"first_name"`
    LastName            string     `db:"last_name" json:"last_name"`
    Email               string     `db:"email" json:"email"`
    Phone               string     `db:"phone" json:"phone"`
    Avatar              *string    `db:"avatar" json:"avatar"`                       // Profile photo URL
    DateOfBirth         *time.Time `db:"date_of_birth" json:"date_of_birth"`
    Gender              *string    `db:"gender" json:"gender"`                       // Male, Female, Other
    
    // Verification
    EmailVerified       bool       `db:"email_verified" json:"email_verified"`
    PhoneVerified       bool       `db:"phone_verified" json:"phone_verified"`
    IdentityVerified    bool       `db:"identity_verified" json:"identity_verified"`
    
    // Address
    HomeAddress         *string    `db:"home_address" json:"home_address"`
    WorkAddress         *string    `db:"work_address" json:"work_address"`
    
    // Status
    Status              string     `db:"status" json:"status"`                       // ACTIVE, INACTIVE, SUSPENDED
    IsActive            bool       `db:"is_active" json:"is_active"`
    
    // Statistics
    TotalRides          int        `db:"total_rides" json:"total_rides"`
    AverageRating       float64    `db:"average_rating" json:"average_rating"`
    TotalSpent          float64    `db:"total_spent" json:"total_spent"`
    LastRideAt          *time.Time `db:"last_ride_at" json:"last_ride_at"`
    
    // Referral
    ReferralCode        string     `db:"referral_code" json:"referral_code"`        // Unique code for referrals
    ReferredBy          *string    `db:"referred_by" json:"referred_by"`            // Referrer's rider ID
    ReferralsCount      int        `db:"referrals_count" json:"referrals_count"`
    
    // Preferences
    Language            string     `db:"language" json:"language"`
    PreferredPayment    *string    `db:"preferred_payment" json:"preferred_payment"`
    ReceiveNotifications bool      `db:"receive_notifications" json:"receive_notifications"`
    
    // Metadata
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
    LastLoginAt         *time.Time `db:"last_login_at" json:"last_login_at"`
}
```

### SavedAddress

```go
type SavedAddress struct {
    ID                  string     `db:"id" json:"id"`
    RiderID             string     `db:"rider_id" json:"rider_id"`
    AddressType         string     `db:"address_type" json:"address_type"`           // HOME, WORK, FREQUENT, OTHER
    Label               string     `db:"label" json:"label"`                         // Custom label (e.g., "Mom's House")
    FullAddress         string     `db:"full_address" json:"full_address"`
    Latitude            float64    `db:"latitude" json:"latitude"`
    Longitude           float64    `db:"longitude" json:"longitude"`
    ZipCode             *string    `db:"zip_code" json:"zip_code"`
    City                string     `db:"city" json:"city"`
    State               *string    `db:"state" json:"state"`
    Country             string     `db:"country" json:"country"`
    IsDefault           bool       `db:"is_default" json:"is_default"`              // Default pickup/dropoff
    IsFrequent          bool       `db:"is_frequent" json:"is_frequent"`            // Frequently used
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### RiderPreferences

```go
type RiderPreferences struct {
    ID                      string     `db:"id" json:"id"`
    RiderID                 string     `db:"rider_id" json:"rider_id"`
    PreferredRideType       string     `db:"preferred_ride_type" json:"preferred_ride_type"`       // Economy, Premium, XL
    ShareWithEmergency      bool       `db:"share_with_emergency" json:"share_with_emergency"`
    AllowCarpooling         bool       `db:"allow_carpooling" json:"allow_carpooling"`
    PreferSilentRides       bool       `db:"prefer_silent_rides" json:"prefer_silent_rides"`
    PreferMaleFemaleDriver  *string    `db:"prefer_male_female_driver" json:"prefer_male_female_driver"` // Any, Male, Female
    ReceivePromotions       bool       `db:"receive_promotions" json:"receive_promotions"`
    ReceiveSafetyAlerts     bool       `db:"receive_safety_alerts" json:"receive_safety_alerts"`
    AllowDataCollection     bool       `db:"allow_data_collection" json:"allow_data_collection"`
    NightModeEnabled        bool       `db:"night_mode_enabled" json:"night_mode_enabled"`
    Language                string     `db:"language" json:"language"`                               // en, es, fr, etc.
    CreatedAt               time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}
```

### RideHistory

```go
type RideHistory struct {
    ID                  string     `db:"id" json:"id"`
    RiderID             string     `db:"rider_id" json:"rider_id"`
    RideID              string     `db:"ride_id" json:"ride_id"`
    DriverID            string     `db:"driver_id" json:"driver_id"`
    PickupLocation      string     `db:"pickup_location" json:"pickup_location"`
    DropoffLocation     string     `db:"dropoff_location" json:"dropoff_location"`
    RideType            string     `db:"ride_type" json:"ride_type"`                 // Economy, Premium, XL
    StartTime           time.Time  `db:"start_time" json:"start_time"`
    EndTime             *time.Time `db:"end_time" json:"end_time"`
    Distance            float64    `db:"distance" json:"distance"`                   // in km
    Duration            int        `db:"duration" json:"duration"`                   // in seconds
    Fare                float64    `db:"fare" json:"fare"`
    Tip                 float64    `db:"tip" json:"tip"`
    PaymentMethod       string     `db:"payment_method" json:"payment_method"`       // Card, Wallet, Cash
    DriverRating        *int       `db:"driver_rating" json:"driver_rating"`
    DriverReview        *string    `db:"driver_review" json:"driver_review"`
    Status              string     `db:"status" json:"status"`                       // COMPLETED, CANCELLED
    CancellationReason  *string    `db:"cancellation_reason" json:"cancellation_reason"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}
```

## DTOs (Data Transfer Objects)

### RiderProfileResponse

```go
type RiderProfileResponse struct {
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
    IdentityVerified    bool       `json:"identity_verified"`
    TotalRides          int        `json:"total_rides"`
    AverageRating       float64    `json:"average_rating"`
    TotalSpent          float64    `json:"total_spent"`
    ReferralCode        string     `json:"referral_code"`
    ReferralsCount      int        `json:"referrals_count"`
    Status              string     `json:"status"`
    CreatedAt           time.Time  `json:"created_at"`
}
```

### UpdateProfileRequest

```go
type UpdateProfileRequest struct {
    FirstName           *string    `json:"first_name"`
    LastName            *string    `json:"last_name"`
    Avatar              *string    `json:"avatar"`
    DateOfBirth         *time.Time `json:"date_of_birth"`
    Gender              *string    `json:"gender"`
    HomeAddress         *string    `json:"home_address"`
    WorkAddress         *string    `json:"work_address"`
}
```

### SavedAddressResponse

```go
type SavedAddressResponse struct {
    ID                  string     `json:"id"`
    AddressType         string     `json:"address_type"`
    Label               string     `json:"label"`
    FullAddress         string     `json:"full_address"`
    Latitude            float64    `json:"latitude"`
    Longitude           float64    `json:"longitude"`
    City                string     `json:"city"`
    IsDefault           bool       `json:"is_default"`
    IsFrequent          bool       `json:"is_frequent"`
}
```

### RiderStatsResponse

```go
type RiderStatsResponse struct {
    TotalRides          int        `json:"total_rides"`
    TotalDistance       float64    `json:"total_distance"`      // in km
    TotalSpent          float64    `json:"total_spent"`
    AverageRideTime     int        `json:"average_ride_time"`   // in seconds
    AverageRating       float64    `json:"average_rating"`
    FavoriteRideType    string     `json:"favorite_ride_type"`
    FavoriteDriver      *string    `json:"favorite_driver"`     // Most frequent driver
    FavoriteCities      []string   `json:"favorite_cities"`
    MostUsedPayment     string     `json:"most_used_payment"`
    MonthlySpending     float64    `json:"monthly_spending"`
    SavedMoney          float64    `json:"saved_money"`         // From promotions
}
```

## Database Schema

### riders table

```sql
CREATE TABLE riders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    avatar TEXT,
    date_of_birth DATE,
    gender VARCHAR(20),
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    identity_verified BOOLEAN DEFAULT FALSE,
    home_address TEXT,
    work_address TEXT,
    status VARCHAR(50) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    is_active BOOLEAN DEFAULT TRUE,
    total_rides INTEGER DEFAULT 0,
    average_rating DECIMAL(3, 2) DEFAULT 0,
    total_spent DECIMAL(10, 2) DEFAULT 0,
    last_ride_at TIMESTAMP,
    referral_code VARCHAR(20) UNIQUE NOT NULL,
    referred_by UUID REFERENCES riders(id),
    referrals_count INTEGER DEFAULT 0,
    language VARCHAR(10) DEFAULT 'en',
    preferred_payment VARCHAR(50),
    receive_notifications BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE INDEX idx_riders_user_id ON riders(user_id);
CREATE INDEX idx_riders_email ON riders(email);
CREATE INDEX idx_riders_phone ON riders(phone);
CREATE INDEX idx_riders_referral_code ON riders(referral_code);
CREATE INDEX idx_riders_status ON riders(status);
CREATE INDEX idx_riders_created_at ON riders(created_at);
```

### saved_addresses table

```sql
CREATE TABLE saved_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rider_id UUID NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    address_type VARCHAR(50) NOT NULL CHECK (address_type IN ('HOME', 'WORK', 'FREQUENT', 'OTHER')),
    label VARCHAR(100),
    full_address TEXT NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    zip_code VARCHAR(20),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    country VARCHAR(100) NOT NULL DEFAULT 'US',
    is_default BOOLEAN DEFAULT FALSE,
    is_frequent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rider_id, address_type, label)
);

CREATE INDEX idx_saved_addresses_rider_id ON saved_addresses(rider_id);
CREATE INDEX idx_saved_addresses_address_type ON saved_addresses(address_type);
```

### rider_preferences table

```sql
CREATE TABLE rider_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rider_id UUID NOT NULL UNIQUE REFERENCES riders(id) ON DELETE CASCADE,
    preferred_ride_type VARCHAR(50) CHECK (preferred_ride_type IN ('ECONOMY', 'PREMIUM', 'XL')),
    share_with_emergency BOOLEAN DEFAULT TRUE,
    allow_carpooling BOOLEAN DEFAULT TRUE,
    prefer_silent_rides BOOLEAN DEFAULT FALSE,
    prefer_male_female_driver VARCHAR(20),
    receive_promotions BOOLEAN DEFAULT TRUE,
    receive_safety_alerts BOOLEAN DEFAULT TRUE,
    allow_data_collection BOOLEAN DEFAULT TRUE,
    night_mode_enabled BOOLEAN DEFAULT FALSE,
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rider_preferences_rider_id ON rider_preferences(rider_id);
```

### ride_history table

```sql
CREATE TABLE ride_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rider_id UUID NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    ride_id UUID NOT NULL REFERENCES rides(id),
    driver_id UUID NOT NULL,
    pickup_location TEXT NOT NULL,
    dropoff_location TEXT NOT NULL,
    ride_type VARCHAR(50) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    distance DECIMAL(10, 2),
    duration INTEGER,
    fare DECIMAL(10, 2),
    tip DECIMAL(10, 2) DEFAULT 0,
    payment_method VARCHAR(50),
    driver_rating INTEGER CHECK (driver_rating >= 1 AND driver_rating <= 5),
    driver_review TEXT,
    status VARCHAR(50) CHECK (status IN ('COMPLETED', 'CANCELLED')),
    cancellation_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rider_id, ride_id)
);

CREATE INDEX idx_ride_history_rider_id ON ride_history(rider_id);
CREATE INDEX idx_ride_history_start_time ON ride_history(start_time);
CREATE INDEX idx_ride_history_status ON ride_history(status);
```

## Use Cases

### Use Case 1: Rider Profile Setup

```
1. Rider signup via Auth module
2. Rider profile created with basic info
3. Email verification email sent
4. Phone verification SMS sent
5. Rider adds saved addresses (home, work)
6. Rider sets preferences
7. Rider adds payment method
8. Profile setup complete
```

### Use Case 2: Saved Addresses and Quick Booking

```
1. Rider saves "Home" address with coordinates
2. Rider saves "Work" address
3. When requesting ride, app suggests saved addresses
4. Rider can tap "Home to Work" for quick booking
5. Quick booking calculates fare instantly
6. Ride proceeds normally
```

### Use Case 3: Ride History and Statistics

```
1. After each completed ride, entry added to ride_history
2. Rider can view past rides with details
3. Statistics automatically calculated (total rides, spending, etc.)
4. Rider can re-book from recent locations
5. Rider can see trends (favorite areas, average cost)
6. Export ride history for tax/expense purposes
```

### Use Case 4: Referral Program

```
1. Each rider has unique referral_code
2. Rider shares code with friends
3. Friend enters code during signup
4. Referral tracked via referred_by field
5. Both get bonus credit when friend completes first ride
6. Rider accumulates referrals_count
7. Leaderboard shows top referrers
```

## Common Operations

### Get Rider Profile

```go
handler := func(c *gin.Context) {
    riderID := c.GetString("user_id")
    
    profile, err := s.riderService.GetProfile(c.Request.Context(), riderID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, profile)
}
```

### Add Saved Address

```go
handler := func(c *gin.Context) {
    riderID := c.GetString("user_id")
    var req AddAddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    address, err := s.riderService.AddAddress(c.Request.Context(), riderID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusCreated, address)
}
```

### Get Ride History

```go
handler := func(c *gin.Context) {
    riderID := c.GetString("user_id")
    filters := &RideHistoryFilters{
        Page: getQueryInt(c, "page", 1),
        Limit: getQueryInt(c, "limit", 10),
        Status: c.Query("status"),
    }

    history, err := s.riderService.GetRideHistory(c.Request.Context(), riderID, filters)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, history)
}
```

### Get Rider Statistics

```go
handler := func(c *gin.Context) {
    riderID := c.GetString("user_id")
    
    stats, err := s.riderService.GetStatistics(c.Request.Context(), riderID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, stats)
}
```

## Error Handling

| Error | Status Code | Message |
|-------|------------|---------|
| Rider not found | 404 | "Rider profile not found" |
| Invalid address | 400 | "Invalid latitude/longitude coordinates" |
| Duplicate address | 409 | "Address with this label already exists" |
| Unauthorized | 403 | "You don't have permission to access this" |
| Invalid referral code | 400 | "Invalid referral code provided" |
| Self-referral | 409 | "Cannot refer yourself" |
| Already referred | 409 | "Rider already referred by another user" |

## Performance Optimization

### Database Indexes
- Index on `user_id` for quick lookups
- Index on `email` and `phone` for verification
- Index on `referral_code` for code lookups
- Index on rider_id in saved_addresses for quick address retrieval
- Index on `created_at` in ride_history for timeline queries

### Caching Strategy
- Cache rider profile (1-hour TTL)
- Cache saved addresses (12-hour TTL)
- Cache ride statistics (real-time updates)
- Cache preferences (1-day TTL)

## Security Considerations

### Data Protection
- Encrypt phone numbers at rest
- Hash referral codes
- Implement data retention policies
- Anonymize ride history after 2 years

### Authorization
- Only rider can view own profile
- Only rider can modify own addresses
- Admins can view aggregate statistics
- Payment methods require additional verification

### Privacy
- GDPR compliance for data collection
- Allow users to opt-out of data collection
- Implement right to be forgotten
- Audit all access to rider data

## Testing Strategy

### Unit Tests
- Profile creation and updates
- Address validation and storage
- Preference management
- Statistics calculation
- Referral code generation

### Integration Tests
- Complete rider registration flow
- Address management workflow
- Ride history tracking
- Statistics aggregation

## Integration Points

### With Auth Module
- Rider links to user account
- Phone/email verification

### With Rides Module
- Rider history pulls from completed rides
- Statistics calculated from ride data
- Addresses used as pickup/dropoff

### With Wallet Module
- Preferred payment method stored
- Spending statistics tracked

## Common Pitfalls

1. **Not handling address coordinates** - Must have valid lat/long
2. **Referral loop issues** - Prevent circular referrals
3. **Stale statistics** - Update in real-time or with cron job
4. **Not validating address type** - Only allow valid types
5. **Missing address defaults** - Should have default pickup address
6. **Not tracking referrals properly** - Ensure referred_by is set once

## Typical Examples

### Get Profile

```
GET /api/v1/riders/profile

Response: 200 OK
{
    "id": "rider-uuid",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "total_rides": 45,
    "average_rating": 4.8,
    "total_spent": 1250.50,
    "referral_code": "JOHN12345",
    "referrals_count": 5
}
```

### Add Saved Address

```
POST /api/v1/riders/addresses
{
    "address_type": "HOME",
    "label": "Home",
    "full_address": "123 Main St, City, State 12345",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "city": "New York",
    "is_default": true
}

Response: 201 Created
{
    "id": "address-uuid",
    "address_type": "HOME",
    "label": "Home",
    "full_address": "123 Main St, City, State 12345",
    "is_default": true
}
```

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
