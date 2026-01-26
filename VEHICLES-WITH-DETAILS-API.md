# Vehicles with Details Endpoint Documentation

## Overview
The **`POST /rides/vehicles-with-details`** endpoint provides a comprehensive view of available vehicles when a rider selects a pickup and destination location. This is the main endpoint for the vehicle selection screen.

## Endpoint Details

**Method:** POST  
**Path:** `/rides/vehicles-with-details`  
**Auth Required:** Yes (Bearer Token)  
**Content-Type:** application/json

## Request Body

```json
{
  "pickupLat": 24.8607,
  "pickupLon": 67.0011,
  "pickupAddress": "123 Main Street, Karachi",
  "dropoffLat": 24.9200,
  "dropoffLon": 67.1000,
  "dropoffAddress": "University of Karachi",
  "radiusKm": 5.0
}
```

### Parameters

| Parameter | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `pickupLat` | float | Yes | Pickup latitude | -90 to 90 |
| `pickupLon` | float | Yes | Pickup longitude | -180 to 180 |
| `pickupAddress` | string | Yes | Pickup address | Max 500 chars |
| `dropoffLat` | float | Yes | Destination latitude | -90 to 90 |
| `dropoffLon` | float | Yes | Destination longitude | -180 to 180 |
| `dropoffAddress` | string | Yes | Destination address | Max 500 chars |
| `radiusKm` | float | No | Search radius for drivers | 0.1 to 50, defaults to 5 |

## Response Structure

```json
{
  "status": "success",
  "message": "Vehicles with details fetched successfully",
  "data": {
    "pickupLat": 24.8607,
    "pickupLon": 67.0011,
    "pickupAddress": "123 Main Street, Karachi",
    "dropoffLat": 24.9200,
    "dropoffLon": 67.1000,
    "dropoffAddress": "University of Karachi",
    "radiusKm": 5.0,
    "tripDistance": 8.5,
    "tripDuration": 1020,
    "tripDurationMins": 17,
    "totalCount": 12,
    "carsCount": 12,
    "vehicles": [
      {
        "id": "uuid-vehicle-1",
        "driverId": "uuid-driver-1",
        "driverName": "Ahmed Hassan",
        "driverRating": 4.8,
        "driverImage": "https://...",
        "acceptanceRate": 98.5,
        "cancellationRate": 1.2,
        "totalTrips": 450,
        "isVerified": true,
        "status": "online",
        
        "vehicleTypeId": "uuid-type-1",
        "vehicleType": "economy",
        "vehicleDisplayName": "Economy",
        "make": "Toyota",
        "model": "Corolla",
        "year": 2022,
        "color": "Silver",
        "licensePlate": "DHA-1234",
        "capacity": 4,
        
        "currentLatitude": 24.8650,
        "currentLongitude": 67.0050,
        "heading": 45,
        "distanceKm": 0.8,
        
        "etaSeconds": 180,
        "etaMinutes": 3,
        "etaFormatted": "3 min",
        
        "baseFare": 100.0,
        "perKmRate": 15.0,
        "perMinRate": 1.5,
        "estimatedFare": 227.5,
        "estimatedDistance": 8.5,
        "estimatedDuration": 1020,
        "estimatedDurationMins": 17,
        
        "surgeMultiplier": 1.2,
        "surgeReason": "peak_hours",
        
        "pendingRequests": 12,
        "availableDrivers": 45,
        "demand": "normal",
        
        "updatedAt": "2026-01-26T10:30:00Z",
        "timestamp": "2026-01-26T10:35:00Z"
      },
      // ... more vehicles
    ],
    "timestamp": "2026-01-26T10:35:00Z"
  }
}
```

## Response Fields Explanation

### Trip Information
- **tripDistance** (float): Total trip distance in kilometers
- **tripDuration** (int): Total trip duration in seconds
- **tripDurationMins** (int): Total trip duration in minutes

### Vehicle Details
- **vehicleType**: Vehicle class (economy, comfort, premium, xl)
- **vehicleDisplayName**: User-friendly display name
- **capacity**: Number of passengers the vehicle can accommodate

### Location Data
- **distanceKm**: Distance from driver to pickup location
- **etaSeconds**: Time for driver to reach pickup (in seconds)
- **etaMinutes**: Time for driver to reach pickup (in minutes)
- **etaFormatted**: Human-readable ETA (e.g., "3 min")
- **heading**: Driver's current heading (0-360 degrees)

### Pricing Information
- **baseFare**: Base fare for this vehicle type
- **perKmRate**: Rate charged per kilometer
- **perMinRate**: Rate charged per minute of wait time
- **estimatedFare**: Total estimated fare for the complete trip
  - Calculation: BaseFare + (Distance × PerKmRate) + (Duration × PerMinRate) × SurgeMultiplier

