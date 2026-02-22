# Fraud Module Documentation

## Overview

The Fraud Module handles fraud detection, prevention, monitoring, and investigation across the platform. It detects suspicious activities, patterns, and behaviors that indicate fraudulent transactions or misuse of the platform.

## Key Responsibilities

- Real-time fraud detection and alerts
- Suspicious activity monitoring
- Pattern analysis and anomaly detection
- Risk scoring and risk management
- Fraud investigation and dispute resolution
- Account restrictions and controls
- Transaction verification
- Blacklist and whitelist management
- Fraud reporting and analytics

## Architecture

### Handler Layer (`fraud/handler.go`)

Handles HTTP requests related to fraud monitoring and investigation.

**Key Endpoints:**

```go
GET /api/v1/fraud/alerts                 // Get fraud alerts
GET /api/v1/fraud/alerts/:id             // Get alert details
POST /api/v1/fraud/report                // Report suspicious activity
GET /api/v1/fraud/cases                  // Get fraud cases (admin)
GET /api/v1/fraud/cases/:id              // Get case details
PUT /api/v1/fraud/cases/:id              // Update case status
GET /api/v1/fraud/risk-score/:user_id    // Get user risk score
POST /api/v1/fraud/review-transaction    // Review transaction
GET /api/v1/fraud/blacklist              // Get blacklisted users (admin)
POST /api/v1/fraud/blacklist             // Add to blacklist (admin)
```

### Service Layer (`fraud/service.go`)

Contains business logic for fraud detection and management.

**Key Methods:**

```go
func (s *FraudService) DetectFraud(ctx context.Context, transaction *Transaction) (*FraudAnalysis, error)
func (s *FraudService) CalculateRiskScore(ctx context.Context, userID string) (float64, error)
func (s *FraudService) CheckAnomalies(ctx context.Context, userID string) ([]*Anomaly, error)
func (s *FraudService) AnalyzePattern(ctx context.Context, userID string) (*PatternAnalysis, error)
func (s *FraudService) BlockAccount(ctx context.Context, userID string, reason string) error
func (s *FraudService) RestrictUser(ctx context.Context, userID string, restrictions *Restrictions) error
func (s *FraudService) InvestigateCase(ctx context.Context, caseID string) (*CaseAnalysis, error)
func (s *FraudService) ReportFraud(ctx context.Context, req *ReportFraudRequest) (*FraudCase, error)
func (s *FraudService) IsBlacklisted(ctx context.Context, userID string) (bool, error)
func (s *FraudService) VerifyTransaction(ctx context.Context, transactionID string) (*VerificationResult, error)
```

### Repository Layer (`fraud/repository.go`)

Manages database operations.

**Key Methods:**

```go
func (r *FraudRepository) CreateAlert(ctx context.Context, alert *FraudAlert) error
func (r *FraudRepository) GetAlerts(ctx context.Context, filters *AlertFilters) ([]*FraudAlert, error)
func (r *FraudRepository) CreateCase(ctx context.Context, caseData *FraudCase) error
func (r *FraudRepository) UpdateCase(ctx context.Context, caseID string, updates map[string]interface{}) error
func (r *FraudRepository) GetRiskScore(ctx context.Context, userID string) (*RiskScore, error)
func (r *FraudRepository) SaveRiskScore(ctx context.Context, riskScore *RiskScore) error
func (r *FraudRepository) AddToBlacklist(ctx context.Context, blacklistEntry *BlacklistEntry) error
func (r *FraudRepository) CheckBlacklist(ctx context.Context, userID string) (bool, error)
```

## Data Models

### FraudAlert

```go
type FraudAlert struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    AlertType           string     `db:"alert_type" json:"alert_type"`              // VELOCITY_CHECK, DUPLICATE_ACCOUNT, UNUSUAL_LOCATION, etc.
    Severity            string     `db:"severity" json:"severity"`                  // LOW, MEDIUM, HIGH, CRITICAL
    Description         string     `db:"description" json:"description"`
    RiskScore           float64    `db:"risk_score" json:"risk_score"`              // 0-100
    TriggeringData      string     `db:"triggering_data" json:"triggering_data"`    // JSON data that triggered alert
    Action              string     `db:"action" json:"action"`                      // NONE, WARN, REVIEW, BLOCK
    Status              string     `db:"status" json:"status"`                      // PENDING, REVIEWED, RESOLVED, FALSE_POSITIVE
    AssignedTo          *string    `db:"assigned_to" json:"assigned_to"`            // Admin ID
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    ResolvedAt          *time.Time `db:"resolved_at" json:"resolved_at"`
    Notes               *string    `db:"notes" json:"notes"`
}
```

### FraudCase

