# Home Services Module Documentation

## Overview

The Home Services Module manages home service requests, service provider matching, scheduling, job tracking, and service completion. It handles various home services such as cleaning, plumbing, electrical work, repairs, and other home-based services.

## Key Responsibilities

- Home service request creation and management
- Service category and type management
- Provider matching and availability checking
- Service scheduling and time slot management
- Job status tracking and updates
- Service completion and handover
- Cost estimation and billing
- Photo/evidence documentation
- Service review and feedback

## Architecture

### Handler Layer (`homeservices/handler.go`)

Handles HTTP requests related to home services.

**Key Endpoints:**

```go
POST /api/v1/home-services/request          // Create service request
GET /api/v1/home-services/requests          // Get user's requests
GET /api/v1/home-services/requests/:id      // Get request details
PUT /api/v1/home-services/requests/:id      // Update request
DELETE /api/v1/home-services/requests/:id   // Cancel request
GET /api/v1/home-services/categories        // Get service categories
GET /api/v1/home-services/available         // Get available providers
POST /api/v1/home-services/requests/:id/schedule // Schedule service
POST /api/v1/home-services/requests/:id/photos   // Upload service photos
GET /api/v1/home-services/requests/:id/photos    // Get service photos
PUT /api/v1/home-services/requests/:id/complete  // Mark as complete
POST /api/v1/home-services/requests/:id/rate     // Rate service
```

### Service Layer (`homeservices/service.go`)

Contains business logic for home service operations.

**Key Methods:**

```go
func (s *HomeServiceService) CreateRequest(ctx context.Context, req *CreateRequestRequest) (*ServiceRequestResponse, error)
func (s *HomeServiceService) GetRequest(ctx context.Context, requestID string) (*ServiceRequestResponse, error)
func (s *HomeServiceService) GetUserRequests(ctx context.Context, userID string, filters *RequestFilters) ([]*ServiceRequestResponse, error)
func (s *HomeServiceService) UpdateRequest(ctx context.Context, requestID string, req *UpdateRequestRequest) (*ServiceRequestResponse, error)
func (s *HomeServiceService) CancelRequest(ctx context.Context, requestID string, reason string) error
func (s *HomeServiceService) FindAvailableProviders(ctx context.Context, req *ProviderSearchRequest) ([]*AvailableProviderResponse, error)
func (s *HomeServiceService) AssignProvider(ctx context.Context, requestID, providerID string) error
func (s *HomeServiceService) ScheduleService(ctx context.Context, requestID string, req *ScheduleRequest) error
func (s *HomeServiceService) UpdateJobStatus(ctx context.Context, jobID string, status string) error
func (s *HomeServiceService) CompleteJob(ctx context.Context, jobID string) error
func (s *HomeServiceService) RateService(ctx context.Context, jobID string, req *RateServiceRequest) error
func (s *HomeServiceService) UploadPhotos(ctx context.Context, jobID string, files []*multipart.FileHeader) ([]string, error)
```

### Repository Layer (`homeservices/repository.go`)

Manages database operations.

**Key Methods:**

```go
func (r *HomeServiceRepository) CreateRequest(ctx context.Context, request *ServiceRequest) error
func (r *HomeServiceRepository) GetRequest(ctx context.Context, requestID string) (*ServiceRequest, error)
func (r *HomeServiceRepository) GetUserRequests(ctx context.Context, userID string) ([]*ServiceRequest, error)
func (r *HomeServiceRepository) UpdateRequest(ctx context.Context, requestID string, updates map[string]interface{}) error
func (r *HomeServiceRepository) FindAvailableProviders(ctx context.Context, criteria *ProviderSearchCriteria) ([]*ServiceProvider, error)
func (r *HomeServiceRepository) CreateJob(ctx context.Context, job *ServiceJob) error
func (r *HomeServiceRepository) GetJob(ctx context.Context, jobID string) (*ServiceJob, error)
func (r *HomeServiceRepository) UpdateJobStatus(ctx context.Context, jobID string, status string) error
```

## Data Models

### ServiceRequest

