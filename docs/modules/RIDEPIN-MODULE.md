# Ride PIN Module Documentation

## Overview

The Ride PIN Module manages ride verification through PIN codes. It provides an additional security layer by requiring drivers and riders to verify each other's identity with a PIN before starting a ride.

## Key Responsibilities

- PIN code generation and validation
- PIN verification for driver-rider matching
- PIN reset and re-generation
- PIN expiry management
- Verification attempt tracking
- Security alerts on failed attempts
- Integration with ride lifecycle

## Architecture

### Handler Layer (`ridepin/handler.go`)

Handles HTTP requests related to ride PIN operations.

**Key Endpoints:**

```go
POST /api/v1/rides/:id/generate-pin        // Generate PIN for ride
GET /api/v1/rides/:id/pin-status           // Get PIN status
POST /api/v1/rides/:id/verify-pin          // Verify rider/driver PIN
POST /api/v1/rides/:id/resend-pin          // Resend PIN via SMS
PUT /api/v1/rides/:id/reset-pin            // Reset PIN if needed
GET /api/v1/rides/:id/verification-history // Get verification attempts
```

### Service Layer (`ridepin/service.go`)

Contains business logic for PIN operations.

**Key Methods:**

```go
func (s *RidePINService) GeneratePIN(ctx context.Context, rideID string) (*PINResponse, error)
func (s *RidePINService) VerifyPIN(ctx context.Context, rideID string, req *VerifyPINRequest) (*VerificationResult, error)
func (s *RidePINService) ValidatePIN(ctx context.Context, rideID string, pin string) (bool, error)
func (s *RidePINService) ResendPIN(ctx context.Context, rideID string) error
func (s *RidePINService) ResetPIN(ctx context.Context, rideID string) (*PINResponse, error)
func (s *RidePINService) GetPINStatus(ctx context.Context, rideID string) (*PINStatus, error)
func (s *RidePINService) RecordVerificationAttempt(ctx context.Context, rideID string, req *VerificationAttempt) error
```

### Repository Layer (`ridepin/repository.go`)

Manages database operations.

**Key Methods:**

```go
func (r *RidePINRepository) CreatePIN(ctx context.Context, pin *RidePin) error
func (r *RidePINRepository) GetPINByRideID(ctx context.Context, rideID string) (*RidePin, error)
func (r *RidePINRepository) UpdatePIN(ctx context.Context, rideID string, updates map[string]interface{}) error
func (r *RidePINRepository) DeletePIN(ctx context.Context, rideID string) error
func (r *RidePINRepository) RecordAttempt(ctx context.Context, attempt *VerificationAttempt) error
func (r *RidePINRepository) GetAttempts(ctx context.Context, rideID string) ([]*VerificationAttempt, error)
```

## Data Models

### RidePin

