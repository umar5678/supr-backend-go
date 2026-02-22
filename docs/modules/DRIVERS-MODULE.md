# Drivers Module Development Guide

## Overview

The Drivers Module manages driver-specific functionality including profile creation and updates, document verification, availability management, earnings tracking, and performance metrics.

## Module Structure

```
drivers/
├── handler.go         # HTTP request handlers
├── service.go         # Business logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. Driver Profile Management - Create and update driver profiles
2. Document Verification - Handle license, insurance, and background checks
3. Availability Management - Track driver online/offline status
4. Earnings Tracking - Monitor driver earnings and payments
5. Driver Ratings - Maintain driver ratings and reviews
6. Account Restrictions - Manage driver suspensions and restrictions
7. Performance Metrics - Track driver KPIs

## Architecture

### Handler Layer (handler.go)

Manages HTTP endpoints for driver operations.

Key methods:

```
GetProfile(c *gin.Context)                    // GET /drivers/profile
UpdateProfile(c *gin.Context)                 // PUT /drivers/profile
UploadDocuments(c *gin.Context)               // POST /drivers/documents
GetDocuments(c *gin.Context)                  // GET /drivers/documents
VerifyDocuments(c *gin.Context)               // POST /drivers/documents/verify (admin)
SetAvailability(c *gin.Context)               // PUT /drivers/availability
GetEarnings(c *gin.Context)                   // GET /drivers/earnings
GetRating(c *gin.Context)                     // GET /drivers/rating
RequestWithdrawal(c *gin.Context)             // POST /drivers/withdrawal
GetRestrictions(c *gin.Context)               // GET /drivers/restrictions
```

Request flow:
1. Extract driver ID from JWT token
2. Validate request data
3. Call service method
4. Return response

### Service Layer (service.go)

Contains driver business logic.

Key interface methods:

```
GetProfile(ctx context.Context, driverID string) (*DriverProfileResponse, error)
UpdateProfile(ctx context.Context, driverID string, updates map[string]interface{}) error
UploadDocuments(ctx context.Context, driverID string, documents []DocumentFile) error
VerifyDocuments(ctx context.Context, driverID string) error
SetAvailability(ctx context.Context, driverID string, isAvailable bool) error
GetEarnings(ctx context.Context, driverID string, period string) (*EarningsResponse, error)
GetRating(ctx context.Context, driverID string) (*RatingResponse, error)
RequestWithdrawal(ctx context.Context, driverID string, amount float64) error
GetRestrictions(ctx context.Context, driverID string) (*RestrictionsResponse, error)
ApplyRestriction(ctx context.Context, driverID, reason string) error
```

Logic flow:
1. Validate input data
2. Check driver exists
3. Apply business rules
4. Update database records
5. Trigger related actions (notifications, webhooks)
6. Log operations

### Repository Layer (repository.go)

Handles database operations for drivers.

Key interface methods:

```
FindByID(ctx context.Context, driverID string) (*models.DriverProfile, error)
Create(ctx context.Context, profile *models.DriverProfile) error
Update(ctx context.Context, driverID string, profile *models.DriverProfile) error
SaveDocuments(ctx context.Context, driverID string, documents []models.Document) error
GetDocuments(ctx context.Context, driverID string) ([]models.Document, error)
UpdateDocumentStatus(ctx context.Context, documentID, status string) error
UpdateAvailability(ctx context.Context, driverID string, isAvailable bool) error
GetEarnings(ctx context.Context, driverID string, fromDate, toDate time.Time) (float64, error)
GetRating(ctx context.Context, driverID string) (float64, int64, error)
AddRestriction(ctx context.Context, driverID, reason string) error
RemoveRestriction(ctx context.Context, driverID string) error
```

Database operations:
- Use transactions for critical updates
- Implement proper indexing
- Store document references securely

## Data Transfer Objects

### DriverProfileResponse

```go
type DriverProfileResponse struct {
    ID                string            `json:"id"`
    UserID            string            `json:"user_id"`
    FirstName         string            `json:"first_name"`
    LastName          string            `json:"last_name"`
    Email             string            `json:"email"`
    Phone             string            `json:"phone"`
    ProfileImage      string            `json:"profile_image,omitempty"`
    Status            string            `json:"status"`
    IsAvailable       bool              `json:"is_available"`
    TotalRides        int64             `json:"total_rides"`
    AverageRating     float64           `json:"average_rating"`
    ResponseRate      float64           `json:"response_rate"`
    AcceptanceRate    float64           `json:"acceptance_rate"`
    CancellationRate  float64           `json:"cancellation_rate"`
    TotalEarnings     float64           `json:"total_earnings"`
    Documents         []DocumentInfo    `json:"documents"`
    Restrictions      []RestrictionInfo `json:"restrictions,omitempty"`
    CreatedAt         time.Time         `json:"created_at"`
    UpdatedAt         time.Time         `json:"updated_at"`
}

