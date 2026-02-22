# Promotions Module Development Guide

## Overview

The Promotions Module manages promotional codes, discounts, campaigns, and redemption tracking.

## Key Responsibilities

1. Promo Code Management - Create and manage codes
2. Discount Application - Apply discounts to transactions
3. Campaign Management - Run promotional campaigns
4. Usage Tracking - Track code usage
5. Redemption Processing - Handle redemptions

## Data Transfer Objects

### CreatePromoRequest

```go
type CreatePromoRequest struct {
    Code                string    `json:"code" binding:"required"`
    DiscountType        string    `json:"discount_type" binding:"required"` // percentage, fixed
    DiscountValue       float64   `json:"discount_value" binding:"required"`
    MaxDiscount         *float64  `json:"max_discount,omitempty"`
    MinOrderValue       float64   `json:"min_order_value,omitempty"`
    ValidFrom           time.Time `json:"valid_from" binding:"required"`
    ValidUntil          time.Time `json:"valid_until" binding:"required"`
    MaxUses             *int      `json:"max_uses,omitempty"`
    UsesPerUser         *int      `json:"uses_per_user,omitempty"`
    ApplicableCategories []string `json:"applicable_categories,omitempty"`
    Description         string    `json:"description,omitempty"`
}
```

### PromoResponse

```go
type PromoResponse struct {
    ID                  string    `json:"id"`
    Code                string    `json:"code"`
    DiscountType        string    `json:"discount_type"`
    DiscountValue       float64   `json:"discount_value"`
    MaxDiscount         float64   `json:"max_discount,omitempty"`
    MinOrderValue       float64   `json:"min_order_value"`
    ValidFrom           time.Time `json:"valid_from"`
    ValidUntil          time.Time `json:"valid_until"`
    MaxUses             int       `json:"max_uses,omitempty"`
    CurrentUses         int       `json:"current_uses"`
    UsesPerUser         int       `json:"uses_per_user"`
    Status              string    `json:"status"`
    IsActive            bool      `json:"is_active"`
    Description         string    `json:"description"`
    ApplicableCategories []string `json:"applicable_categories"`
}
```

### ValidatePromoRequest

```go
type ValidatePromoRequest struct {
    PromoCode   string  `json:"promo_code" binding:"required"`
    OrderValue  float64 `json:"order_value" binding:"required"`
    Category    string  `json:"category"`
    UserID      string  `json:"user_id"`
}
```

### ValidatePromoResponse

```go
type ValidatePromoResponse struct {
    IsValid             bool    `json:"is_valid"`
    DiscountAmount      float64 `json:"discount_amount"`
    FinalAmount         float64 `json:"final_amount"`
    Message             string  `json:"message,omitempty"`
    ExpiresAt           time.Time `json:"expires_at,omitempty"`
    RemainingUses       int     `json:"remaining_uses,omitempty"`
}
```

## Handler Methods

```
CreatePromo(c *gin.Context)            // POST /promotions (admin)
GetPromo(c *gin.Context)               // GET /promotions/{id}
ListPromos(c *gin.Context)             // GET /promotions
UpdatePromo(c *gin.Context)            // PUT /promotions/{id} (admin)
DeletePromo(c *gin.Context)            // DELETE /promotions/{id} (admin)
ValidatePromo(c *gin.Context)          // POST /promotions/validate
ApplyPromo(c *gin.Context)             // POST /promotions/apply
GetActivePromos(c *gin.Context)        // GET /promotions/active
GetUserRedemptions(c *gin.Context)     // GET /promotions/user/{id}/redemptions
```

## Service Methods

```
CreatePromo(ctx context.Context, req CreatePromoRequest) (*PromoResponse, error)
GetPromo(ctx context.Context, promoID string) (*PromoResponse, error)
ListPromos(ctx context.Context, filters map[string]interface{}) ([]PromoResponse, error)
UpdatePromo(ctx context.Context, promoID string, updates map[string]interface{}) error
DeletePromo(ctx context.Context, promoID string) error
ValidatePromo(ctx context.Context, req ValidatePromoRequest) (*ValidatePromoResponse, error)
ApplyPromo(ctx context.Context, userID, promoCode string, orderValue float64) (float64, error)
GetActivePromos(ctx context.Context) ([]PromoResponse, error)
GetUserRedemptions(ctx context.Context, userID string) ([]RedemptionResponse, error)
CheckUserEligibility(ctx context.Context, userID, promoCode string) (bool, string, error)
CreateCampaign(ctx context.Context, campaign *Campaign) error
DistributePromos(ctx context.Context, campaignID string) error
```

