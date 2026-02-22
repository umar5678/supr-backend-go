# Tracking Module Development Guide

## Overview

The Tracking Module provides real-time location tracking capabilities for rides and services. It handles GPS data collection, location history storage, distance calculations, and ETA estimations using WebSocket connections and external mapping services.

## Module Structure

```
tracking/
├── handler.go         # HTTP request handlers
├── service.go         # Business logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. Real-time Location Updates - Process GPS updates via WebSocket
2. Location History - Store and retrieve location history
3. Distance Calculation - Calculate trip distances
4. ETA Estimation - Estimate arrival times
5. Geofencing - Handle zone-based logic
6. Route Optimization - Calculate optimal routes

## Architecture

### Handler Layer (handler.go)

Manages HTTP endpoints and WebSocket connections.

Key methods:

```
UpdateLocation(c *gin.Context)              // POST /tracking/location
GetCurrentLocation(c *gin.Context)          // GET /tracking/location/{rideID}
GetLocationHistory(c *gin.Context)          // GET /tracking/history/{rideID}
EstimateArrival(c *gin.Context)             // POST /tracking/eta
WebSocketHandler(c *gin.Context)            // WS /tracking/ws/{rideID}
GetRoute(c *gin.Context)                    // POST /tracking/route
CheckGeofence(c *gin.Context)               // POST /tracking/geofence-check
```

Request flow for REST:
1. Extract parameters from request
2. Validate location data
3. Call service method
4. Return response

Request flow for WebSocket:
1. Upgrade HTTP to WebSocket
2. Authenticate connection
3. Subscribe to location updates
4. Send real-time updates to clients
5. Handle disconnection cleanup

### Service Layer (service.go)

Contains tracking business logic.

Key interface methods:

```
UpdateLocation(ctx context.Context, rideID, userID string, lat, lon float64) error
GetCurrentLocation(ctx context.Context, rideID string) (*LocationResponse, error)
GetLocationHistory(ctx context.Context, rideID string, limit int) ([]LocationPoint, error)
EstimateArrival(ctx context.Context, fromLat, fromLon, toLat, toLon float64) (*ETAResponse, error)
CalculateDistance(ctx context.Context, fromLat, fromLon, toLat, toLon float64) (float64, error)
GetRoute(ctx context.Context, fromLat, fromLon, toLat, toLon float64) (*RouteResponse, error)
CheckGeofence(ctx context.Context, lat, lon float64, zones []GeofenceZone) ([]string, error)
BroadcastLocationUpdate(ctx context.Context, rideID string, location LocationData)
```

Logic flow:
1. Validate location coordinates
2. Store location in database
3. Calculate distance if multiple points
4. Calculate ETA using mapping service
5. Check geofence boundaries
6. Broadcast updates to connected WebSocket clients
7. Log location updates for audit trail

### Repository Layer (repository.go)

Handles database operations for tracking data.

Key interface methods:

```
SaveLocation(ctx context.Context, rideID string, lat, lon float64) error
GetLatestLocation(ctx context.Context, rideID string) (*Location, error)
GetLocationHistory(ctx context.Context, rideID string, limit int) ([]Location, error)
SaveRoute(ctx context.Context, rideID string, route RouteData) error
GetRoute(ctx context.Context, rideID string) (RouteData, error)
ClearOldLocations(ctx context.Context, retentionDays int) error
```

Database operations:
- Use efficient time-series storage
- Implement proper indexing on ride_id and timestamp
- Store location history with retention policy
- Use partitioning for large datasets

## Data Transfer Objects

### LocationData

```go
type LocationData struct {
    Latitude      float64   `json:"latitude" binding:"required,min=-90,max=90"`
    Longitude     float64   `json:"longitude" binding:"required,min=-180,max=180"`
    Accuracy      float64   `json:"accuracy,omitempty"` // meters
    Altitude      float64   `json:"altitude,omitempty"` // meters
    Speed         float64   `json:"speed,omitempty"`    // m/s
    Heading       float64   `json:"heading,omitempty"`  // degrees (0-360)
    Timestamp     time.Time `json:"timestamp,omitempty"`
}
```

### LocationResponse

```go
type LocationResponse struct {
    RideID        string        `json:"ride_id"`
    DriverID      string        `json:"driver_id"`
    CurrentLat    float64       `json:"current_lat"`
    CurrentLon    float64       `json:"current_lon"`
    Accuracy      float64       `json:"accuracy,omitempty"`
    DestinationLat float64      `json:"destination_lat"`
    DestinationLon float64      `json:"destination_lon"`
    ETA           int           `json:"eta"` // seconds
    DistanceLeft  float64       `json:"distance_left"` // meters
    TotalDistance float64       `json:"total_distance"` // meters
    TraveledDistance float64    `json:"traveled_distance"` // meters
    UpdatedAt     time.Time     `json:"updated_at"`
}
```

### ETAResponse

```go
type ETAResponse struct {
    Destination   Location `json:"destination"`
    ETA           int      `json:"eta"` // seconds
    Distance      float64  `json:"distance"` // meters
    Traffic       string   `json:"traffic"` // light, moderate, heavy
    AlternateRoutes []RouteOption `json:"alternate_routes,omitempty"`
}

