# Vehicles Module Documentation

## Overview

The Vehicles Module manages vehicle registration, information, documentation, and driver-vehicle associations. It handles vehicle types, registration details, insurance information, and links vehicles to drivers.

## Key Responsibilities

- Vehicle registration and profile management
- Vehicle type and category management
- Vehicle documentation (registration, insurance, emissions)
- Driver-vehicle associations
- Vehicle verification and approval workflow
- Vehicle status tracking (active, inactive, suspended)
- Insurance and registration expiry tracking

## Architecture

### Handler Layer (`vehicles/handler.go`)

Handles HTTP requests related to vehicle operations.

**Key Endpoints:**

```go
POST /api/v1/vehicles/register           // Register a new vehicle
GET /api/v1/vehicles/:id                 // Get vehicle details
PUT /api/v1/vehicles/:id                 // Update vehicle information
GET /api/v1/vehicles                     // List vehicles (with filters)
POST /api/v1/vehicles/:id/documents      // Upload vehicle documents
GET /api/v1/vehicles/:id/documents       // Get vehicle documents
PUT /api/v1/vehicles/:id/verify          // Verify vehicle documents
DELETE /api/v1/vehicles/:id              // Delete/deactivate vehicle
POST /api/v1/drivers/:driver_id/vehicles // Assign vehicle to driver
GET /api/v1/drivers/:driver_id/vehicles  // Get driver's vehicles
```

### Service Layer (`vehicles/service.go`)

Contains business logic for vehicle operations.

**Key Methods:**

```go
func (s *VehicleService) RegisterVehicle(ctx context.Context, req *RegisterVehicleRequest) (*VehicleResponse, error)
func (s *VehicleService) GetVehicle(ctx context.Context, vehicleID string) (*VehicleResponse, error)
func (s *VehicleService) UpdateVehicle(ctx context.Context, vehicleID string, req *UpdateVehicleRequest) (*VehicleResponse, error)
func (s *VehicleService) ListVehicles(ctx context.Context, filters *VehicleFilters) ([]*VehicleResponse, error)
func (s *VehicleService) UploadDocuments(ctx context.Context, vehicleID string, docs *VehicleDocuments) error
func (s *VehicleService) VerifyVehicle(ctx context.Context, vehicleID string, status string) error
func (s *VehicleService) AssignToDriver(ctx context.Context, vehicleID, driverID string) error
func (s *VehicleService) UnassignFromDriver(ctx context.Context, vehicleID, driverID string) error
func (s *VehicleService) GetDriverVehicles(ctx context.Context, driverID string) ([]*VehicleResponse, error)
```

### Repository Layer (`vehicles/repository.go`)

Manages database operations for vehicles.

**Key Methods:**

```go
func (r *VehicleRepository) Create(ctx context.Context, vehicle *Vehicle) error
func (r *VehicleRepository) GetByID(ctx context.Context, vehicleID string) (*Vehicle, error)
func (r *VehicleRepository) Update(ctx context.Context, vehicleID string, updates map[string]interface{}) error
func (r *VehicleRepository) Delete(ctx context.Context, vehicleID string) error
func (r *VehicleRepository) FindByDriverID(ctx context.Context, driverID string) ([]*Vehicle, error)
func (r *VehicleRepository) FindByPlateNumber(ctx context.Context, plateNumber string) (*Vehicle, error)
func (r *VehicleRepository) List(ctx context.Context, filters *VehicleFilters) ([]*Vehicle, error)
func (r *VehicleRepository) UpdateStatus(ctx context.Context, vehicleID string, status string) error
```

## Data Models

### Vehicle