## Discount Types

```
PERCENTAGE:
- Discount = OrderValue * (DiscountValue / 100)
- Max capped at MaxDiscount if set

FIXED:
- Discount = DiscountValue (fixed amount)
- Cannot exceed OrderValue

COMBINATION:
- Multiple promos may be applicable
- Apply highest benefit to user
- No stacking unless explicitly allowed
```

## Typical Use Cases

### 1. Create Promotional Code

Request:
```
POST /promotions
{
    "code": "SUMMER20",
    "discount_type": "percentage",
    "discount_value": 20,
    "max_discount": 50,
    "min_order_value": 100,
    "valid_from": "2024-06-01T00:00:00Z",
    "valid_until": "2024-08-31T23:59:59Z",
    "max_uses": 1000,
    "uses_per_user": 5,
    "description": "Summer special offer"
}
```

Flow:
1. Validate code doesn't already exist
2. Validate date range
3. Create promo record
4. Set status to ACTIVE
5. Return confirmation

### 2. Validate Promo Code

Request:
```
POST /promotions/validate
{
    "promo_code": "SUMMER20",
    "order_value": 250.00,
    "user_id": "user-123"
}
```

Response:
```json
{
    "is_valid": true,
    "discount_amount": 50.00,
    "final_amount": 200.00,
    "expires_at": "2024-08-31T23:59:59Z",
    "remaining_uses": 4
}
```

Flow:
1. Find promo by code
2. Check if expired
3. Check usage limits
4. Check user eligibility
5. Calculate discount
6. Return validation result

### 3. Apply Promo to Order

Request:
```
POST /promotions/apply
{
    "promo_code": "SUMMER20"
}
```

Flow:
1. Validate code (call ValidatePromo)
2. Record redemption
3. Increment usage count
4. Return discount amount
5. Deduct from order total

### 4. Create Campaign

Campaign Structure:
```
Campaign:
- Target audience (new users, regular users, high-spenders)
- Promo codes to distribute
- Distribution method (SMS, email, in-app)
- Start and end date
- Expected reach
```

Flow:
1. Create campaign record
2. Generate promo codes for campaign
3. Target user segments
4. Schedule distribution
5. Track campaign effectiveness

## Validation Rules

```
Before applying promo:
1. Code exists and is active
2. Code hasn't expired
3. User hasn't exceeded max uses per user
4. Total uses haven't exceeded limit
5. Order value meets minimum
6. User is eligible (new/existing)
7. Category matches (if restricted)
8. No conflicting promos active
```

## Promo Code Generation

```
Format: CAMPAIGN_RANDOM_CODE

Examples:
- SUMMER20_ABC123
- NEWYEAR50_XYZ789
- REFERRAL25_LMN456

Uniqueness: Guaranteed
Length: Configurable (8-15 chars)
Charset: Alphanumeric (A-Z, 0-9)
```

## Campaign Types

```
SEASONAL:
- Summer, Winter, Holiday campaigns
- Fixed date ranges
- Theme-based codes

USER_BASED:
- New user onboarding
- VIP rewards
- Loyalty programs
- Win-back campaigns

EVENT_BASED:
- Launch events
- Milestone celebrations
- Sports/entertainment events
```

## Database Schema

### Promotions Table

```sql
CREATE TABLE promotions (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_type VARCHAR(50),
    discount_value DECIMAL(10, 2),
    max_discount DECIMAL(10, 2),
    min_order_value DECIMAL(10, 2),
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    max_uses INT,
    current_uses INT DEFAULT 0,
    uses_per_user INT,
    status VARCHAR(50),
    description TEXT,
    created_by VARCHAR(36),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    INDEX (code),
    INDEX (status, valid_until)
);
```

### Promotions Categories Table

```sql
CREATE TABLE promotion_categories (
    id VARCHAR(36) PRIMARY KEY,
    promotion_id VARCHAR(36) NOT NULL,
    category VARCHAR(50),
    FOREIGN KEY (promotion_id) REFERENCES promotions(id)
);
```

### Redemptions Table

```sql
CREATE TABLE redemptions (
    id VARCHAR(36) PRIMARY KEY,
    promotion_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    order_id VARCHAR(36),
    discount_amount DECIMAL(10, 2),
    redeemed_at TIMESTAMP,
    FOREIGN KEY (promotion_id) REFERENCES promotions(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (user_id, promotion_id),
    INDEX (redeemed_at)
);
```

---
