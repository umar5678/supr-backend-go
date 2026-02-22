# Rides Module Development Guide

## Overview

The Rides Module is the core component of the ride-sharing platform. It manages the complete lifecycle of ride requests from creation through completion, including driver matching, real-time tracking, status management, and WebSocket-based live updates.

## Module Structure

```
rides/
├── handler.go         # HTTP request handlers and WebSocket
├── service.go         # Business logic and orchestration
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. Ride Request Management - Create and manage ride requests
2. Driver Matching - Assign available drivers to requests
3. Real-time Tracking - Handle live location updates via WebSocket
4. Status Management - Manage ride lifecycle states
5. Fare Calculation - Integrate with pricing module
6. Ride Completion - Process ride completion and settlement
7. WebSocket Communication - Real-time updates to clients

## Architecture

### Handler Layer (handler.go)

Manages HTTP endpoints and WebSocket connections.

Key methods:

```
CreateRide(c *gin.Context)                 // POST /rides
GetRide(c *gin.Context)                    // GET /rides/{id}
ListRides(c *gin.Context)                  // GET /rides
CancelRide(c *gin.Context)                 // POST /rides/{id}/cancel
CompleteRide(c *gin.Context)               // POST /rides/{id}/complete
UpdateRideStatus(c *gin.Context)           // PUT /rides/{id}/status
GetRideTracking(c *gin.Context)            // GET /rides/{id}/tracking
AcceptRide(c *gin.Context)                 // POST /rides/{id}/accept
WebSocketHandler(c *gin.Context)           // WS /rides/{id}/track
```

Request flow:
1. Extract parameters from URL/Query/Body
2. Validate request data
3. Call service method
4. Return response via response utility
5. WebSocket: Upgrade connection and handle live updates

### Service Layer (service.go)

Contains ride business logic and orchestration.

Key interface methods:

```
CreateRide(ctx context.Context, userID string, req CreateRideRequest) (*RideResponse, error)
GetRide(ctx context.Context, userID, rideID string) (*RideResponse, error)
ListRides(ctx context.Context, userID string, filters map[string]interface{}) ([]*RideResponse, error)
CancelRide(ctx context.Context, userID, rideID string) error
CompleteRide(ctx context.Context, rideID string, actualFare float64) error
AcceptRide(ctx context.Context, driverID, rideID string) error
UpdateRideStatus(ctx context.Context, rideID string, status RideStatus) error
GetRideTracking(ctx context.Context, rideID string) (*TrackingData, error)
UpdateDriverLocation(ctx context.Context, rideID, driverID string, lat, lon float64) error
CalculateETA(ctx context.Context, driverLat, driverLon, destLat, destLon float64) (int, error)
```

Logic flow:
1. Validate ride request details
2. Check driver availability
3. Calculate initial fare estimate using Pricing module
4. Perform driver matching algorithm
5. Update ride status through state machine
6. Handle concurrent requests safely
7. Trigger notifications through Messages module
8. Log all significant state transitions

### Repository Layer (repository.go)

Handles database operations for rides.

Key interface methods:

```
CreateRide(ctx context.Context, ride *models.Ride) error
FindRideByID(ctx context.Context, rideID string) (*models.Ride, error)
ListRidesByUser(ctx context.Context, userID string) ([]*models.Ride, error)
UpdateRideStatus(ctx context.Context, rideID string, status RideStatus) error
UpdateDriverAssignment(ctx context.Context, rideID, driverID string) error
SaveLocation(ctx context.Context, rideID string, lat, lon float64) error
GetLocationHistory(ctx context.Context, rideID string) ([]LocationPoint, error)
UpdateFare(ctx context.Context, rideID string, fare float64) error
CancelRide(ctx context.Context, rideID string) error
CompleteRide(ctx context.Context, rideID string, actualFare, actualDuration int) error
```

Database operations:
- Use transactions for critical multi-table updates
- Implement proper locking for concurrent drivers
- Store location history with timestamps
- Track all status transitions

## Data Transfer Objects

### CreateRideRequest

```go
type CreateRideRequest struct {
    PickupLocation   Location `json:"pickup_location" binding:"required"`
    DropoffLocation  Location `json:"dropoff_location" binding:"required"`
    RideType         string   `json:"ride_type" binding:"required"`
    ScheduledTime    *time.Time `json:"scheduled_time,omitempty"`
    PaymentMethod    string   `json:"payment_method" binding:"required"`
    SpecialRequests  string   `json:"special_requests,omitempty"`
}

