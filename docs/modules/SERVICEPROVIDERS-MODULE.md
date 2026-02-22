# Service Providers Module Documentation

## Overview

The Service Providers Module manages service provider accounts, profiles, service listings, approval workflows, earnings, and ratings. It handles laundry professionals, home service providers, and other service-based providers in the platform.

## Key Responsibilities

- Service provider account creation and profile management
- Service offerings and pricing management
- Service provider verification and approval workflow
- Availability and scheduling management
- Order/request fulfillment tracking
- Earnings calculation and settlements
- Rating and review management
- Service provider documents and certifications
- Commission and payment management

## Architecture

### Handler Layer (`serviceproviders/handler.go`)

Handles HTTP requests related to service provider operations.

**Key Endpoints:**

```go
POST /api/v1/service-providers/register         // Register as service provider
GET /api/v1/service-providers/profile           // Get provider profile
PUT /api/v1/service-providers/profile           // Update profile
GET /api/v1/service-providers/services          // Get provided services
POST /api/v1/service-providers/services         // Add service offering
PUT /api/v1/service-providers/services/:id      // Update service
DELETE /api/v1/service-providers/services/:id   // Remove service
GET /api/v1/service-providers/availability      // Get availability schedule
PUT /api/v1/service-providers/availability      // Set availability
GET /api/v1/service-providers/orders            // Get all orders
GET /api/v1/service-providers/orders/:id        // Get order details
PUT /api/v1/service-providers/orders/:id        // Update order status
GET /api/v1/service-providers/earnings          // Get earnings summary
GET /api/v1/service-providers/ratings           // Get reviews and ratings
```

### Service Layer (`serviceproviders/service.go`)

Contains business logic for service provider operations.

**Key Methods:**

```go
func (s *ServiceProviderService) RegisterProvider(ctx context.Context, req *RegisterProviderRequest) (*ProviderResponse, error)
func (s *ServiceProviderService) GetProfile(ctx context.Context, providerID string) (*ProviderProfileResponse, error)
func (s *ServiceProviderService) UpdateProfile(ctx context.Context, providerID string, req *UpdateProfileRequest) (*ProviderProfileResponse, error)
func (s *ServiceProviderService) AddService(ctx context.Context, providerID string, req *AddServiceRequest) (*ServiceOfferingResponse, error)
func (s *ServiceProviderService) UpdateService(ctx context.Context, serviceID string, req *UpdateServiceRequest) (*ServiceOfferingResponse, error)
func (s *ServiceProviderService) GetServices(ctx context.Context, providerID string) ([]*ServiceOfferingResponse, error)
func (s *ServiceProviderService) SetAvailability(ctx context.Context, providerID string, req *AvailabilityRequest) error
func (s *ServiceProviderService) GetOrders(ctx context.Context, providerID string, filters *OrderFilters) ([]*OrderResponse, error)
func (s *ServiceProviderService) UpdateOrderStatus(ctx context.Context, orderID string, status string) error
func (s *ServiceProviderService) GetEarnings(ctx context.Context, providerID string) (*EarningsResponse, error)
func (s *ServiceProviderService) GetRatings(ctx context.Context, providerID string) (*RatingsResponse, error)
```

### Repository Layer (`serviceproviders/repository.go`)

Manages database operations.

**Key Methods:**

```go
func (r *ServiceProviderRepository) Create(ctx context.Context, provider *ServiceProvider) error
func (r *ServiceProviderRepository) GetByID(ctx context.Context, providerID string) (*ServiceProvider, error)
func (r *ServiceProviderRepository) Update(ctx context.Context, providerID string, updates map[string]interface{}) error
func (r *ServiceProviderRepository) GetServices(ctx context.Context, providerID string) ([]*ServiceOffering, error)
func (r *ServiceProviderRepository) AddService(ctx context.Context, service *ServiceOffering) error
func (r *ServiceProviderRepository) GetOrders(ctx context.Context, providerID string, filters *OrderFilters) ([]*Order, error)
func (r *ServiceProviderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string) error
```

## Data Models

### ServiceProvider