```go
type Vehicle struct {
    ID                  string    `db:"id" json:"id"`
    DriverID            string    `db:"driver_id" json:"driver_id"`
    VehicleType         string    `db:"vehicle_type" json:"vehicle_type"`              // Car, Bike, Truck, etc.
    Make                string    `db:"make" json:"make"`                             // Manufacturer
    Model               string    `db:"model" json:"model"`
    Year                int       `db:"year" json:"year"`
    Color               string    `db:"color" json:"color"`
    PlateNumber         string    `db:"plate_number" json:"plate_number"`             // Unique registration number
    VIN                 string    `db:"vin" json:"vin"`                               // Vehicle Identification Number
    Capacity            int       `db:"capacity" json:"capacity"`                     // Number of passengers
    FuelType            string    `db:"fuel_type" json:"fuel_type"`                   // Petrol, Diesel, Electric, Hybrid
    TransmissionType    string    `db:"transmission_type" json:"transmission_type"`   // Manual, Automatic
    
    // Documentation
    RegistrationNumber  string    `db:"registration_number" json:"registration_number"`
    RegistrationExpiry  time.Time `db:"registration_expiry" json:"registration_expiry"`
    InsuranceProvider   string    `db:"insurance_provider" json:"insurance_provider"`
    InsurancePolicyNo   string    `db:"insurance_policy_no" json:"insurance_policy_no"`
    InsuranceExpiry     time.Time `db:"insurance_expiry" json:"insurance_expiry"`
    EmissionsTest       string    `db:"emissions_test" json:"emissions_test"`
    EmissionsExpiry     time.Time `db:"emissions_expiry" json:"emissions_expiry"`
    
    // Status
    Status              string    `db:"status" json:"status"`                         // ACTIVE, INACTIVE, SUSPENDED, PENDING_VERIFICATION
    VerificationStatus  string    `db:"verification_status" json:"verification_status"` // PENDING, APPROVED, REJECTED
    IsActive            bool      `db:"is_active" json:"is_active"`
    
    // Metadata
    CreatedAt           time.Time `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
    ApprovedAt          *time.Time `db:"approved_at" json:"approved_at"`
    ApprovedBy          *string   `db:"approved_by" json:"approved_by"`               // Admin ID
    
    // Tracking
    LastInspectionDate  *time.Time `db:"last_inspection_date" json:"last_inspection_date"`
    NextInspectionDue   *time.Time `db:"next_inspection_due" json:"next_inspection_due"`
}
```

### VehicleDocuments

```go
type VehicleDocuments struct {
    ID                  string    `db:"id" json:"id"`
    VehicleID           string    `db:"vehicle_id" json:"vehicle_id"`
    RegistrationDoc     string    `db:"registration_doc" json:"registration_doc"`     // S3/Cloud URL
    InsuranceDoc        string    `db:"insurance_doc" json:"insurance_doc"`
    EmissionsDoc        string    `db:"emissions_doc" json:"emissions_doc"`
    MOTCertificate      string    `db:"mot_certificate" json:"mot_certificate"`
    PollutionCert       string    `db:"pollution_cert" json:"pollution_cert"`
    OtherDocs           string    `db:"other_docs" json:"other_docs"`                 // JSON array of URLs
    DocumentsVerified   bool      `db:"documents_verified" json:"documents_verified"`
    VerifiedAt          *time.Time `db:"verified_at" json:"verified_at"`
    VerifiedBy          *string   `db:"verified_by" json:"verified_by"`               // Admin ID
    VerificationNotes   *string   `db:"verification_notes" json:"verification_notes"`
    CreatedAt           time.Time `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}
```

### VehicleType

```go
type VehicleType struct {
    ID                  string    `db:"id" json:"id"`
    Name                string    `db:"name" json:"name"`                             // Car, Motorcycle, Truck
    Code                string    `db:"code" json:"code"`                             // CAR, BIKE, TRUCK
    Description         string    `db:"description" json:"description"`
    MaxCapacity         int       `db:"max_capacity" json:"max_capacity"`
    MinCapacity         int       `db:"min_capacity" json:"min_capacity"`
    Category            string    `db:"category" json:"category"`                     // Economy, Premium, XL
    IsActive            bool      `db:"is_active" json:"is_active"`
    CreatedAt           time.Time `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}