type Location struct {
    Latitude  float64 `json:"latitude" binding:"required"`
    Longitude float64 `json:"longitude" binding:"required"`
    Address   string  `json:"address,omitempty"`
}
```

### RideResponse

```go
type RideResponse struct {
    ID              string           `json:"id"`
    RiderID         string           `json:"rider_id"`
    DriverID        *string          `json:"driver_id,omitempty"`
    PickupLocation  Location         `json:"pickup_location"`
    DropoffLocation Location         `json:"dropoff_location"`
    Status          string           `json:"status"`
    EstimatedFare   float64          `json:"estimated_fare"`
    ActualFare      *float64         `json:"actual_fare,omitempty"`
    Distance        *float64         `json:"distance,omitempty"`
    Duration        *int             `json:"duration,omitempty"` // seconds
    RideType        string           `json:"ride_type"`
    CreatedAt       time.Time        `json:"created_at"`
    UpdatedAt       time.Time        `json:"updated_at"`
    DriverInfo      *DriverInfo      `json:"driver_info,omitempty"`
    RideUpdates     []RideStatusUpdate `json:"ride_updates,omitempty"`
}

type DriverInfo struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Phone     string `json:"phone,omitempty"`
    Rating    float64 `json:"rating"`
    Vehicle   VehicleInfo `json:"vehicle"`
}
```

### TrackingData

```go
type TrackingData struct {
    RideID       string          `json:"ride_id"`
    DriverID     string          `json:"driver_id"`
    CurrentLat   float64         `json:"current_lat"`
    CurrentLon   float64         `json:"current_lon"`
    DestLat      float64         `json:"dest_lat"`
    DestLon      float64         `json:"dest_lon"`
    ETA          int             `json:"eta"` // seconds
    DistanceLeft float64         `json:"distance_left"` // meters
    UpdatedAt    time.Time       `json:"updated_at"`
}
```

## Ride Status State Machine

```
REQUESTED
    |
    v
ACCEPTED (driver assigned)
    |
    v
ARRIVED (driver at pickup)
    |
    v
STARTED (passenger boarded)
    |
    v
COMPLETED (ride finished)

Cancellation paths:
REQUESTED -> CANCELLED (by rider)
ACCEPTED -> CANCELLED (by driver or rider)
ARRIVED -> CANCELLED (by rider)
```

## Typical Use Cases

### 1. Create Ride Request

Request:
```
POST /rides
{
    "pickup_location": {
        "latitude": 40.7128,
        "longitude": -74.0060,
        "address": "123 Main St, NYC"
    },
    "dropoff_location": {
        "latitude": 40.7580,
        "longitude": -73.9855,
        "address": "Times Square, NYC"
    },
    "ride_type": "economy",
    "payment_method": "wallet",
    "special_requests": "Non-smoking please"
}
```

Flow:
1. Extract rider ID from JWT token
2. Validate pickup and dropoff locations
3. Call pricing module to get fare estimate
4. Create ride record with status REQUESTED
5. Trigger driver matching in background
6. Return ride response with estimated fare

### 2. Driver Accepts Ride

Request:
```
POST /rides/{rideID}/accept
{
    "driver_id": "driver-123"
}
```

Flow:
1. Extract driver ID from JWT token
2. Find ride by ID
3. Verify ride status is REQUESTED
4. Check driver is available
5. Update ride with driver assignment
6. Change status to ACCEPTED
7. Notify rider via Messages module
8. Start tracking driver location

### 3. Real-time Location Tracking

WebSocket Connection:
```
WS /rides/{rideID}/track
```

Flow:
1. Upgrade HTTP connection to WebSocket
2. Authenticate user (rider or driver)
3. Send initial tracking data
4. Receive driver location updates
5. Calculate ETA on each update
6. Broadcast updates to connected clients
7. Clean up connection on close

Location Update Message:
```json
{
    "type": "location_update",
    "driver_id": "driver-123",
    "latitude": 40.7200,
    "longitude": -73.9900,
    "eta": 420,
    "distance_left": 1500,
    "timestamp": "2024-01-20T10:30:00Z"
}
```

### 4. Complete Ride

Request:
```
POST /rides/{rideID}/complete
{
    "actual_fare": 25.50,
    "actual_duration": 1200
}
```

Flow:
1. Find ride by ID
2. Calculate actual fare if not provided
3. Verify ride status is STARTED
4. Update ride with actual duration and fare
5. Change status to COMPLETED
6. Trigger wallet payment
7. Create ride record for history
8. Notify both parties

### 5. Cancel Ride

Request:
```
POST /rides/{rideID}/cancel
{
    "reason": "Driver is taking too long"
}
```

Flow:
1. Find ride by ID
2. Verify ride can be cancelled (check status)
3. Determine cancellation charges based on status
4. Update ride status to CANCELLED
5. Process refund if applicable
6. Notify all parties
7. Log cancellation reason

## WebSocket Message Types

### From Client to Server

```json
{
    "type": "location_update",
    "latitude": 40.7200,
    "longitude": -73.9900
}