type RouteOption struct {
    Distance float64 `json:"distance"` // meters
    Duration int     `json:"duration"` // seconds
    Traffic  string  `json:"traffic"`
}
```

### RouteResponse

```go
type RouteResponse struct {
    StartPoint     Location  `json:"start_point"`
    EndPoint       Location  `json:"end_point"`
    Distance       float64   `json:"distance"` // kilometers
    Duration       int       `json:"duration"` // seconds
    Polyline       string    `json:"polyline"` // Encoded polyline
    Instructions   []string  `json:"instructions,omitempty"`
    Waypoints      []Location `json:"waypoints,omitempty"`
}
```

### LocationHistory

```go
type LocationHistory struct {
    RideID         string           `json:"ride_id"`
    Locations      []LocationPoint  `json:"locations"`
    TotalDistance  float64          `json:"total_distance"`
    TotalDuration  int              `json:"total_duration"`
    AverageSpeed   float64          `json:"average_speed"`
}

type LocationPoint struct {
    Latitude      float64   `json:"latitude"`
    Longitude     float64   `json:"longitude"`
    Accuracy      float64   `json:"accuracy,omitempty"`
    Speed         float64   `json:"speed,omitempty"`
    Timestamp     time.Time `json:"timestamp"`
}
```

## Typical Use Cases

### 1. Update Driver Location

Request (REST):
```
POST /tracking/location
{
    "ride_id": "ride-123",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "accuracy": 10.5,
    "speed": 15.2,
    "heading": 270.0
}
```

Flow:
1. Validate coordinates
2. Save location to database
3. Calculate distance from previous point
4. Calculate speed if multiple points
5. Check if arrived at destination
6. Broadcast update to connected WebSocket clients
7. Return success response

### 2. WebSocket Location Streaming

Connection:
```
WS /tracking/ws/ride-123
```

Client sends (every 2-3 seconds):
```json
{
    "type": "location_update",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "accuracy": 10.5,
    "timestamp": "2024-02-20T10:30:00Z"
}
```

Server broadcasts to connected clients:
```json
{
    "type": "driver_location_updated",
    "driver_id": "driver-123",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "eta_seconds": 420,
    "distance_left": 1500,
    "speed": 15.2,
    "timestamp": "2024-02-20T10:30:00Z"
}
```

### 3. Get ETA Estimate

Request:
```
POST /tracking/eta
{
    "from_latitude": 40.7128,
    "from_longitude": -74.0060,
    "to_latitude": 40.7580,
    "to_longitude": -73.9855
}
```

Response:
```json
{
    "destination": {
        "latitude": 40.7580,
        "longitude": -73.9855
    },
    "eta": 420,
    "distance": 2100,
    "traffic": "moderate",
    "alternate_routes": [
        {
            "distance": 2500,
            "duration": 480,
            "traffic": "light"
        }
    ]
}
```

Flow:
1. Call external mapping service (Google Maps, Mapbox)
2. Get ETA and route data
3. Cache result for 1-2 minutes
4. Return to client
5. Subscribe for traffic updates if available

### 4. Get Location History

Request:
```
GET /tracking/history/ride-123?limit=100
```

Response:
```json
{
    "ride_id": "ride-123",
    "locations": [
        {
            "latitude": 40.7128,
            "longitude": -74.0060,
            "accuracy": 10.5,
            "speed": 15.2,
            "timestamp": "2024-02-20T10:25:00Z"
        }
    ],
    "total_distance": 2150,
    "total_duration": 450,
    "average_speed": 17.2
}
```

Flow:
1. Fetch location history from database
2. Order by timestamp descending
3. Apply limit for pagination
4. Calculate statistics
5. Return complete history

### 5. Geofence Check

Request:
```
POST /tracking/geofence-check
{
    "latitude": 40.7128,
    "longitude": -74.0060,
    "zones": [
        {
            "id": "zone-1",
            "center_latitude": 40.7100,
            "center_longitude": -74.0000,
            "radius": 500
        }
    ]
}
```

Response:
```json
{
    "inside_zones": ["zone-1"],
    "zone_details": [
        {
            "zone_id": "zone-1",
            "distance": 150,
            "inside": true
        }
    ]
}
```

Flow:
1. Calculate distance to each zone center
2. Check if inside zone (distance < radius)
3. Return matching zones
4. Trigger zone-specific actions

## Distance Calculation

Using Haversine formula for great-circle distances:

```
a = sin²(Δφ/2) + cos(φ1) * cos(φ2) * sin²(Δλ/2)
c = 2 * atan2(√a, √(1-a))
d = R * c