```go
type ServiceRequest struct {
    ID                  string     `db:"id" json:"id"`
    UserID              string     `db:"user_id" json:"user_id"`                     // Customer
    ServiceCategory     string     `db:"service_category" json:"service_category"`   // Cleaning, Plumbing, Electrical, etc.
    ServiceType         string     `db:"service_type" json:"service_type"`           // Specific type within category
    Description         string     `db:"description" json:"description"`
    Location            string     `db:"location" json:"location"`                   // Full address
    Latitude            float64    `db:"latitude" json:"latitude"`
    Longitude           float64    `db:"longitude" json:"longitude"`
    ScheduledDate       time.Time  `db:"scheduled_date" json:"scheduled_date"`
    ScheduledTimeSlot   string     `db:"scheduled_time_slot" json:"scheduled_time_slot"` // e.g., "09:00-11:00"
    
    // Provider
    AssignedProvider    *string    `db:"assigned_provider" json:"assigned_provider"`
    
    // Cost Estimation
    EstimatedCost       *float64   `db:"estimated_cost" json:"estimated_cost"`
    ActualCost          *float64   `db:"actual_cost" json:"actual_cost"`
    CurrencyCode        string     `db:"currency_code" json:"currency_code"`
    
    // Status
    Status              string     `db:"status" json:"status"`                       // PENDING, ASSIGNED, IN_PROGRESS, COMPLETED, CANCELLED
    RequestSource       string     `db:"request_source" json:"request_source"`       // WEB, MOBILE, PHONE
    
    // Images
    PhotoURLs           []string   `db:"photo_urls" json:"photo_urls"`               // JSONB array
    
    // Metadata
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
    ScheduledAt         *time.Time `db:"scheduled_at" json:"scheduled_at"`
    CompletedAt         *time.Time `db:"completed_at" json:"completed_at"`
    CancelledAt         *time.Time `db:"cancelled_at" json:"cancelled_at"`
    CancellationReason  *string    `db:"cancellation_reason" json:"cancellation_reason"`
}
```

### ServiceJob

```go
type ServiceJob struct {
    ID                  string     `db:"id" json:"id"`
    RequestID           string     `db:"request_id" json:"request_id"`
    ProviderID          string     `db:"provider_id" json:"provider_id"`
    UserID              string     `db:"user_id" json:"user_id"`
    
    ServiceCategory     string     `db:"service_category" json:"service_category"`
    ServiceType         string     `db:"service_type" json:"service_type"`
    Description         string     `db:"description" json:"description"`
    
    Location            string     `db:"location" json:"location"`
    Latitude            float64    `db:"latitude" json:"latitude"`
    Longitude           float64    `db:"longitude" json:"longitude"`
    
    ScheduledStartTime  time.Time  `db:"scheduled_start_time" json:"scheduled_start_time"`
    ScheduledEndTime    time.Time  `db:"scheduled_end_time" json:"scheduled_end_time"`
    ActualStartTime     *time.Time `db:"actual_start_time" json:"actual_start_time"`
    ActualEndTime       *time.Time `db:"actual_end_time" json:"actual_end_time"`
    
    Status              string     `db:"status" json:"status"`                       // PENDING, ACCEPTED, IN_PROGRESS, COMPLETED, CANCELLED
    Priority            string     `db:"priority" json:"priority"`                   // LOW, MEDIUM, HIGH, URGENT
    
    EstimatedCost       float64    `db:"estimated_cost" json:"estimated_cost"`
    ActualCost          *float64   `db:"actual_cost" json:"actual_cost"`
    Materials           *string    `db:"materials" json:"materials"`                 // JSON list of materials used
    LaborHours          *float64   `db:"labor_hours" json:"labor_hours"`
    
    Rating              *int       `db:"rating" json:"rating"`                       // 1-5
    Review              *string    `db:"review" json:"review"`
    
    PhotoURLs           []string   `db:"photo_urls" json:"photo_urls"`               // JSONB array
    Signature           *string    `db:"signature" json:"signature"`                 // Customer signature URL
    
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### ServiceCategory

```go
type ServiceCategory struct {
    ID                  string     `db:"id" json:"id"`
    Name                string     `db:"name" json:"name"`                           // Cleaning, Plumbing, etc.
    Code                string     `db:"code" json:"code"`
    Description         string     `db:"description" json:"description"`
    Icon                *string    `db:"icon" json:"icon"`
    IsActive            bool       `db:"is_active" json:"is_active"`
    EstimatedDuration   int        `db:"estimated_duration" json:"estimated_duration"` // in minutes
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}
```

## DTOs (Data Transfer Objects)

### CreateRequestRequest

```go
type CreateRequestRequest struct {
    ServiceCategory     string     `json:"service_category" binding:"required"`
    ServiceType         string     `json:"service_type" binding:"required"`
    Description         string     `json:"description" binding:"required"`
    Location            string     `json:"location" binding:"required"`
    Latitude            float64    `json:"latitude" binding:"required"`
    Longitude           float64    `json:"longitude" binding:"required"`
    ScheduledDate       time.Time  `json:"scheduled_date" binding:"required"`
    PreferredTimeSlot   string     `json:"preferred_time_slot"`
    EstimatedBudget     *float64   `json:"estimated_budget"`
}
```

### ServiceRequestResponse

```go
type ServiceRequestResponse struct {
    ID                  string     `json:"id"`
    ServiceCategory     string     `json:"service_category"`
    ServiceType         string     `json:"service_type"`
    Description         string     `json:"description"`
    Location            string     `json:"location"`
    ScheduledDate       time.Time  `json:"scheduled_date"`
    Status              string     `json:"status"`
    AssignedProvider    *string    `json:"assigned_provider"`
    EstimatedCost       *float64   `json:"estimated_cost"`
    ActualCost          *float64   `json:"actual_cost"`
    PhotoURLs           []string   `json:"photo_urls"`
    CreatedAt           time.Time  `json:"created_at"`
}
```

### ScheduleRequest

```go
type ScheduleRequest struct {
    ScheduledDate       time.Time  `json:"scheduled_date" binding:"required"`
    TimeSlotStart       string     `json:"time_slot_start" binding:"required"`      // HH:MM
    TimeSlotEnd         string     `json:"time_slot_end" binding:"required"`        // HH:MM
    Notes               *string    `json:"notes"`
}
```

### RateServiceRequest

```go
type RateServiceRequest struct {
    Rating              int        `json:"rating" binding:"required,min=1,max=5"`
    Review              *string    `json:"review"`
    ReportIssue         bool       `json:"report_issue"`
    IssueDescription    *string    `json:"issue_description"`
}
```

## Database Schema

### service_requests table

```sql
CREATE TABLE service_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    service_category VARCHAR(100) NOT NULL,
    service_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    location TEXT NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    scheduled_date DATE NOT NULL,
    scheduled_time_slot VARCHAR(20),
    assigned_provider UUID,
    estimated_cost DECIMAL(10, 2),
    actual_cost DECIMAL(10, 2),
    currency_code VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ASSIGNED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    request_source VARCHAR(50),
    photo_urls JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    scheduled_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT
);

