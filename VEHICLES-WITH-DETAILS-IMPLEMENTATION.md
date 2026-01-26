# Vehicles with Details Implementation Guide

## What Was Added

A new comprehensive endpoint for displaying available vehicles with complete pricing and demand information when a user selects pickup and destination locations.

## Files Modified

### 1. **internal/modules/rides/dto/vehicles_details.go** (NEW)
- `VehicleDetailsRequest`: Request DTO with pickup/dropoff coordinates
- `VehicleWithDetailsResponse`: Single vehicle response with all details
- `VehiclesWithDetailsListResponse`: List response with trip info

### 2. **internal/modules/rides/handler.go**
- Added `GetVehiclesWithDetails()` handler method
- Swagger documentation included

### 3. **internal/modules/rides/service.go**
- Added `GetVehiclesWithDetails()` method to Service interface
- Implemented service logic that:
  - Fetches nearby drivers
  - Calculates trip distance and duration
  - Gets fare estimates from pricing service
  - Retrieves surge multipliers
  - Gets current demand data
  - Enriches driver data with all pricing information

### 4. **internal/modules/rides/routes.go**
- Registered new route: `POST /rides/vehicles-with-details`

## How It Works

### Request Flow
```
Client sends POST request with pickup & dropoff coordinates
    ↓
Handler validates request
    ↓
Service layer:
  1. Calculates trip distance & duration
  2. Finds nearby drivers in radius
  3. For each driver:
     - Gets vehicle details
     - Calculates distance to pickup
     - Gets fare estimate
     - Gets surge multiplier
     - Gets demand data
  4. Sorts by distance
  5. Returns top 20 vehicles
    ↓
Response with all vehicle details
```

### Data Aggregation

The endpoint combines data from:

1. **Drivers Module**: 
   - Driver profiles and ratings
   - Vehicle information
   - Current location and heading
   - Online status and verification

2. **Pricing Module**:
   - Fare estimation (BaseFare + Distance + Time)
   - Surge multipliers (time-based and demand-based)
   - Demand tracking (pending requests vs available drivers)

3. **Location Utilities**:
   - Distance calculation (Haversine formula)
   - ETA calculation (based on 40 km/h average speed)

## Key Features

### 1. Comprehensive Pricing Information
```json
{
  "baseFare": 100.0,
  "perKmRate": 15.0,
  "perMinRate": 1.5,
  "estimatedFare": 227.5,
  "surgeMultiplier": 1.2,
  "surgeReason": "peak_hours"
}
```

### 2. Demand Tracking
```json
{
  "pendingRequests": 12,
  "availableDrivers": 45,
  "demand": "normal"
}
```

### 3. ETA Information
- Time for driver to reach pickup location
- Both in seconds and formatted string ("3 min")
- Based on current driver location

### 4. Trip Estimates
- Total trip distance
- Total trip duration
- Displayed in both seconds and minutes

### 5. Driver Quality Metrics
```json
{
  "driverRating": 4.8,
  "acceptanceRate": 98.5,
  "cancellationRate": 1.2,
  "totalTrips": 450,
  "isVerified": true
}
```

## Usage in Frontend

### Basic Implementation
```typescript
// 1. When user selects pickup and destination, call this endpoint
const response = await api.post('/rides/vehicles-with-details', {
  pickupLat: 24.8607,
  pickupLon: 67.0011,
  pickupAddress: 'Current Location',
  dropoffLat: 24.9200,
  dropoffLon: 67.1000,
  dropoffAddress: 'Destination',
  radiusKm: 5.0
});

const { vehicles, tripDistance, tripDurationMins } = response.data.data;

// 2. Display vehicles in a list
vehicles.forEach(vehicle => {
  console.log(`${vehicle.vehicleDisplayName} - Rs. ${vehicle.estimatedFare}`);
  console.log(`Driver: ${vehicle.driverName} (${vehicle.driverRating}⭐)`);
  console.log(`ETA: ${vehicle.etaFormatted}`);
  console.log(`Surge: ${vehicle.surgeMultiplier}x`);
});

// 3. When user selects a vehicle, use estimatedFare in your booking
```