```

## DTOs (Data Transfer Objects)

### RegisterVehicleRequest

```go
type RegisterVehicleRequest struct {
    VehicleType         string `json:"vehicle_type" binding:"required"`
    Make                string `json:"make" binding:"required"`
    Model               string `json:"model" binding:"required"`
    Year                int    `json:"year" binding:"required"`
    Color               string `json:"color" binding:"required"`
    PlateNumber         string `json:"plate_number" binding:"required"`
    VIN                 string `json:"vin" binding:"required"`
    Capacity            int    `json:"capacity" binding:"required"`
    FuelType            string `json:"fuel_type" binding:"required"`
    TransmissionType    string `json:"transmission_type"`
    RegistrationNumber  string `json:"registration_number" binding:"required"`
    RegistrationExpiry  string `json:"registration_expiry" binding:"required"`
    InsuranceProvider   string `json:"insurance_provider" binding:"required"`
    InsurancePolicyNo   string `json:"insurance_policy_no" binding:"required"`
    InsuranceExpiry     string `json:"insurance_expiry" binding:"required"`
}
```

### UpdateVehicleRequest

```go
type UpdateVehicleRequest struct {
    Make                *string `json:"make"`
    Model               *string `json:"model"`
    Color               *string `json:"color"`
    Capacity            *int    `json:"capacity"`
    FuelType            *string `json:"fuel_type"`
    TransmissionType    *string `json:"transmission_type"`
    RegistrationExpiry  *string `json:"registration_expiry"`
    InsuranceProvider   *string `json:"insurance_provider"`
    InsurancePolicyNo   *string `json:"insurance_policy_no"`
    InsuranceExpiry     *string `json:"insurance_expiry"`
}
```

### VehicleResponse

```go
type VehicleResponse struct {
    ID                  string    `json:"id"`
    DriverID            string    `json:"driver_id"`
    VehicleType         string    `json:"vehicle_type"`
    Make                string    `json:"make"`
    Model               string    `json:"model"`
    Year                int       `json:"year"`
    Color               string    `json:"color"`
    PlateNumber         string    `json:"plate_number"`
    Capacity            int       `json:"capacity"`
    Status              string    `json:"status"`
    VerificationStatus  string    `json:"verification_status"`
    RegistrationExpiry  time.Time `json:"registration_expiry"`
    InsuranceExpiry     time.Time `json:"insurance_expiry"`
    Documents           *VehicleDocuments `json:"documents,omitempty"`
    CreatedAt           time.Time `json:"created_at"`
}
```

### VehicleFilters

```go
type VehicleFilters struct {
    DriverID            *string
    VehicleType         *string
    Status              *string
    VerificationStatus  *string
    PlateNumber         *string
    Page                int
    Limit               int
    SortBy              string    // created_at, updated_at, plate_number
    SortOrder           string    // asc, desc
}
```

## Database Schema

### vehicles table

```sql
CREATE TABLE vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    vehicle_type VARCHAR(50) NOT NULL,
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    color VARCHAR(50),
    plate_number VARCHAR(20) UNIQUE NOT NULL,
    vin VARCHAR(17) UNIQUE NOT NULL,
    capacity INTEGER NOT NULL,
    fuel_type VARCHAR(50),
    transmission_type VARCHAR(50),
    registration_number VARCHAR(100) UNIQUE NOT NULL,
    registration_expiry TIMESTAMP NOT NULL,
    insurance_provider VARCHAR(100),
    insurance_policy_no VARCHAR(100) UNIQUE,
    insurance_expiry TIMESTAMP NOT NULL,
    emissions_test VARCHAR(50),
    emissions_expiry TIMESTAMP,
    status VARCHAR(50) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED', 'PENDING_VERIFICATION')),
    verification_status VARCHAR(50) DEFAULT 'PENDING' CHECK (verification_status IN ('PENDING', 'APPROVED', 'REJECTED')),
    is_active BOOLEAN DEFAULT TRUE,
    approved_at TIMESTAMP,
    approved_by UUID REFERENCES admins(id),
    last_inspection_date TIMESTAMP,
    next_inspection_due TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_expiry CHECK (insurance_expiry > registration_expiry)
);