```go
type ServiceProvider struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`                     // Links to user in auth
    BusinessName        string     `db:"business_name" json:"business_name"`
    FirstName           string     `db:"first_name" json:"first_name"`
    LastName            string     `db:"last_name" json:"last_name"`
    Email               string     `db:"email" json:"email"`
    Phone               string     `db:"phone" json:"phone"`
    Avatar              *string    `db:"avatar" json:"avatar"`
    
    // Business Details
    BusinessType        string     `db:"business_type" json:"business_type"`         // Laundry, Cleaning, Plumbing, etc.
    BusinessLicense     string     `db:"business_license" json:"business_license"`
    LicenseExpiry       *time.Time `db:"license_expiry" json:"license_expiry"`
    TaxID               *string    `db:"tax_id" json:"tax_id"`
    
    // Address
    BusinessAddress     string     `db:"business_address" json:"business_address"`
    City                string     `db:"city" json:"city"`
    State               *string    `db:"state" json:"state"`
    ZipCode             *string    `db:"zip_code" json:"zip_code"`
    Country             string     `db:"country" json:"country"`
    Latitude            float64    `db:"latitude" json:"latitude"`
    Longitude           float64    `db:"longitude" json:"longitude"`
    
    // Verification
    VerificationStatus  string     `db:"verification_status" json:"verification_status"` // PENDING, APPROVED, REJECTED
    DocumentsVerified   bool       `db:"documents_verified" json:"documents_verified"`
    VerifiedAt          *time.Time `db:"verified_at" json:"verified_at"`
    VerifiedBy          *string    `db:"verified_by" json:"verified_by"`              // Admin ID
    
    // Status
    Status              string     `db:"status" json:"status"`                        // ACTIVE, INACTIVE, SUSPENDED
    IsActive            bool       `db:"is_active" json:"is_active"`
    
    // Statistics
    TotalOrders         int        `db:"total_orders" json:"total_orders"`
    CompletedOrders     int        `db:"completed_orders" json:"completed_orders"`
    AverageRating       float64    `db:"average_rating" json:"average_rating"`
    TotalEarnings       float64    `db:"total_earnings" json:"total_earnings"`
    MonthlyEarnings     float64    `db:"monthly_earnings" json:"monthly_earnings"`
    ResponseTime        int        `db:"response_time" json:"response_time"`         // average in minutes
    CompletionRate      float64    `db:"completion_rate" json:"completion_rate"`
    
    // Metadata
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
    LastActiveAt        *time.Time `db:"last_active_at" json:"last_active_at"`
}
```

### ServiceOffering

```go
type ServiceOffering struct {
    ID                  string     `db:"id" json:"id"`
    ProviderID          string     `db:"provider_id" json:"provider_id"`
    ServiceType         string     `db:"service_type" json:"service_type"`           // Type of service
    Name                string     `db:"name" json:"name"`                           // Service name
    Description         string     `db:"description" json:"description"`
    BasePrice           float64    `db:"base_price" json:"base_price"`
    CurrencyCode        string     `db:"currency_code" json:"currency_code"`         // USD, EUR, etc.
    PricingModel        string     `db:"pricing_model" json:"pricing_model"`         // FIXED, HOURLY, WEIGHT_BASED
    EstimatedDuration   int        `db:"estimated_duration" json:"estimated_duration"` // in minutes
    Category            string     `db:"category" json:"category"`
    IsActive            bool       `db:"is_active" json:"is_active"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### ProviderAvailability

```go
type ProviderAvailability struct {
    ID                  string     `db:"id" json:"id"`
    ProviderID          string     `db:"provider_id" json:"provider_id"`
    DayOfWeek           int        `db:"day_of_week" json:"day_of_week"`             // 0=Sunday to 6=Saturday
    StartTime           string     `db:"start_time" json:"start_time"`               // HH:MM format
    EndTime             string     `db:"end_time" json:"end_time"`                   // HH:MM format
    IsAvailable         bool       `db:"is_available" json:"is_available"`
    MaxOrdersPerDay     int        `db:"max_orders_per_day" json:"max_orders_per_day"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### ProviderOrder

```go
type ProviderOrder struct {
    ID                  string     `db:"id" json:"id"`
    ProviderID          string     `db:"provider_id" json:"provider_id"`
    RequestedBy         string     `db:"requested_by" json:"requested_by"`          // Rider/Laundry user ID
    ServiceID           string     `db:"service_id" json:"service_id"`
    OrderStatus         string     `db:"order_status" json:"order_status"`          // REQUESTED, ACCEPTED, IN_PROGRESS, COMPLETED, CANCELLED
    PickupLocation      *string    `db:"pickup_location" json:"pickup_location"`
    DropoffLocation     *string    `db:"dropoff_location" json:"dropoff_location"`
    ScheduledTime       time.Time  `db:"scheduled_time" json:"scheduled_time"`
    StartTime           *time.Time `db:"start_time" json:"start_time"`
    CompletionTime      *time.Time `db:"completion_time" json:"completion_time"`
    Amount              float64    `db:"amount" json:"amount"`
    Commission          float64    `db:"commission" json:"commission"`              // Platform commission
    ProviderEarnings    float64    `db:"provider_earnings" json:"provider_earnings"`
    PaymentStatus       string     `db:"payment_status" json:"payment_status"`      // PENDING, COMPLETED, REFUNDED
    Rating              *int       `db:"rating" json:"rating"`                       // 1-5 stars
    Review              *string    `db:"review" json:"review"`
    CancellationReason  *string    `db:"cancellation_reason" json:"cancellation_reason"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

## DTOs (Data Transfer Objects)

### RegisterProviderRequest

```go
type RegisterProviderRequest struct {
    BusinessName        string `json:"business_name" binding:"required"`
    FirstName           string `json:"first_name" binding:"required"`
    LastName            string `json:"last_name" binding:"required"`
    Email               string `json:"email" binding:"required,email"`
    Phone               string `json:"phone" binding:"required"`
    BusinessType        string `json:"business_type" binding:"required"`
    BusinessLicense     string `json:"business_license" binding:"required"`
    TaxID               *string `json:"tax_id"`
    BusinessAddress     string `json:"business_address" binding:"required"`
    City                string `json:"city" binding:"required"`
    Latitude            float64 `json:"latitude" binding:"required"`
    Longitude           float64 `json:"longitude" binding:"required"`
}
```

### ProviderProfileResponse

```go
type ProviderProfileResponse struct {
    ID                  string     `json:"id"`
    BusinessName        string     `json:"business_name"`
    FirstName           string     `json:"first_name"`
    LastName            string     `json:"last_name"`
    Email               string     `json:"email"`
    Phone               string     `json:"phone"`
    BusinessType        string     `json:"business_type"`
    City                string     `json:"city"`
    VerificationStatus  string     `json:"verification_status"`
    Status              string     `json:"status"`
    TotalOrders         int        `json:"total_orders"`
    CompletedOrders     int        `json:"completed_orders"`
    AverageRating       float64    `json:"average_rating"`
    MonthlyEarnings     float64    `json:"monthly_earnings"`
    CompletionRate      float64    `json:"completion_rate"`
    CreatedAt           time.Time  `json:"created_at"`
}
```

### ServiceOfferingResponse

```go
type ServiceOfferingResponse struct {
    ID                  string     `json:"id"`
    ServiceType         string     `json:"service_type"`
    Name                string     `json:"name"`
    Description         string     `json:"description"`
    BasePrice           float64    `json:"base_price"`
    PricingModel        string     `json:"pricing_model"`
    EstimatedDuration   int        `json:"estimated_duration"`
    Category            string     `json:"category"`
    IsActive            bool       `json:"is_active"`
}
```

### EarningsResponse

```go
type EarningsResponse struct {
    TotalEarnings       float64    `json:"total_earnings"`
    ThisMonthEarnings   float64    `json:"this_month_earnings"`
    ThisWeekEarnings    float64    `json:"this_week_earnings"`
    TodayEarnings       float64    `json:"today_earnings"`
    PendingEarnings     float64    `json:"pending_earnings"`
    SettledEarnings     float64    `json:"settled_earnings"`
    CommissionRate      float64    `json:"commission_rate"`
    TotalOrders         int        `json:"total_orders"`
    AvgOrderValue       float64    `json:"avg_order_value"`
}
```

## Database Schema

### service_providers table

```sql
CREATE TABLE service_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    business_name VARCHAR(200) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    avatar TEXT,
    business_type VARCHAR(100) NOT NULL,
    business_license VARCHAR(100) UNIQUE NOT NULL,
    license_expiry TIMESTAMP,
    tax_id VARCHAR(50),
    business_address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    zip_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'US',
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    verification_status VARCHAR(50) DEFAULT 'PENDING' CHECK (verification_status IN ('PENDING', 'APPROVED', 'REJECTED')),
    documents_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    verified_by UUID REFERENCES admins(id),
    status VARCHAR(50) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    is_active BOOLEAN DEFAULT TRUE,
    total_orders INTEGER DEFAULT 0,
    completed_orders INTEGER DEFAULT 0,
    average_rating DECIMAL(3, 2) DEFAULT 0,
    total_earnings DECIMAL(15, 2) DEFAULT 0,
    monthly_earnings DECIMAL(15, 2) DEFAULT 0,
    response_time INTEGER,
    completion_rate DECIMAL(5, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP
);

CREATE INDEX idx_service_providers_user_id ON service_providers(user_id);
CREATE INDEX idx_service_providers_business_type ON service_providers(business_type);
CREATE INDEX idx_service_providers_city ON service_providers(city);
CREATE INDEX idx_service_providers_verification_status ON service_providers(verification_status);
CREATE INDEX idx_service_providers_status ON service_providers(status);
```

### service_offerings table

```sql
CREATE TABLE service_offerings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    service_type VARCHAR(100) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    base_price DECIMAL(10, 2) NOT NULL,
    currency_code VARCHAR(3) DEFAULT 'USD',
    pricing_model VARCHAR(50) CHECK (pricing_model IN ('FIXED', 'HOURLY', 'WEIGHT_BASED')),
    estimated_duration INTEGER,
    category VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_service_offerings_provider_id ON service_offerings(provider_id);
CREATE INDEX idx_service_offerings_service_type ON service_offerings(service_type);
CREATE INDEX idx_service_offerings_category ON service_offerings(category);
```

### provider_availability table

```sql
CREATE TABLE provider_availability (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES service_providers(id) ON DELETE CASCADE,
    day_of_week INTEGER CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_available BOOLEAN DEFAULT TRUE,
    max_orders_per_day INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, day_of_week)
);

CREATE INDEX idx_provider_availability_provider_id ON provider_availability(provider_id);
```

### provider_orders table

```sql
CREATE TABLE provider_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES service_providers(id),
    requested_by UUID NOT NULL,
    service_id UUID NOT NULL REFERENCES service_offerings(id),
    order_status VARCHAR(50) CHECK (order_status IN ('REQUESTED', 'ACCEPTED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    pickup_location TEXT,
    dropoff_location TEXT,
    scheduled_time TIMESTAMP NOT NULL,
    start_time TIMESTAMP,
    completion_time TIMESTAMP,
    amount DECIMAL(10, 2),
    commission DECIMAL(10, 2),
    provider_earnings DECIMAL(10, 2),
    payment_status VARCHAR(50) CHECK (payment_status IN ('PENDING', 'COMPLETED', 'REFUNDED')),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    cancellation_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_provider_orders_provider_id ON provider_orders(provider_id);
CREATE INDEX idx_provider_orders_order_status ON provider_orders(order_status);
CREATE INDEX idx_provider_orders_scheduled_time ON provider_orders(scheduled_time);
```

## Use Cases

### Use Case 1: Service Provider Registration and Approval

```
1. Provider submits registration with business details
2. System validates business license uniqueness
3. Documents uploaded and stored
4. Admin reviews documents (1-3 days)
5. If approved: status changes to ACTIVE
6. If rejected: provider notified with reasons
7. Provider can resubmit after fixing issues
```

### Use Case 2: Service Provider Creates Service Offerings

```
1. Provider logs in to dashboard
2. Clicks "Add Service"
3. Fills in service details (name, type, price, duration)
4. System validates pricing
5. Service is immediately available for booking
6. Users can see service in search/browse
```

### Use Case 3: Order Fulfillment Workflow

```
1. Customer requests service
2. System notifies matching providers
3. Provider accepts order
4. System sends pickup/delivery instructions
5. Provider starts service (status: IN_PROGRESS)
6. Provider completes and uploads proof
7. Customer rates service
8. Earnings added to provider wallet
9. Provider can request settlement
```

### Use Case 4: Earnings and Settlement

```
1. Each completed order adds to monthly_earnings
2. Commission deducted (e.g., 20% platform fee)
3. Provider earnings = Amount - Commission
4. At end of month (or on demand):
5. Earnings transferred to provider's bank account
6. Transaction record created
7. Provider can view detailed earnings breakdown
```

## Common Operations

### Register as Service Provider

```go
handler := func(c *gin.Context) {
    var req RegisterProviderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    provider, err := s.serviceProviderService.RegisterProvider(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusCreated, provider)
}
```

### Add Service Offering

```go
handler := func(c *gin.Context) {
    providerID := c.GetString("user_id")
    var req AddServiceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    service, err := s.serviceProviderService.AddService(c.Request.Context(), providerID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusCreated, service)
}
```

### Accept Order

```go
handler := func(c *gin.Context) {
    orderID := c.Param("id")
    providerID := c.GetString("user_id")

    err := s.serviceProviderService.UpdateOrderStatus(c.Request.Context(), orderID, "ACCEPTED")
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Send notification to customer
    // Update provider's response_time metrics

    c.JSON(http.StatusOK, SuccessResponse{Message: "Order accepted"})
}
```

### Get Earnings

```go
handler := func(c *gin.Context) {
    providerID := c.GetString("user_id")
    
    earnings, err := s.serviceProviderService.GetEarnings(c.Request.Context(), providerID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, earnings)
}
```

## Error Handling

| Error | Status Code | Message |
|-------|------------|---------|
| Provider not found | 404 | "Service provider not found" |
| Invalid business license | 400 | "Business license format invalid" |
| License expired | 400 | "Business license has expired" |
| Unauthorized | 403 | "You don't have permission for this action" |
| Service not found | 404 | "Service offering not found" |
| Invalid price | 400 | "Price must be greater than 0" |
| Order not found | 404 | "Order not found" |
| Invalid status transition | 409 | "Cannot transition from {current} to {target}" |
| Not available | 409 | "Provider is not available at this time" |
| Max orders exceeded | 409 | "Maximum daily orders limit reached" |

## Performance Optimization

### Database Indexes
- Index on provider_id for quick lookups
- Index on business_type for filtering
- Index on city for location-based searches
- Index on order status for quick filtering
- Index on scheduled_time for date range queries

### Caching Strategy
- Cache provider profiles (1-hour TTL)
- Cache service offerings (12-hour TTL)
- Cache availability schedule (1-day TTL)
- Cache earnings (real-time updates)

### Query Optimization
- Use pagination for order lists
- Load only active services in searches
- Index on location for nearby provider searches
- Batch order status updates

## Security Considerations

### Verification
- Admin approval required before going active
- License verification with government databases
- Tax ID validation
- Address verification

### Authorization
- Only provider can view own orders and earnings
- Only admin can approve providers
- Payment transactions require verification
- Audit all earnings calculations

### Data Protection
- Encrypt business license documents
- Store tax ID securely
- Implement data retention policies
- Regular compliance audits

## Testing Strategy

### Unit Tests
- Provider registration validation
- Service offering creation
- Order status transitions
- Earnings calculation
- Availability validation

### Integration Tests
- Complete registration to first order workflow
- Service creation and booking workflow
- Multi-order earnings calculation
- Availability schedule enforcement

## Integration Points

### With Auth Module
- Provider links to user account
- Phone/email verification

### With Laundry Module
- Service providers fulfill laundry orders
- Order tracking and status updates

### With Wallet Module
- Earnings transferred to provider wallet
- Settlement and payment processing

### With Ratings Module
- Provider ratings and reviews
- Reputation management

## Common Pitfalls

1. **Not validating business license** - Must be unique and non-expired
2. **Missing availability check** - Ensure provider is available before assigning order
3. **Not tracking response time** - Critical for metrics
4. **Incorrect commission calculation** - Should be applied consistently
5. **Not handling license expiry** - Should suspend provider
6. **Missing document verification** - Critical for trust

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