### Screen Layout Recommendation

```
┌─────────────────────────────────┐
│ 📍 Current Location             │
│ → Distance Zone: 8.5 km         │
│ 📍 Destination                  │
│ Estimated time: 17 min          │
└─────────────────────────────────┘

[Vehicle List - Sorted by distance]

┌─────────────────────────────────┐
│ Driver Name          ⭐ 4.8      │
│ Toyota Corolla - Silver         │
│ 0.8 km away    3 min away       │
│ Rs. 227.50      ⚡ 1.2x surge   │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ Driver Name 2        ⭐ 4.5      │
│ Honda Civic - Blue              │
│ 1.2 km away    4 min away       │
│ Rs. 210.00      Normal pricing  │
└─────────────────────────────────┘
```

## Filtering & Sorting

### Filtering (Automatic)
- Online status: "online" only
- Verified drivers only
- Rating filter: High-rated riders (≥4.5) exclude low-rated drivers (<2.5)
- Valid vehicle and location data

### Sorting
- Primary: By distance to pickup (closest first)
- Limit: Top 20 results

## Error Handling

```typescript
try {
  const response = await api.post('/rides/vehicles-with-details', request);
  
  if (response.data.status === 'error') {
    // Handle validation errors
    console.error(response.data.message);
    
    if (response.data.message.includes('0.5 km')) {
      // Destination too close
    } else if (response.data.message.includes('100 km')) {
      // Destination too far
    }
  } else {
    // Process vehicles
    const { vehicles } = response.data.data;
    if (vehicles.length === 0) {
      // No drivers available - increase radius or retry
    }
  }
} catch (error) {
  console.error('Request failed:', error);
}
```

## Performance Optimization

1. **Caching**: Demand and surge data are cached
2. **Batch Processing**: Multiple parallel queries to pricing service
3. **Limit**: Maximum 20 vehicles returned
4. **Filtering**: Done in-memory after fetch

## Testing

### Test Case 1: Normal Demand
```bash
POST /rides/vehicles-with-details
{
  "pickupLat": 24.8607,
  "pickupLon": 67.0011,
  "pickupAddress": "Main Street",
  "dropoffLat": 24.9200,
  "dropoffLon": 67.1000,
  "dropoffAddress": "Destination"
}
```
Expected: 5-15 vehicles with surge ~1.0x

### Test Case 2: Peak Hours
```bash
POST /rides/vehicles-with-details
[same as above, but call during 8-9 AM or 5-6 PM]
```
Expected: Same vehicles with surge 1.2x - 2.0x

### Test Case 3: No Drivers
```bash
POST /rides/vehicles-with-details
{
  "pickupLat": 38.5816,
  "pickupLon": 68.1111,
  "pickupAddress": "Remote Area",
  "dropoffLat": 38.5900,
  "dropoffLon": 68.1200,
  "dropoffAddress": "Remote Destination"
}
```
Expected: Empty vehicles array

### Test Case 4: Invalid Distance
```bash
POST /rides/vehicles-with-details
{
  "pickupLat": 24.8607,
  "pickupLon": 67.0011,
  "pickupAddress": "Same Location",
  "dropoffLat": 24.8607,
  "dropoffLon": 67.0011,
  "dropoffAddress": "Same Location"
}
```
Expected: Error - "Minimum trip distance is 0.5 km"

## Future Enhancements

1. **Preferences**: Filter by vehicle type, price range
2. **Favorites**: Show favorite drivers
3. **Real-time Updates**: WebSocket for live driver updates
4. **Promos**: Apply promo codes to get discounted estimates
5. **Scheduling**: Show availability for scheduled rides
6. **Carbon Footprint**: Show eco-friendly vehicle options

## Notes

- All times are in UTC
- Distances in kilometers
- Prices in local currency (PKR for Pakistan)
- Vehicle capacity includes driver
- Driver ratings are out of 5.0
- Surge multiplier can go below 1.0 during low-demand periods (discount pricing)