CREATE INDEX idx_vehicles_driver_id ON vehicles(driver_id);
CREATE INDEX idx_vehicles_plate_number ON vehicles(plate_number);
CREATE INDEX idx_vehicles_status ON vehicles(status);
CREATE INDEX idx_vehicles_verification_status ON vehicles(verification_status);
CREATE INDEX idx_vehicles_created_at ON vehicles(created_at);
```

### vehicle_documents table

```sql
CREATE TABLE vehicle_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    registration_doc TEXT,
    insurance_doc TEXT,
    emissions_doc TEXT,
    mot_certificate TEXT,
    pollution_cert TEXT,
    other_docs JSONB,
    documents_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    verified_by UUID REFERENCES admins(id),
    verification_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vehicle_documents_vehicle_id ON vehicle_documents(vehicle_id);
CREATE INDEX idx_vehicle_documents_verified ON vehicle_documents(documents_verified);
```

### vehicle_types table

```sql
CREATE TABLE vehicle_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL,
    description TEXT,
    max_capacity INTEGER NOT NULL,
    min_capacity INTEGER NOT NULL,
    category VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vehicle_types_code ON vehicle_types(code);
CREATE INDEX idx_vehicle_types_category ON vehicle_types(category);
```

## Use Cases

### Use Case 1: Driver Registers a Vehicle

```
1. Driver submits RegisterVehicleRequest with vehicle details
2. System validates plate number uniqueness
3. System validates VIN format and uniqueness
4. System validates insurance expiry > registration expiry
5. Vehicle created with PENDING_VERIFICATION status
6. Documents uploaded and verified by admin
7. Vehicle status changes to ACTIVE
8. Driver can use vehicle for rides
```

### Use Case 2: Vehicle Documentation Verification

```
1. Admin receives document upload notification
2. Admin reviews all required documents
3. Admin verifies authenticity and validity
4. Admin sets verification_status to APPROVED/REJECTED
5. If APPROVED: Vehicle status becomes ACTIVE
6. If REJECTED: Driver notified with rejection reasons
7. Driver can upload new documents for re-verification
```

### Use Case 3: Vehicle Insurance Expiry Alert

```
1. System runs daily check for expiring insurance (30 days before)
2. System sends notification to driver
3. Driver updates insurance_expiry date
4. System verifies new document
5. If expired: Vehicle becomes SUSPENDED
6. Driver cannot create new rides until updated
```

### Use Case 4: Assign Vehicle to Driver

```
1. Driver selects vehicle to use
2. System validates vehicle status (must be ACTIVE)
3. System validates vehicle ownership
4. Assigns vehicle to driver profile
5. System updates driver's active vehicle
6. Driver can now create rides with this vehicle
```

## Common Operations

### Register a Vehicle

```go
handler := func(c *gin.Context) {
    var req RegisterVehicleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    driverID := c.GetString("user_id")
    
    vehicle, err := s.vehicleService.RegisterVehicle(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusCreated, vehicle)
}
```

### Get Vehicle Details

```go
handler := func(c *gin.Context) {
    vehicleID := c.Param("id")
    
    vehicle, err := s.vehicleService.GetVehicle(c.Request.Context(), vehicleID)
    if err != nil {
        c.JSON(http.StatusNotFound, ErrorResponse{Error: "Vehicle not found"})
        return
    }

    c.JSON(http.StatusOK, vehicle)
}
```

### Upload Vehicle Documents

```go
handler := func(c *gin.Context) {
    vehicleID := c.Param("id")
    
    // Parse multipart form (10MB max)
    if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid form data"})
        return
    }

    documents := &VehicleDocuments{
        VehicleID: vehicleID,
    }

    // Upload files to S3/Cloud storage
    if file, header, err := c.FormFile("registration_doc"); err == nil {
        documents.RegistrationDoc = uploadToCloudStorage(file, header)
    }
    
    if file, header, err := c.FormFile("insurance_doc"); err == nil {
        documents.InsuranceDoc = uploadToCloudStorage(file, header)
    }

    if err := s.vehicleService.UploadDocuments(c.Request.Context(), vehicleID, documents); err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, SuccessResponse{Message: "Documents uploaded"})
}
```

### List Vehicles with Filters

```go
handler := func(c *gin.Context) {
    driverID := c.GetString("user_id")
    
    filters := &VehicleFilters{
        DriverID: &driverID,
        Status: c.Query("status"),
        Page: getQueryInt(c, "page", 1),
        Limit: getQueryInt(c, "limit", 10),
    }

    vehicles, err := s.vehicleService.ListVehicles(c.Request.Context(), filters)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, PaginatedResponse{
        Data: vehicles,
        Total: len(vehicles),
        Page: filters.Page,
        Limit: filters.Limit,
    })
}
```

## Error Handling

### Common Errors

| Error | Status Code | Message |
|-------|------------|---------|
| Vehicle not found | 404 | "Vehicle with ID {id} not found" |
| Plate number exists | 409 | "Vehicle with plate number {plate} already exists" |
| Invalid VIN | 400 | "Invalid VIN format" |
| Insurance expired | 400 | "Insurance has expired" |
| Invalid expiry dates | 400 | "Insurance expiry must be after registration expiry" |
| Unauthorized | 403 | "You don't have permission to modify this vehicle" |
| Missing documents | 400 | "All required documents must be uploaded" |
| Document verification failed | 422 | "Document verification failed: {reason}" |
| Vehicle suspended | 409 | "Vehicle is suspended and cannot be used" |
| Capacity invalid | 400 | "Capacity must match vehicle type limits" |

### Error Response Format

```go
type ErrorResponse struct {
    Error      string      `json:"error"`
    Details    string      `json:"details,omitempty"`
    StatusCode int         `json:"status_code"`
    Timestamp  time.Time   `json:"timestamp"`
}
```

## Performance Optimization

### Database Indexes
- Index on `driver_id` for quick driver vehicle lookup
- Index on `plate_number` for unique constraint enforcement
- Index on `status` for status-based filtering
- Index on `verification_status` for admin reviews
- Index on `created_at` for chronological queries

### Caching Strategy
- Cache vehicle types (rarely change)
- Cache active vehicle count per driver
- Cache vehicle details with 1-hour TTL
- Invalidate cache on updates

### Query Optimization
- Use pagination for list endpoints
- Load documents only when requested
- Use database projections to limit fields
- Batch document uploads with async processing

## Security Considerations

### Input Validation
- Validate plate number format (alphanumeric, specific length)
- Validate VIN length (17 characters) and format
- Validate all dates are in future
- Validate capacity is positive integer

### Authorization
- Only vehicle owner/assigned driver can view details
- Only admins can verify documents
- Only vehicle owner can update vehicle
- Only admins can suspend vehicles

### Data Protection
- Encrypt sensitive documents at rest
- Use secure cloud storage (S3, GCS) for documents
- Implement document expiry alerts
- Audit all verification actions

### Document Security
- Scan documents for malware
- Verify document authenticity with third-party APIs
- Store document metadata separately
- Implement document retention policies

## Testing Strategy

### Unit Tests

```go
func TestRegisterVehicle(t *testing.T) {
    // Test valid registration
    // Test duplicate plate number
    // Test invalid VIN
    // Test expiry date validation
    // Test capacity validation
}