### Surge Pricing
- **surgeMultiplier**: Current surge pricing multiplier (1.0 = no surge, 2.0 = 2x price)
- **surgeReason**: Reason for surge: "normal", "peak_hours", "high_demand", "extreme_demand"
  - Multiplier > 2.0: "extreme"
  - Multiplier > 1.5: "high"
  - Multiplier > 1.0: "high"
  - Multiplier < 1.0: "low"

### Demand Information
- **pendingRequests**: Number of pending ride requests in the area
- **availableDrivers**: Number of online drivers in the area
- **demand**: Demand level indicator:
  - "low": Surge < 1.0 (Low demand, cheaper fares)
  - "normal": Surge 1.0 - 1.0 (Normal demand)
  - "high": Surge 1.0 - 1.5 (Higher demand, increased prices)
  - "extreme": Surge > 2.0 (Extreme demand, significantly increased prices)

### Driver Information
- **driverRating**: Driver rating from 0-5 stars
- **acceptanceRate**: Percentage of ride requests driver accepts
- **cancellationRate**: Percentage of accepted rides driver cancels
- **totalTrips**: Total number of rides driver has completed
- **isVerified**: Whether driver identity is verified

## Sorting & Limits

- Vehicles are **sorted by distance** (closest drivers first)
- Maximum **20 vehicles** are returned
- Drivers are filtered for:
  - Status: "online"
  - Verified drivers only
  - Rating filters applied based on rider profile (high-rated riders exclude low-rated drivers)

## Error Responses

### Invalid Location
```json
{
  "status": "error",
  "message": "Minimum trip distance is 0.5 km"
}
```

### No Drivers Available
```json
{
  "status": "success",
  "message": "Vehicles with details fetched successfully",
  "data": {
    "totalCount": 0,
    "carsCount": 0,
    "vehicles": []
  }
}
```

### Validation Error
```json
{
  "status": "error",
  "message": "Invalid request body"
}
```

## Usage Example

### JavaScript/TypeScript

```typescript
const response = await fetch('https://api.example.com/rides/vehicles-with-details', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    pickupLat: 24.8607,
    pickupLon: 67.0011,
    pickupAddress: '123 Main Street, Karachi',
    dropoffLat: 24.9200,
    dropoffLon: 67.1000,
    dropoffAddress: 'University of Karachi',
    radiusKm: 5.0
  })
});

const data = await response.json();
console.log('Available vehicles:', data.data.vehicles);
```

### cURL

```bash
curl -X POST https://api.example.com/rides/vehicles-with-details \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pickupLat": 24.8607,
    "pickupLon": 67.0011,
    "pickupAddress": "123 Main Street, Karachi",
    "dropoffLat": 24.9200,
    "dropoffLon": 67.1000,
    "dropoffAddress": "University of Karachi",
    "radiusKm": 5.0
  }'
```

## Data Sources (from pricing endpoints)

This endpoint aggregates data from multiple internal endpoints:

| Data | Source | Endpoint |
|------|--------|----------|
| Driver locations & vehicles | Drivers module | FindNearbyDrivers |
| Fare estimates | Pricing module | GetFareEstimate |
| Surge multipliers | Pricing module | CalculateCombinedSurge |
| Demand tracking | Pricing module | GetCurrentDemand |
| Driver profiles | Drivers module | Driver profiles |

## Performance Notes

- Response time: ~300-500ms (includes multiple service calls)
- Caches are utilized for demand and surge data
- Drivers are filtered on-the-fly to ensure only verified, online drivers are shown

## Frontend Display Guide

Recommended UI elements:
1. **Trip Summary**: Show pickupAddress → dropoffAddress with tripDistance and tripDurationMins
2. **Vehicle Cards**: Display in a scrollable list, sorted by distance
3. **Price Highlight**: Show estimatedFare prominently with surge indicator
4. **Driver Info**: Show driverName, driverRating with star icons
5. **ETA Badge**: Display etaFormatted (e.g., "3 min away")
6. **Demand Indicator**: Show demand level with color coding
7. **Vehicle Details**: Make, model, color, license plate below driver info

## Common Issues & Solutions

### No vehicles returned
- Check if `radiusKm` is large enough
- Verify pickup/dropoff coordinates are valid
- Check if there are online drivers in the area

### High estimated fares
- Check surge multiplier (peak hours may have 1.5x-2.0x multiplier)
- Trip distance calculation - verify coordinates are correct
- Check vehicle type's pricing rates

### Incorrect ETAs
- Ensure coordinates are in valid lat/lon format
- ETA is based on 40 km/h average city speed