{
    "type": "ride_update",
    "status": "arrived"
}

{
    "type": "ping"
}
```

### From Server to Client

```json
{
    "type": "driver_location_updated",
    "data": { ... }
}

{
    "type": "ride_status_changed",
    "status": "arrived"
}

{
    "type": "eta_updated",
    "eta_seconds": 420
}

{
    "type": "pong"
}
```

## Driver Matching Algorithm

The matching algorithm considers:

1. Proximity - Closest available driver
2. Acceptance Rate - Drivers with higher acceptance rates
3. Rating - Higher-rated drivers preferred
4. Current Load - Drivers with fewer concurrent rides
5. Vehicle Type - Matching requested ride type
6. Surge Pricing - Increase rates in high-demand areas

```
score = (1 / distance) * acceptanceRate * rating * (1 / currentLoad)
```

## Error Handling

Common error scenarios:

1. Ride Not Found
   - Response: 404 Not Found
   - Action: Return error response

2. Invalid Status Transition
   - Response: 400 Bad Request
   - Action: Log attempt, return error

3. Rider Already Has Active Ride
   - Response: 409 Conflict
   - Action: Return existing ride

4. No Available Drivers
   - Response: 202 Accepted
   - Action: Queue request, wait for driver

5. WebSocket Connection Lost
   - Action: Save last known state
   - Action: Attempt reconnection client-side

## Testing Strategy

### Unit Tests (Service Layer)

```go
Test_CreateRide_Success()
Test_CreateRide_InvalidLocation()
Test_AcceptRide_DriverAvailable()
Test_UpdateStatus_InvalidTransition()
Test_CancelRide_RefundCalculation()
Test_CalculateETA()
```

### Integration Tests (Repository Layer)

```go
Test_CreateRide_SaveToDatabase()
Test_FindRide_Retrieval()
Test_UpdateRideStatus_Persistence()
Test_SaveLocation_History()
```

### End-to-End Tests (Handler Layer)

```go
Test_CreateRide_FullFlow()
Test_DriverAcceptance_FullFlow()
Test_RideCompletion_FullFlow()
Test_WebSocket_LocationTracking()
```

## Database Schema

### Rides Table

```sql
CREATE TABLE rides (
    id VARCHAR(36) PRIMARY KEY,
    rider_id VARCHAR(36) NOT NULL,
    driver_id VARCHAR(36),
    pickup_location JSON,
    dropoff_location JSON,
    status VARCHAR(50),
    estimated_fare DECIMAL(10, 2),
    actual_fare DECIMAL(10, 2),
    ride_type VARCHAR(50),
    distance DECIMAL(10, 2),
    duration INT,
    payment_method VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (rider_id) REFERENCES users(id),
    FOREIGN KEY (driver_id) REFERENCES users(id)
);
```

### Ride Locations Table

```sql
CREATE TABLE ride_locations (
    id VARCHAR(36) PRIMARY KEY,
    ride_id VARCHAR(36) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    timestamp TIMESTAMP,
    FOREIGN KEY (ride_id) REFERENCES rides(id)
);
```

## Integration Points

1. Pricing Module - For fare estimation and calculation
2. Tracking Module - For GPS and ETA data
3. Wallet Module - For payment processing
4. Messages Module - For notifications
5. Ratings Module - For post-ride reviews
6. Drivers Module - For driver information
7. Riders Module - For rider information

## Performance Optimization

1. Use database indexes on frequently queried fields
2. Implement location caching for tracking
3. Batch location updates every 2-3 seconds
4. Use Redis for WebSocket session management
5. Implement ride queue with priority
6. Cache ETA calculations

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture
- See PRICING-MODULE.md for fare calculations
- See TRACKING-MODULE.md for location handling
- See internal/websocket for WebSocket implementation

## Common Pitfalls

1. Race conditions in driver assignment
2. Not handling WebSocket disconnections
3. Inefficient location query patterns
4. Missing transaction for multi-step operations
5. Not validating status transitions
6. Insufficient error logging
7. Memory leaks in WebSocket handlers

## Future Enhancements

1. AI-based driver matching
2. Predictive ETA using machine learning
3. Ride pooling support
4. Scheduled rides
5. Recurring rides
6. Multiple stops support
7. Ride insurance options
8. Accessibility features