func TestUpdateVehicle(t *testing.T) {
    // Test successful update
    // Test unauthorized update
    // Test invalid expiry dates
}

func TestVerifyDocuments(t *testing.T) {
    // Test successful verification
    // Test missing documents
    // Test invalid documents
}
```

### Integration Tests

```go
func TestVehicleWorkflow(t *testing.T) {
    // 1. Register vehicle
    // 2. Upload documents
    // 3. Verify documents (admin)
    // 4. Assign to driver
    // 5. Check vehicle can be used
}

func TestDocumentExpiry(t *testing.T) {
    // 1. Create vehicle with expiring insurance
    // 2. Wait for expiry alert
    // 3. Verify vehicle suspended
    // 4. Update insurance
    // 5. Verify vehicle active
}
```

## Integration Points

### With Drivers Module
- Vehicle linked to driver via `driver_id`
- Driver can have multiple vehicles
- Driver selects active vehicle for rides

### With Rides Module
- Vehicle used in ride details
- Vehicle information shown to passengers
- Vehicle rating/reviews

### With Admin Module
- Admins verify vehicle documents
- Admins suspend/reactivate vehicles
- Admins view vehicle statistics

## Common Pitfalls

1. **Not validating expiry dates** - Insurance must be after registration
2. **Missing document verification** - Documents must be verified before vehicle can be used
3. **Not handling status transitions** - Vehicle status affects ride eligibility
4. **Duplicate plate numbers** - Should be globally unique
5. **Not setting inspection reminders** - Track maintenance schedules
6. **Ignoring capacity limits** - Validate against vehicle type
7. **Not auditing verification** - Track who verified and when

## Typical Use Cases and Examples

### Scenario 1: New Driver Vehicle Registration

```
POST /api/v1/vehicles/register
{
    "vehicle_type": "Car",
    "make": "Toyota",
    "model": "Camry",
    "year": 2023,
    "color": "Silver",
    "plate_number": "ABC-123",
    "vin": "12345678901234567",
    "capacity": 4,
    "fuel_type": "Petrol",
    "transmission_type": "Automatic",
    "registration_number": "REG-2023-001",
    "registration_expiry": "2025-12-31",
    "insurance_provider": "XYZ Insurance",
    "insurance_policy_no": "POL-123456",
    "insurance_expiry": "2026-12-31"
}