```go
type FraudCase struct {
    ID                  string     `db:"id" json:"id"`
    ReportedBy          string     `db:"reported_by" json:"reported_by"`            // User or admin
    ReportedUser        string     `db:"reported_user" json:"reported_user"`
    CaseType            string     `db:"case_type" json:"case_type"`                // PAYMENT_FRAUD, IDENTITY_THEFT, ACCOUNT_TAKEOVER, CHARGEBACK, etc.
    Status              string     `db:"status" json:"status"`                      // OPEN, INVESTIGATING, RESOLVED, CLOSED
    Severity            string     `db:"severity" json:"severity"`                  // LOW, MEDIUM, HIGH, CRITICAL
    Description         string     `db:"description" json:"description"`
    Evidence            string     `db:"evidence" json:"evidence"`                  // JSON array of evidence URLs/data
    RelatedAlerts       []string   `db:"related_alerts" json:"related_alerts"`      // JSONB array of alert IDs
    TransactionIDs      []string   `db:"transaction_ids" json:"transaction_ids"`    // JSONB array
    
    // Investigation
    AssignedTo          *string    `db:"assigned_to" json:"assigned_to"`            // Admin ID
    InvestigationNotes  *string    `db:"investigation_notes" json:"investigation_notes"`
    Resolution          *string    `db:"resolution" json:"resolution"`
    
    // Actions Taken
    ActionTaken         *string    `db:"action_taken" json:"action_taken"`          // WARN, SUSPEND, BLOCK, ACCOUNT_RESTRICTION
    ActionDetails       *string    `db:"action_details" json:"action_details"`      // JSON details of action
    
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
    ResolvedAt          *time.Time `db:"resolved_at" json:"resolved_at"`
}
```

### RiskScore

```go
type RiskScore struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    OverallScore        float64    `db:"overall_score" json:"overall_score"`        // 0-100
    VelocityScore       float64    `db:"velocity_score" json:"velocity_score"`      // Too many transactions
    LocationScore       float64    `db:"location_score" json:"location_score"`      // Impossible travel
    DeviceScore         float64    `db:"device_score" json:"device_score"`          // Device changes
    PatternScore        float64    `db:"pattern_score" json:"pattern_score"`        // Unusual patterns
    RatingScore         float64    `db:"rating_score" json:"rating_score"`          // Suspicious ratings
    AccountScore        float64    `db:"account_score" json:"account_score"`        // New account, no history
    
    RiskLevel           string     `db:"risk_level" json:"risk_level"`              // LOW, MEDIUM, HIGH, CRITICAL
    LastUpdatedAt       time.Time  `db:"last_updated_at" json:"last_updated_at"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}
```

### BlacklistEntry

```go
type BlacklistEntry struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    Reason              string     `db:"reason" json:"reason"`                      // Repeated fraud, chargeback, etc.
    Severity            string     `db:"severity" json:"severity"`                  // TEMPORARY, PERMANENT
    ExpiryDate          *time.Time `db:"expiry_date" json:"expiry_date"`            // For temporary blocks
    AddedBy             string     `db:"added_by" json:"added_by"`                  // Admin ID
    Email               string     `db:"email" json:"email"`                        // Associated email
    Phone               *string    `db:"phone" json:"phone"`                        // Associated phone
    IPAddresses         []string   `db:"ip_addresses" json:"ip_addresses"`          // JSONB array
    Devices             []string   `db:"devices" json:"devices"`                    // JSONB array
    Notes               *string    `db:"notes" json:"notes"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### Anomaly

```go
type Anomaly struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`
    AnomalyType         string     `db:"anomaly_type" json:"anomaly_type"`         // VELOCITY, LOCATION, DEVICE, AMOUNT, TIME, PATTERN
    Description         string     `db:"description" json:"description"`
    Value               float64    `db:"value" json:"value"`                        // Current value
    Threshold           float64    `db:"threshold" json:"threshold"`                // Normal threshold
    Severity            string     `db:"severity" json:"severity"`                  // LOW, MEDIUM, HIGH
    Status              string     `db:"status" json:"status"`                      // DETECTED, REVIEWING, APPROVED, REJECTED
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

## DTOs (Data Transfer Objects)

### FraudAnalysis

```go
type FraudAnalysis struct {
    TransactionID       string     `json:"transaction_id"`
    UserID              string     `json:"user_id"`
    RiskScore           float64    `json:"risk_score"`                              // 0-100
    RiskLevel           string     `json:"risk_level"`                              // LOW, MEDIUM, HIGH, CRITICAL
    Factors             []string   `json:"factors"`                                 // List of risk factors
    Recommendation      string     `json:"recommendation"`                          // APPROVE, REVIEW, DECLINE
    ShouldBlock         bool       `json:"should_block"`
    CreatedAt           time.Time  `json:"created_at"`
}
```

### ReportFraudRequest

```go
type ReportFraudRequest struct {
    FraudType           string     `json:"fraud_type" binding:"required"`
    ReportedUserID      string     `json:"reported_user_id" binding:"required"`
    Description         string     `json:"description" binding:"required"`
    Evidence            []string   `json:"evidence"`                                 // URLs to evidence
    RelatedTransaction  *string    `json:"related_transaction"`
    ContactMethod       string     `json:"contact_method"`                           // Email or phone for follow-up
}
```

### VerificationResult

```go
type VerificationResult struct {
    TransactionID       string     `json:"transaction_id"`
    Status              string     `json:"status"`                                   // VERIFIED, SUSPICIOUS, BLOCKED
    RiskScore           float64    `json:"risk_score"`
    Reason              string     `json:"reason"`
    RequiresApproval    bool       `json:"requires_approval"`
}
```

## Database Schema

### fraud_alerts table

```sql
CREATE TABLE fraud_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    description TEXT,
    risk_score DECIMAL(5, 2),
    triggering_data JSONB,
    action VARCHAR(50),
    status VARCHAR(50) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'REVIEWED', 'RESOLVED', 'FALSE_POSITIVE')),
    assigned_to UUID REFERENCES admins(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    notes TEXT
);