Where:
- φ is latitude in radians
- λ is longitude in radians
- R is Earth's radius (6371 km)
```

## ETA Estimation

Factors considered:

1. Distance - Straight line distance
2. Traffic - Current traffic conditions
3. Time of Day - Typical traffic patterns
4. Weather - Adverse weather impacts
5. Route Type - Highway vs city roads
6. Restrictions - Speed limits, turn restrictions

## Location Data Quality

Quality metrics:

1. Accuracy - GPS accuracy in meters
2. Freshness - Time since last update
3. Outliers - Unrealistic jumps in location
4. Frequency - Update frequency (recommended 2-3 seconds)

Data cleaning:
- Filter out-of-range coordinates
- Detect and smooth teleportation jumps
- Interpolate missing points
- Remove low-accuracy readings

## WebSocket Message Types

### Client to Server

```json
{
    "type": "location_update",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "accuracy": 10.5,
    "speed": 15.2
}

{
    "type": "ping"
}

{
    "type": "disconnect"
}
```

### Server to Client

```json
{
    "type": "driver_location_updated",
    "driver_id": "driver-123",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "eta_seconds": 420
}

{
    "type": "arrived_at_destination"
}

{
    "type": "pong"
}

{
    "type": "error",
    "message": "Invalid location data"
}
```

## Error Handling

Common error scenarios:

1. Invalid Coordinates
   - Response: 400 Bad Request
   - Message: "Invalid latitude: must be between -90 and 90"

2. Location Update Too Frequent
   - Response: 429 Too Many Requests
   - Action: Buffer updates

3. Mapping Service Timeout
   - Response: 504 Gateway Timeout
   - Action: Use cached ETA or alternative service

4. Geofence Configuration Invalid
   - Response: 400 Bad Request
   - Message: "Invalid geofence parameters"

## Testing Strategy

### Unit Tests (Service Layer)

```go
Test_CalculateDistance_Haversine()
Test_EstimateArrival_WithTraffic()
Test_CheckGeofence_Inside()
Test_CheckGeofence_Outside()
Test_BroadcastLocationUpdate()
```

### Integration Tests (Repository Layer)

```go
Test_SaveLocation_History()
Test_GetLocationHistory_Retrieval()
Test_ClearOldLocations()
```

### End-to-End Tests (Handler Layer)

```go
Test_UpdateLocation_FullFlow()
Test_WebSocket_LocationStreaming()
Test_GetETA_FullFlow()
```

## Database Schema

### Locations Table (Time-Series)

```sql
CREATE TABLE locations (
    id VARCHAR(36) PRIMARY KEY,
    ride_id VARCHAR(36) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    accuracy DECIMAL(8, 2),
    altitude DECIMAL(10, 2),
    speed DECIMAL(8, 2),
    heading DECIMAL(6, 2),
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (ride_id) REFERENCES rides(id),
    INDEX (ride_id, timestamp),
    INDEX (timestamp) -- for retention cleanup
);
```

### Routes Table

```sql
CREATE TABLE routes (
    id VARCHAR(36) PRIMARY KEY,
    ride_id VARCHAR(36) UNIQUE,
    start_latitude DECIMAL(10, 8),
    start_longitude DECIMAL(11, 8),
    end_latitude DECIMAL(10, 8),
    end_longitude DECIMAL(11, 8),
    distance DECIMAL(10, 2),
    duration INT,
    polyline LONGTEXT,
    created_at TIMESTAMP,
    FOREIGN KEY (ride_id) REFERENCES rides(id)
);
```

## External Service Integration

### Mapping Services

Options:
1. Google Maps API
2. Mapbox
3. Open Street Map
4. Here Maps

Configuration:
```yaml
tracking:
  mapping_service: "google_maps"
  google_maps:
    api_key: "your-api-key"
    max_requests_per_second: 50
  caching:
    enabled: true
    ttl: 120s
```

## Performance Optimization

1. Use time-series database (InfluxDB, TimescaleDB)
2. Batch location updates (every 2-3 seconds)
3. Cache ETA results for common routes
4. Use geohashing for geofence queries
5. Implement connection pooling for mapping services
6. Archive old location data monthly

## Integration Points

1. Rides Module - For ride location tracking
2. Pricing Module - For distance-based fare
3. Drivers Module - For driver location
4. Messages Module - For arrival notifications

## Configuration

Typical tracking configuration:

```yaml
tracking:
  update_frequency: 2s
  max_batch_size: 100
  retention_days: 90
  geofence:
    default_radius: 500
    update_frequency: 10s
  eta:
    cache_ttl: 120s
    traffic_update_interval: 60s
  websocket:
    ping_interval: 30s
    max_connections_per_ride: 10
```

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture
- See RIDES-MODULE.md for ride tracking integration
- See PRICING-MODULE.md for distance calculations

## Common Pitfalls

1. Not validating coordinate ranges
2. Inefficient distance calculations
3. Missing WebSocket connection cleanup
4. Not handling mapping service failures
5. Insufficient location data retention policy
6. Race conditions in concurrent updates
7. Memory leaks from unclosed WebSocket connections

## Future Enhancements

1. Machine learning-based traffic prediction
2. Real-time traffic integration
3. Offline location caching
4. Route optimization algorithms
5. Heatmap generation for popular routes
6. Predictive arrival times
7. Toll and fuel cost estimates
8. Multi-modal transportation routing