Response: 201 Created
{
    "id": "vehicle-uuid",
    "driver_id": "driver-uuid",
    "vehicle_type": "Car",
    "status": "PENDING_VERIFICATION",
    "verification_status": "PENDING",
    "created_at": "2026-02-22T10:00:00Z"
}
```

### Scenario 2: Admin Verifying Documents

```
PUT /api/v1/vehicles/{id}/verify
{
    "status": "APPROVED",
    "verification_notes": "All documents verified and valid"
}

Response: 200 OK
{
    "id": "vehicle-uuid",
    "status": "ACTIVE",
    "verification_status": "APPROVED",
    "approved_at": "2026-02-22T10:30:00Z",
    "approved_by": "admin-uuid"
}
```

### Scenario 3: Update Vehicle Insurance

```
PUT /api/v1/vehicles/{id}
{
    "insurance_expiry": "2027-12-31",
    "insurance_provider": "New Insurance Co"
}

Response: 200 OK
{
    "id": "vehicle-uuid",
    "insurance_expiry": "2027-12-31",
    "insurance_provider": "New Insurance Co",
    "updated_at": "2026-02-22T11:00:00Z"
}
```

## Deployment Checklist

- [ ] All indexes created on production database
- [ ] Document storage configured (S3/GCS)
- [ ] Document scanning service deployed
- [ ] Verification alerts configured
- [ ] Expiry check scheduled daily
- [ ] Backup strategy for documents configured
- [ ] Document retention policy implemented
- [ ] Admin verification workflow configured
- [ ] Load testing for document upload completed
- [ ] Security scan of document endpoints completed

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