```go
type RidePin struct {
    ID                  string     `db:"id" json:"id"`
    RideID              string     `db:"ride_id" json:"ride_id"`
    DriverID            string     `db:"driver_id" json:"driver_id"`
    RiderID             string     `db:"rider_id" json:"rider_id"`
    
    // PIN Details
    PIN                 string     `db:"pin" json:"pin"`                             // 6-digit numeric code
    PINHash             string     `db:"pin_hash" json:"pin_hash"`                   // Hashed for security
    GeneratedAt         time.Time  `db:"generated_at" json:"generated_at"`
    ExpiryAt            time.Time  `db:"expiry_at" json:"expiry_at"`                 // Usually 10 minutes
    
    // Verification Status
    DriverVerified      bool       `db:"driver_verified" json:"driver_verified"`
    DriverVerifiedAt    *time.Time `db:"driver_verified_at" json:"driver_verified_at"`
    RiderVerified       bool       `db:"rider_verified" json:"rider_verified"`
    RiderVerifiedAt     *time.Time `db:"rider_verified_at" json:"rider_verified_at"`
    BothVerified        bool       `db:"both_verified" json:"both_verified"`         // Both verified ride can start
    BothVerifiedAt      *time.Time `db:"both_verified_at" json:"both_verified_at"`
    
    // Attempt Tracking
    DriverAttempts      int        `db:"driver_attempts" json:"driver_attempts"`     // Failed attempts
    RiderAttempts       int        `db:"rider_attempts" json:"rider_attempts"`
    MaxAttempts         int        `db:"max_attempts" json:"max_attempts"`           // Usually 3
    
    // Security
    IsBlocked           bool       `db:"is_blocked" json:"is_blocked"`               // Blocked after max attempts
    BlockedReason       *string    `db:"blocked_reason" json:"blocked_reason"`
    
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### VerificationAttempt

```go
type VerificationAttempt struct {
    ID                  string     `db:"id" json:"id"`
    RideID              string     `db:"ride_id" json:"ride_id"`
    AttemptedBy         string     `db:"attempted_by" json:"attempted_by"`          // Driver or Rider ID
    AttemptType         string     `db:"attempt_type" json:"attempt_type"`          // DRIVER, RIDER
    PINProvided         string     `db:"pin_provided" json:"pin_provided"`          // User entered PIN (hashed)
    IsCorrect           bool       `db:"is_correct" json:"is_correct"`
    Message             string     `db:"message" json:"message"`                    // "Correct PIN", "Wrong PIN", "Expired"
    IPAddress           string     `db:"ip_address" json:"ip_address"`
    UserAgent           string     `db:"user_agent" json:"user_agent"`
    Location            *string    `db:"location" json:"location"`                  // GPS coordinates
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}
```

## DTOs (Data Transfer Objects)

### PINResponse

```go
type PINResponse struct {
    RideID              string     `json:"ride_id"`
    PIN                 string     `json:"pin"`                                     // Only sent to recipient
    ExpiryTime          int        `json:"expiry_time"`                             // Seconds until expiry
    Message             string     `json:"message"`                                 // "PIN sent to rider"
}
```

### VerifyPINRequest

```go
type VerifyPINRequest struct {
    PIN                 string     `json:"pin" binding:"required,len=6"`
}
```

### VerificationResult

```go
type VerificationResult struct {
    Success             bool       `json:"success"`
    Message             string     `json:"message"`
    RemainingAttempts   int        `json:"remaining_attempts"`
    CanStartRide        bool       `json:"can_start_ride"`                          // Both verified?
}
```

### PINStatus

```go
type PINStatus struct {
    RideID              string     `json:"ride_id"`
    DriverVerified      bool       `json:"driver_verified"`
    RiderVerified       bool       `json:"rider_verified"`
    BothVerified        bool       `json:"both_verified"`
    PINExpired          bool       `json:"pin_expired"`
    DriverAttempts      int        `json:"driver_attempts"`
    RiderAttempts       int        `json:"rider_attempts"`
    IsBlocked           bool       `json:"is_blocked"`
}
```

## Database Schema

### ride_pins table

```sql
CREATE TABLE ride_pins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL UNIQUE REFERENCES rides(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL,
    rider_id UUID NOT NULL,
    pin VARCHAR(6) NOT NULL,
    pin_hash VARCHAR(255) NOT NULL,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_at TIMESTAMP NOT NULL,
    driver_verified BOOLEAN DEFAULT FALSE,
    driver_verified_at TIMESTAMP,
    rider_verified BOOLEAN DEFAULT FALSE,
    rider_verified_at TIMESTAMP,
    both_verified BOOLEAN DEFAULT FALSE,
    both_verified_at TIMESTAMP,
    driver_attempts INTEGER DEFAULT 0,
    rider_attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    is_blocked BOOLEAN DEFAULT FALSE,
    blocked_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ride_pins_ride_id ON ride_pins(ride_id);
CREATE INDEX idx_ride_pins_driver_id ON ride_pins(driver_id);
CREATE INDEX idx_ride_pins_rider_id ON ride_pins(rider_id);
CREATE INDEX idx_ride_pins_expiry_at ON ride_pins(expiry_at);
```

### verification_attempts table

```sql
CREATE TABLE verification_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    attempted_by UUID NOT NULL,
    attempt_type VARCHAR(50) CHECK (attempt_type IN ('DRIVER', 'RIDER')),
    pin_provided VARCHAR(255),
    is_correct BOOLEAN,
    message TEXT,
    ip_address INET,
    user_agent TEXT,
    location POINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_verification_attempts_ride_id ON verification_attempts(ride_id);
CREATE INDEX idx_verification_attempts_is_correct ON verification_attempts(is_correct);
```

## Use Cases

### Use Case 1: PIN Generation and Sharing

```
1. Ride accepted by driver
2. System generates 6-digit random PIN
3. PIN hashed and stored in database
4. PIN sent to rider via SMS and push notification
5. PIN also available in app (copy/share button)
6. PIN expires in 10 minutes
7. If not verified, system can auto-send reminder
```

### Use Case 2: PIN Verification by Rider

```
1. Driver arrives at pickup location
2. Driver asks rider for PIN
3. Rider reads PIN from notification or app
4. Rider provides PIN (verbally or app)
5. System verifies PIN against hash
6. If correct: Mark rider_verified = true
7. If wrong: Increment rider_attempts, show remaining attempts
8. After 3 failed: Block PIN, generate new one
```

### Use Case 3: Both Parties Verified - Ride Can Start

```
1. Driver verifies PIN from rider (or vice versa)
2. Both driver_verified and rider_verified are true
3. System sets both_verified = true
4. Driver can now start the ride
5. Ride status changes from ARRIVED to STARTED
6. Time tracking begins
```

### Use Case 4: Failed Verification Handling

```
1. User enters wrong PIN three times
2. System blocks PIN for that ride
3. User cannot try more times
4. System auto-generates new PIN
5. New PIN sent to original recipient
6. User can retry with new PIN
7. Or driver/rider can cancel and new ride created
```

## Common Operations

### Generate PIN

```go
handler := func(c *gin.Context) {
    rideID := c.Param("id")
    
    pinResponse, err := s.ridePINService.GeneratePIN(c.Request.Context(), rideID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Send PIN via SMS to rider
    go s.sendPINViaSMS(pinResponse.RiderID, pinResponse.PIN)
    
    // Send via push notification
    go s.sendPINViaPush(pinResponse.RiderID, pinResponse.PIN)

    c.JSON(http.StatusOK, pinResponse)
}
```

### Verify PIN

```go
handler := func(c *gin.Context) {
    rideID := c.Param("id")
    userID := c.GetString("user_id")
    
    var req VerifyPINRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    result, err := s.ridePINService.VerifyPIN(c.Request.Context(), rideID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Log verification attempt
    s.ridePINService.RecordVerificationAttempt(c.Request.Context(), rideID, &VerificationAttempt{
        RideID: rideID,
        AttemptedBy: userID,
        IsCorrect: result.Success,
    })

    c.JSON(http.StatusOK, result)
}
```

## Error Handling

| Error | Status Code | Message |
|-------|------------|---------|
| Ride not found | 404 | "Ride not found" |
| PIN not generated | 400 | "PIN not yet generated for this ride" |
| PIN expired | 410 | "PIN has expired, requesting new one" |
| Wrong PIN | 401 | "Incorrect PIN, {X} attempts remaining" |
| Max attempts exceeded | 429 | "Too many failed attempts, PIN blocked" |
| PIN already verified | 409 | "PIN already verified, cannot re-verify" |
| Both parties not verified | 423 | "Both driver and rider must verify PIN" |

## Performance Optimization

### PIN Generation
- Use cryptographically secure random generator
- Cache PIN in Redis with TTL
- Hash PIN for storage (bcrypt)

### Verification
- Use constant-time comparison for security
- Cache recent attempts in Redis
- Rate limiting on verification attempts

## Security Considerations

### PIN Security
- Generate cryptographically secure random pins
- Never log actual PIN values
- Use bcrypt hashing for storage
- PIN expires after 10 minutes
- Block after 3 failed attempts

### Anti-Fraud
- Track verification location
- Detect impossible travel for second PIN usage
- Alert if different devices used
- Monitor for brute force patterns

### Privacy
- Don't share PIN details in logs
- Audit all PIN access
- Implement data retention policies

## Testing Strategy

### Unit Tests
- PIN generation randomness
- PIN hashing and verification
- Expiry calculation
- Attempt tracking

### Integration Tests
- End-to-end PIN workflow
- Multiple verification scenarios
- Expiry and re-generation
- Failed attempt handling

## Integration Points

### With Rides Module
- PIN generated when ride accepted
- PIN verified before ride can start
- Verification status blocks ride start

### With Notifications Module
- Send PIN via SMS and push
- Reminder notifications if not verified

### With Fraud Module
- Detect suspicious PIN patterns
- Track failed verification patterns

## Common Pitfalls

1. **Not hashing PIN** - Stored PIN can be compromised
2. **Too long PIN** - Users forget, too many failed attempts
3. **No expiry** - Old PINs become security risk
4. **Not tracking attempts** - Can't detect brute force
5. **Allowing too many retries** - Enables brute force attacks
6. **Sending PIN in SMS without encryption indicator** - Privacy concern

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