CREATE INDEX idx_fraud_alerts_user_id ON fraud_alerts(user_id);
CREATE INDEX idx_fraud_alerts_severity ON fraud_alerts(severity);
CREATE INDEX idx_fraud_alerts_status ON fraud_alerts(status);
CREATE INDEX idx_fraud_alerts_created_at ON fraud_alerts(created_at);
```

### fraud_cases table

```sql
CREATE TABLE fraud_cases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reported_by UUID NOT NULL,
    reported_user UUID NOT NULL REFERENCES users(id),
    case_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'INVESTIGATING', 'RESOLVED', 'CLOSED')),
    severity VARCHAR(50) CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    description TEXT NOT NULL,
    evidence JSONB,
    related_alerts UUID[] DEFAULT ARRAY[]::UUID[],
    transaction_ids UUID[] DEFAULT ARRAY[]::UUID[],
    assigned_to UUID REFERENCES admins(id),
    investigation_notes TEXT,
    resolution TEXT,
    action_taken VARCHAR(100),
    action_details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

CREATE INDEX idx_fraud_cases_reported_user ON fraud_cases(reported_user);
CREATE INDEX idx_fraud_cases_status ON fraud_cases(status);
CREATE INDEX idx_fraud_cases_created_at ON fraud_cases(created_at);
```

### risk_scores table

```sql
CREATE TABLE risk_scores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    overall_score DECIMAL(5, 2),
    velocity_score DECIMAL(5, 2),
    location_score DECIMAL(5, 2),
    device_score DECIMAL(5, 2),
    pattern_score DECIMAL(5, 2),
    rating_score DECIMAL(5, 2),
    account_score DECIMAL(5, 2),
    risk_level VARCHAR(50),
    last_updated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_risk_scores_user_id ON risk_scores(user_id);