type DocumentInfo struct {
    ID              string    `json:"id"`
    Type            string    `json:"type"` // license, insurance, registration, background_check
    Status          string    `json:"status"` // pending, verified, rejected, expired
    DocumentURL     string    `json:"document_url,omitempty"`
    ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
    VerifiedAt      *time.Time `json:"verified_at,omitempty"`
    RejectionReason string    `json:"rejection_reason,omitempty"`
}
```

### UpdateProfileRequest

```go
type UpdateProfileRequest struct {
    FirstName   string `json:"first_name,omitempty"`
    LastName    string `json:"last_name,omitempty"`
    Email       string `json:"email,omitempty"`
    PhoneNumber string `json:"phone_number,omitempty"`
    ProfileImage string `json:"profile_image,omitempty"` // Base64 or URL
}
```

### DocumentFile

```go
type DocumentFile struct {
    Type         string `json:"type" binding:"required"` // license, insurance, registration, background_check
    FileData     []byte `json:"file_data" binding:"required"`
    FileName     string `json:"file_name"`
    ExpiryDate   *time.Time `json:"expiry_date,omitempty"`
    DocumentNumber string `json:"document_number,omitempty"`
}
```

### EarningsResponse

```go
type EarningsResponse struct {
    TotalEarnings     float64            `json:"total_earnings"`
    TotalRides        int64              `json:"total_rides"`
    AveragePerRide    float64            `json:"average_per_ride"`
    Period            string             `json:"period"` // daily, weekly, monthly
    Breakdown         []EarningBreakdown `json:"breakdown"`
    PendingWithdrawal float64            `json:"pending_withdrawal"`
    SettledAmount     float64            `json:"settled_amount"`
}