CREATE INDEX idx_service_requests_user_id ON service_requests(user_id);
CREATE INDEX idx_service_requests_status ON service_requests(status);
CREATE INDEX idx_service_requests_scheduled_date ON service_requests(scheduled_date);
CREATE INDEX idx_service_requests_service_category ON service_requests(service_category);
```

### service_jobs table

```sql
CREATE TABLE service_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES service_requests(id),
    provider_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    service_category VARCHAR(100),
    service_type VARCHAR(100),
    description TEXT,
    location TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    scheduled_start_time TIMESTAMP NOT NULL,
    scheduled_end_time TIMESTAMP NOT NULL,
    actual_start_time TIMESTAMP,
    actual_end_time TIMESTAMP,
    status VARCHAR(50) DEFAULT 'PENDING',
    priority VARCHAR(50) DEFAULT 'MEDIUM',
    estimated_cost DECIMAL(10, 2),
    actual_cost DECIMAL(10, 2),
    materials JSONB,
    labor_hours DECIMAL(5, 2),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    photo_urls JSONB,
    signature TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_service_jobs_request_id ON service_jobs(request_id);
CREATE INDEX idx_service_jobs_provider_id ON service_jobs(provider_id);
CREATE INDEX idx_service_jobs_status ON service_jobs(status);
```

### service_categories table

```sql
CREATE TABLE service_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    icon TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    estimated_duration INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_service_categories_code ON service_categories(code);
```

## Use Cases

### Use Case 1: Customer Requests Home Service

```
1. Customer opens app, selects "Home Services"
2. Chooses service category (e.g., Cleaning)
3. Selects service type (e.g., Full House Cleaning)
4. Enters location and preferred date/time
5. System estimates cost based on category and location
6. Customer confirms and creates request
7. Request status set to PENDING
8. System notifies nearby available providers
```

### Use Case 2: Provider Accepts and Completes Job

```
1. Provider receives job notification
2. Reviews job details and location
3. Provider accepts job
4. System sets status to ACCEPTED
5. Customer receives provider details
6. Provider arrives and starts job (status: IN_PROGRESS)
7. Provider takes before/after photos
8. Provider completes job (status: COMPLETED)
9. Customer can now rate the service
```

### Use Case 3: Service Scheduling

```
1. Customer creates request for future date
2. System shows available time slots
3. Customer selects preferred time (e.g., 09:00-11:00)
4. Provider receives scheduling notification
5. Provider accepts and confirms time
6. Calendar event created for both parties
7. Reminder sent 24 hours before
8. Day of service, provider arrives in slot
```

### Use Case 4: Rating and Feedback

```
1. Service marked as COMPLETED
2. Customer receives rating prompt
3. Customer rates 1-5 stars
4. Optional written review
5. Can report issues if unsatisfied
6. Rating saved in database
7. Provider rating updated
8. Issue flagged for admin review if needed
```

## Common Operations

### Create Service Request

```go
handler := func(c *gin.Context) {
    userID := c.GetString("user_id")
    var req CreateRequestRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    request, err := s.homeServiceService.CreateRequest(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Notify nearby providers
    go s.notifyNearbyProviders(userID, request)

    c.JSON(http.StatusCreated, request)
}
```

### Find Available Providers

```go
handler := func(c *gin.Context) {
    requestID := c.Param("id")
    
    request, _ := s.homeServiceService.GetRequest(c.Request.Context(), requestID)
    
    searchReq := &ProviderSearchRequest{
        ServiceCategory: request.ServiceCategory,
        Latitude: request.Latitude,
        Longitude: request.Longitude,
        ScheduledDate: request.ScheduledDate,
    }

    providers, err := s.homeServiceService.FindAvailableProviders(c.Request.Context(), searchReq)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, providers)
}
```

### Complete Job with Photos

```go
handler := func(c *gin.Context) {
    jobID := c.Param("id")
    
    // Parse multipart form for photos
    if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid form data"})
        return
    }

    files := c.Request.MultipartForm.File["photos"]
    
    photoURLs, err := s.homeServiceService.UploadPhotos(c.Request.Context(), jobID, files)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Mark job as completed
    err = s.homeServiceService.CompleteJob(c.Request.Context(), jobID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Job completed",
        "photo_urls": photoURLs,
    })
}
```

### Rate Service

```go
handler := func(c *gin.Context) {
    jobID := c.Param("id")
    var req RateServiceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    err := s.homeServiceService.RateService(c.Request.Context(), jobID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, SuccessResponse{Message: "Rating submitted"})
}
```

## Error Handling

| Error | Status Code | Message |
|-------|------------|---------|
| Request not found | 404 | "Service request not found" |
| Invalid service category | 400 | "Invalid service category" |
| No providers available | 404 | "No providers available for this service" |
| Invalid time slot | 400 | "Selected time slot is not available" |
| Location invalid | 400 | "Invalid location coordinates" |
| Request cancelled | 409 | "Cannot modify cancelled request" |
| Job in progress | 409 | "Job is currently in progress" |
| Photo upload failed | 422 | "Failed to upload photos" |
| Invalid rating | 400 | "Rating must be between 1 and 5" |

## Performance Optimization

### Database Indexes
- Index on user_id for quick lookups
- Index on service_category for filtering
- Index on scheduled_date for date range queries
- Index on status for status-based filtering
- Geospatial index on latitude/longitude for nearby searches

### Caching Strategy
- Cache service categories (rarely change)
- Cache provider availability (hourly update)
- Cache recent requests (short TTL)

### Notification Optimization
- Batch provider notifications
- Use geohashing for radius searches
- Implement notification queue system

## Security Considerations

### Location Privacy
- Allow anonymous location sharing
- Don't expose address in notifications
- Implement location-based access controls

### Payment Security
- Validate costs before charging
- Implement dispute resolution
- Audit all transactions

### Data Protection
- Encrypt customer addresses
- Implement data retention policies
- Regular backups of service history

## Testing Strategy

### Unit Tests
- Request creation validation
- Provider availability calculation
- Cost estimation accuracy
- Rating validation

### Integration Tests
- Complete service request workflow
- Provider assignment and scheduling
- Photo upload and storage
- Rating and review process

## Integration Points

### With Service Providers Module
- Providers find available jobs
- Job assignment and tracking

### With Ratings Module
- Service ratings and reviews
- Provider reputation management

### With Wallet Module
- Cost calculation and payment
- Provider settlement

## Common Pitfalls

1. **Not validating time slots** - Ensure provider availability
2. **Missing location validation** - Invalid coordinates cause routing issues
3. **Incorrect cost estimation** - Should account for service type and location
4. **Not handling cancellations properly** - Track reason and refund logic
5. **Missing photo evidence** - Critical for dispute resolution
6. **Not updating job status** - Important for notifications

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