CREATE INDEX idx_risk_scores_overall_score ON risk_scores(overall_score);
```

### blacklist table

```sql
CREATE TABLE blacklist (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    reason VARCHAR(200) NOT NULL,
    severity VARCHAR(50) CHECK (severity IN ('TEMPORARY', 'PERMANENT')),
    expiry_date TIMESTAMP,
    added_by UUID NOT NULL REFERENCES admins(id),
    email VARCHAR(100),
    phone VARCHAR(20),
    ip_addresses JSONB,
    devices JSONB,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_blacklist_user_id ON blacklist(user_id);
CREATE INDEX idx_blacklist_email ON blacklist(email);
CREATE UNIQUE INDEX idx_blacklist_permanent ON blacklist(user_id) WHERE severity = 'PERMANENT';
```

### anomalies table

```sql
CREATE TABLE anomalies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    anomaly_type VARCHAR(100) NOT NULL,
    description TEXT,
    value DECIMAL(15, 2),
    threshold DECIMAL(15, 2),
    severity VARCHAR(50),
    status VARCHAR(50) DEFAULT 'DETECTED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_anomalies_user_id ON anomalies(user_id);
CREATE INDEX idx_anomalies_anomaly_type ON anomalies(anomaly_type);
```

## Fraud Detection Rules

### Velocity Check
- More than X transactions in Y minutes
- More than X amount in Y minutes
- More than X rides in a day

Scores: 0-25 points based on violation severity

### Location Anomaly
- Impossible travel (two transactions in different cities within unrealistic time)
- Sudden location change from usual area
- Multiple locations in parallel (account compromise indicator)

Scores: 0-25 points based on distance and time delta

### Device Anomaly
- New device with different characteristics
- Device profile mismatch
- Multiple devices accessing simultaneously

Scores: 0-20 points

### Pattern Analysis
- Rating patterns (too many 5 stars, rating own rides)
- Refund patterns (frequent cancellations for refunds)
- Payment patterns (alternating payment methods to avoid limits)

Scores: 0-15 points

### Account Age and History
- Brand new account (created < 24 hours)
- No history but high-value transaction
- Multiple accounts from same IP/device

Scores: 0-15 points

## Use Cases

### Use Case 1: Real-Time Transaction Verification

```
1. User initiates transaction (ride, order, etc.)
2. System calls DetectFraud()
3. Fraud analysis runs across all factors
4. Risk score calculated (0-100)
5. If score < 30: APPROVE automatically
6. If 30-70: REVIEW - send to admin queue
7. If > 70: BLOCK and flag account
8. Result returned within milliseconds
9. Transaction proceeds or blocked
```

### Use Case 2: Anomaly Detection

```
1. Scheduled job runs every 5 minutes
2. Checks velocity of recent transactions
3. Analyzes location patterns
4. Checks for device anomalies
5. If anomaly detected:
   - Create FraudAlert
   - Update RiskScore
   - Notify admin if severity > MEDIUM
6. User may be contacted for verification
```

### Use Case 3: Fraud Investigation

```
1. Admin receives fraud report or alert
2. Creates FraudCase with evidence
3. Links related alerts and transactions
4. Assigns to investigator
5. Investigator reviews evidence
6. Determines if fraud confirmed
7. If confirmed: Block account or apply restrictions
8. If false positive: Mark and whitelist
9. Case closed with resolution
```

### Use Case 4: Account Restriction

```
1. Multiple fraud alerts on account
2. Risk score becomes CRITICAL
3. System triggers restrictions:
   - Daily spending cap
   - Transaction limit per hour
   - Require 2FA for all transactions
   - Require manual approval
4. User notified of restrictions
5. Restrictions can be lifted after verification
```

## Common Operations

### Detect Fraud in Transaction

```go
handler := func(c *gin.Context) {
    var req TransactionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    transaction := &Transaction{
        UserID: c.GetString("user_id"),
        Amount: req.Amount,
        Timestamp: time.Now(),
        IPAddress: c.ClientIP(),
    }

    analysis, err := s.fraudService.DetectFraud(c.Request.Context(), transaction)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    if analysis.ShouldBlock {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "Transaction blocked for security",
            "reason": "Suspicious activity detected",
        })
        return
    }

    c.JSON(http.StatusOK, analysis)
}
```

### Calculate Risk Score

```go
handler := func(c *gin.Context) {
    userID := c.GetString("user_id")
    
    riskScore, err := s.fraudService.CalculateRiskScore(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, riskScore)
}
```

### Report Fraud

```go
handler := func(c *gin.Context) {
    var req ReportFraudRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    fraudCase, err := s.fraudService.ReportFraud(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusCreated, fraudCase)
}
```

## Error Handling

| Error | Status Code | Message |
|-------|------------|---------|
| User not found | 404 | "User not found" |
| Transaction not found | 404 | "Transaction not found" |
| Case not found | 404 | "Fraud case not found" |
| Invalid case type | 400 | "Invalid fraud case type" |
| Insufficient evidence | 400 | "Fraud report requires evidence" |
| User blacklisted | 403 | "User account is blacklisted" |
| Analysis failed | 500 | "Fraud analysis failed" |

## Performance Optimization

### Real-Time Processing
- Cache user risk profiles
- Use Redis for velocity checks
- Queue fraud detection jobs
- Async alert notifications

### Database
- Index on user_id for quick lookups
- Index on timestamps for time-based queries
- Partition tables by date
- Archive old cases

## Security Considerations

### Data Protection
- Encrypt case evidence
- Audit all investigations
- Implement data retention policies
- Secure blacklist access

### Investigation Privacy
- Only authorized admins can view cases
- Log all case access
- Anonymize sensitive information
- Implement approval workflows

## Testing Strategy

### Unit Tests
- Risk score calculation
- Anomaly detection algorithms
- Velocity checks
- Pattern analysis

### Integration Tests
- End-to-end fraud detection
- Case creation and investigation
- Blacklist enforcement
- Account restrictions

## Integration Points

### With Transaction Modules
- Check fraud before processing
- Block suspicious transactions

### With Wallet Module
- Verify high-value transactions
- Detect refund patterns

### With Ratings Module
- Detect suspicious rating patterns
- Flag rating manipulation

## Common Pitfalls

1. **Not updating risk scores regularly** - Stale scores cause false positives
2. **Ignoring location anomalies** - Account takeover indicator
3. **Missing device tracking** - Can't detect device anomalies
4. **Incorrect velocity thresholds** - Too strict = false positives, too loose = missed fraud
5. **Not investigating false positives** - Whitelist legitimate patterns
6. **Slow fraud detection** - Should be real-time or near-real-time

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