type EarningBreakdown struct {
    Date       string  `json:"date"`
    Amount     float64 `json:"amount"`
    RideCount  int64   `json:"ride_count"`
}
```

## Driver Status Types

```
PENDING_VERIFICATION - Awaiting document verification
ACTIVE - Fully verified and can accept rides
OFFLINE - Not available for rides
ON_RIDE - Currently on a ride
SUSPENDED - Account suspended by admin
RESTRICTED - Account has restrictions
INACTIVE - Inactive due to inactivity
```

## Document Requirements

### Essential Documents

1. Driver License
   - Valid and unexpired
   - Verification required
   - Expiry date tracked

2. Vehicle Registration
   - Current registration certificate
   - Matches vehicle on file
   - Verification required

3. Insurance Certificate
   - Valid commercial insurance
   - Coverage verified
   - Expiry date tracked

4. Background Check
   - Government approved check
   - No criminal record
   - Verified once

### Additional Documents

1. Bank Account Details
   - For payment settlement
   - Account verification
   - Account holder verification

2. Tax Information
   - PAN or TIN
   - GST registration (if applicable)
   - Address proof

3. Health Certificate
   - Optional but recommended
   - Regular renewal

## Typical Use Cases

### 1. Driver Registration

Request:
```
POST /drivers/register
{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "phone": "+1234567890"
}
```

Flow:
1. Create driver profile linked to user account
2. Initialize empty document list
3. Set status to PENDING_VERIFICATION
4. Return profile information
5. Send verification document upload instructions

### 2. Upload Documents

Request:
```
POST /drivers/documents
{
    "documents": [
        {
            "type": "license",
            "file_data": "base64_encoded_file",
            "expiry_date": "2026-05-20",
            "document_number": "DL-123456789"
        }
    ]
}
```

Flow:
1. Validate document types
2. Store files securely (S3/Cloud)
3. Create document records in database
4. Set document status to pending
5. Queue for admin verification
6. Return confirmation

### 3. Admin Verify Documents

Request:
```
POST /drivers/{driverID}/documents/verify
{
    "document_id": "doc-123",
    "status": "verified"
}
```

Flow:
1. Update document status
2. Check if all required documents verified
3. If all verified, update driver status to ACTIVE
4. Notify driver of approval
5. Create onboarding completion record

### 4. Get Driver Earnings

Request:
```
GET /drivers/earnings?period=monthly&month=2024-02
```

Response:
```json
{
    "total_earnings": 2500.00,
    "total_rides": 120,
    "average_per_ride": 20.83,
    "period": "monthly",
    "breakdown": [
        {
            "date": "2024-02-01",
            "amount": 85.50,
            "ride_count": 4
        }
    ],
    "pending_withdrawal": 0,
    "settled_amount": 2500.00
}
```

Flow:
1. Find driver by ID
2. Get earnings for specified period
3. Calculate average per ride
4. Aggregate by date
5. Get withdrawal status
6. Return comprehensive earnings data

### 5. Request Withdrawal

Request:
```
POST /drivers/withdrawal
{
    "amount": 1000.00,
    "bank_account_id": "bank-456"
}
```

Flow:
1. Verify driver has sufficient balance
2. Verify bank account is registered and verified
3. Check withdrawal limit (min: 100, max: 10000)
4. Create withdrawal request with status PENDING
5. Queue for batch processing
6. Send confirmation to driver
7. Process bank transfer in daily settlement

### 6. Set Availability

Request:
```
PUT /drivers/availability
{
    "is_available": true
}
```

Flow:
1. Update driver availability status
2. Log status change
3. Broadcast availability to matching system
4. Send confirmation
5. Track last availability change time

### 7. Get Driver Rating

Request:
```
GET /drivers/rating
```

Response:
```json
{
    "average_rating": 4.85,
    "total_ratings": 342,
    "distribution": {
        "5": 280,
        "4": 45,
        "3": 12,
        "2": 3,
        "1": 2
    },
    "recent_ratings": [
        {
            "rating": 5,
            "comment": "Great driver!",
            "date": "2024-02-20"
        }
    ]
}
```

Flow:
1. Find driver by ID
2. Calculate average rating from ratings table
3. Count total number of ratings
4. Get distribution by star
5. Get recent ratings
6. Return comprehensive rating data

## Performance Metrics

Drivers can view key performance metrics:

1. Response Rate - % of ride requests accepted
2. Acceptance Rate - % of offers accepted
3. Cancellation Rate - % of rides cancelled
4. Average Rating - Star rating from riders
5. On-time Rate - % of on-time arrivals
6. Completion Rate - % of accepted rides completed

## Restrictions and Suspensions

### Automatic Restrictions

Applied based on metrics:

- Cancellation Rate > 30% - Warning, then restricted
- Average Rating < 2.5 - Restricted
- Non-responsiveness > 3 days - Restricted
- Documents expired - Automatic suspension
- Payment defaults - Restricted

### Manual Restrictions

Applied by admin:

- Fraud detection
- User complaints
- Policy violations
- Criminal issues
- Safety concerns

## Error Handling

Common error scenarios:

1. Driver Not Found
   - Response: 404 Not Found
   - Action: Check driver ID

2. Document Verification Failed
   - Response: 400 Bad Request
   - Message: "Document rejected: Poor quality"

3. Insufficient Earnings for Withdrawal
   - Response: 400 Bad Request
   - Message: "Insufficient balance. Available: 50, Requested: 100"

4. Expired Documents
   - Response: 409 Conflict
   - Action: Notify driver to renew documents

## Testing Strategy

### Unit Tests (Service Layer)

```go
Test_GetProfile()
Test_UpdateProfile_Success()
Test_VerifyDocuments_Success()
Test_CalculateEarnings()
Test_GetRating()
Test_SetAvailability()
```

### Integration Tests (Repository Layer)

```go
Test_CreateDriver()
Test_SaveDocuments()
Test_UpdateAvailability_Persistence()
```

### End-to-End Tests (Handler Layer)

```go
Test_DriverRegistration_FullFlow()
Test_DocumentUpload_FullFlow()
Test_EarningsRetrieval()
```

## Database Schema

### Driver Profiles Table

```sql
CREATE TABLE driver_profiles (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    phone VARCHAR(20),
    profile_image VARCHAR(500),
    status VARCHAR(50),
    is_available BOOLEAN DEFAULT false,
    total_rides INT DEFAULT 0,
    average_rating DECIMAL(3, 2) DEFAULT 0,
    response_rate DECIMAL(5, 2) DEFAULT 0,
    acceptance_rate DECIMAL(5, 2) DEFAULT 0,
    cancellation_rate DECIMAL(5, 2) DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Driver Documents Table

```sql
CREATE TABLE driver_documents (
    id VARCHAR(36) PRIMARY KEY,
    driver_id VARCHAR(36) NOT NULL,
    type VARCHAR(50),
    status VARCHAR(50),
    document_url VARCHAR(500),
    document_number VARCHAR(100),
    expiry_date DATE,
    verified_at TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (driver_id) REFERENCES driver_profiles(id)
);
```

## Integration Points

1. Auth Module - For driver user creation
2. Rides Module - For ride assignment
3. Wallet Module - For earnings and settlements
4. Ratings Module - For driver ratings
5. Vehicles Module - For vehicle information
6. Messages Module - For notifications

## Performance Optimization

1. Cache driver profiles with Redis
2. Index on status and availability columns
3. Use prepared statements for document queries
4. Batch document verification processing
5. Cache driver ratings

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture
- See VEHICLES-MODULE.md for vehicle management
- See WALLET-MODULE.md for earnings tracking
- See RATINGS-MODULE.md for rating management

## Common Pitfalls

1. Not validating document expiry dates
2. Race conditions in availability updates
3. Insufficient logging of document verifications
4. Not implementing proper document storage
5. Missing security checks on sensitive operations
6. Not tracking document changes

## Future Enhancements

1. AI-based document verification
2. Biometric driver verification
3. Real-time vehicle location tracking
4. Predictive maintenance alerts
5. Driver training modules
6. Performance-based incentives
7. Insurance integration
8. Multi-language support
