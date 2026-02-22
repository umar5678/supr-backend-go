# Pricing Module Development Guide

## Overview

The Pricing Module handles all fare-related calculations, surge pricing logic, and dynamic pricing management. It provides fare estimates, determines surge multipliers based on demand, and manages pricing rules.

## Module Structure

```
pricing/
├── handler.go         # HTTP request handlers
├── service.go         # Pricing calculation logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. Fare Estimation - Calculate estimated fares for ride requests
2. Surge Pricing - Determine surge multipliers based on demand
3. Pricing Rules - Manage and apply pricing configurations
4. Distance/Time Rates - Apply distance and duration-based pricing
5. Special Pricing - Handle promotions and discounts
6. Base Fare - Manage minimum fare amounts

## Architecture

### Handler Layer (handler.go)

Manages HTTP endpoints for pricing operations.

Key methods:

```
GetFareEstimate(c *gin.Context)         // POST /pricing/estimate
GetSurgeMultiplier(c *gin.Context)      // GET /pricing/surge
GetPricingRules(c *gin.Context)         // GET /pricing/rules
CreatePricingRule(c *gin.Context)       // POST /pricing/rules (admin)
UpdatePricingRule(c *gin.Context)       // PUT /pricing/rules/{id} (admin)
DeletePricingRule(c *gin.Context)       // DELETE /pricing/rules/{id} (admin)
GetCancellationFee(c *gin.Context)      // GET /pricing/cancellation-fee
```

Request flow:
1. Extract parameters from request
2. Validate input data
3. Call service method
4. Return pricing response

### Service Layer (service.go)

Contains pricing business logic and calculations.

Key interface methods:

```
GetFareEstimate(ctx context.Context, req FareEstimateRequest) (*FareEstimateResponse, error)
GetSurgeMultiplier(ctx context.Context, latitude, longitude float64) (float64, error)
CalculateFare(ctx context.Context, distance float64, duration int, rideType string) (float64, error)
GetPricingRules(ctx context.Context, rideType string) ([]*models.PricingRule, error)
CreatePricingRule(ctx context.Context, rule *models.PricingRule) error
UpdatePricingRule(ctx context.Context, ruleID string, updates map[string]interface{}) error
GetCancellationFee(ctx context.Context, rideStatus string) (float64, error)
ApplyPromotion(ctx context.Context, fare float64, promoCode string) (float64, error)
```

Logic flow:
1. Fetch applicable pricing rules
2. Calculate base fare
3. Add distance charges
4. Add time charges
5. Apply surge multiplier
6. Apply promotions/discounts
7. Apply minimum/maximum fare caps
8. Return final fare

### Repository Layer (repository.go)

Handles database operations for pricing.

Key interface methods:

```
GetPricingRules(ctx context.Context, filters map[string]interface{}) ([]*models.PricingRule, error)
CreatePricingRule(ctx context.Context, rule *models.PricingRule) error
UpdatePricingRule(ctx context.Context, ruleID string, rule *models.PricingRule) error
DeletePricingRule(ctx context.Context, ruleID string) error
GetSurgeRules(ctx context.Context) ([]*models.SurgeRule, error)
GetRideTypePricing(ctx context.Context, rideType string) (*models.RidePricing, error)
```

Database operations:
- Store pricing rules with effective dates
- Manage surge rule configurations
- Track pricing history
- Use caching for frequently accessed rules

## Data Transfer Objects

### FareEstimateRequest

```go
type FareEstimateRequest struct {
    PickupLocation   Location `json:"pickup_location" binding:"required"`
    DropoffLocation  Location `json:"dropoff_location" binding:"required"`
    RideType         string   `json:"ride_type" binding:"required"`
    ScheduledTime    *time.Time `json:"scheduled_time,omitempty"`
    PromoCode        string   `json:"promo_code,omitempty"`
}

type Location struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}
```

### FareEstimateResponse

```go
type FareEstimateResponse struct {
    EstimatedFare    float64         `json:"estimated_fare"`
    BaseFare         float64         `json:"base_fare"`
    DistanceCharge   float64         `json:"distance_charge"`
    TimeCharge       float64         `json:"time_charge"`
    SurgeMultiplier  float64         `json:"surge_multiplier"`
    DiscountAmount   float64         `json:"discount_amount,omitempty"`
    TaxAmount        float64         `json:"tax_amount,omitempty"`
    EstimatedDistance float64        `json:"estimated_distance"` // meters
    EstimatedDuration int            `json:"estimated_duration"` // seconds
    Breakdown        []ChargeItem    `json:"breakdown"`
    Notes            string          `json:"notes,omitempty"`
}

type ChargeItem struct {
    Label  string  `json:"label"`
    Amount float64 `json:"amount"`
}
```

### SurgeMultiplierResponse

```go
type SurgeMultiplierResponse struct {
    Latitude         float64   `json:"latitude"`
    Longitude        float64   `json:"longitude"`
    SurgeMultiplier  float64   `json:"surge_multiplier"`
    Demand           string    `json:"demand"` // low, medium, high, very_high
    AvailableDrivers int       `json:"available_drivers"`
    Message          string    `json:"message"`
}
```

## Fare Calculation Formula

### Basic Formula

```
Fare = (BaseFare + (Distance * DistanceRate) + (Duration * TimeRate)) * SurgeMultiplier
```

### With Promotions and Tax

```
Fare = ((BaseFare + (Distance * DistanceRate) + (Duration * TimeRate)) * SurgeMultiplier) - Discount + Tax

Where:
- BaseFare: Minimum fare (ride type specific)
- Distance: Trip distance in kilometers
- DistanceRate: Per km rate (ride type specific)
- Duration: Trip duration in minutes
- TimeRate: Per minute rate (ride type specific)
- SurgeMultiplier: Dynamic multiplier (1.0 to 3.0+)
- Discount: Promotion/coupon discount
- Tax: Local taxes (percentage-based)
```

## Ride Types and Pricing

### Economy

```
Base Fare: 2.00
Distance Rate: 0.75 per km
Time Rate: 0.25 per minute
Min Fare: 4.00
Max Surge: 2.5x
```

### Premium

```
Base Fare: 3.50
Distance Rate: 1.25 per km
Time Rate: 0.45 per minute
Min Fare: 6.00
Max Surge: 2.0x
```

### XL

```
Base Fare: 5.00
Distance Rate: 1.50 per km
Time Rate: 0.50 per minute
Min Fare: 8.00
Max Surge: 2.0x
```

## Surge Pricing Algorithm

Surge multiplier is calculated based on:

1. Demand Indicator
   - Active ride requests in zone
   - Completed rides in last hour
   - Average wait time

2. Supply Indicator
   - Available drivers in zone
   - Drivers on rides
   - New drivers arriving

3. Time Factor
   - Peak hours (higher multiplier)
   - Off-peak hours (lower multiplier)
   - Special events

4. Weather Factor
   - Rain or adverse conditions
   - Visibility issues

```
DemandScore = (ActiveRequests / 10) + (CompletedRides / 5)
SupplyScore = (AvailableDrivers / 20) + (ArrivingDrivers / 10)
DemandRatio = DemandScore / (SupplyScore + 0.1)

BaseMultiplier = min(1.0 + (DemandRatio * 0.5), MaxMultiplier)
TimeAdjustment = isPeakHour ? 1.2 : 1.0
WeatherAdjustment = hasAdverseWeather ? 1.3 : 1.0

SurgeMultiplier = BaseMultiplier * TimeAdjustment * WeatherAdjustment
```

## Cancellation Fees

```
If cancelled within 5 minutes: No charge
If cancelled 5-10 minutes after request: 10% of estimated fare
If cancelled after driver arrival: 25% of estimated fare
If cancelled after ride started: Full fare charged
```

## Typical Use Cases

### 1. Get Fare Estimate

Request:
```
POST /pricing/estimate
{
    "pickup_location": {
        "latitude": 40.7128,
        "longitude": -74.0060
    },
    "dropoff_location": {
        "latitude": 40.7580,
        "longitude": -73.9855
    },
    "ride_type": "economy"
}
```

Flow:
1. Validate locations
2. Call distance matrix service for distance/duration
3. Get surge multiplier for pickup location
4. Apply pricing rules for ride type
5. Calculate fare using formula
6. Return detailed breakdown

### 2. Get Surge Multiplier

Request:
```
GET /pricing/surge?latitude=40.7128&longitude=-74.0060
```

Flow:
1. Get active requests in zone (1km radius)
2. Count available drivers in zone
3. Calculate demand/supply ratio
4. Apply surge formula
5. Return multiplier with demand level

### 3. Apply Promotion Code

Request:
```
POST /pricing/estimate
{
    ...,
    "promo_code": "SUMMER20"
}
```

Flow:
1. Validate promo code exists and is active
2. Check code usage limits
3. Verify code applies to ride type
4. Calculate discount amount
5. Subtract from total fare
6. Apply minimum fare if applicable

## Pricing Rules Management

### Create Pricing Rule

Request:
```
POST /pricing/rules
{
    "ride_type": "economy",
    "base_fare": 2.50,
    "distance_rate": 0.80,
    "time_rate": 0.30,
    "min_fare": 4.50,
    "max_fare": 100.00,
    "effective_from": "2024-02-01",
    "effective_to": "2024-12-31"
}
```

### Update Pricing Rule

Request:
```
PUT /pricing/rules/{ruleID}
{
    "base_fare": 2.75,
    "distance_rate": 0.85
}
```

## Error Handling

Common error scenarios:

1. Invalid Locations
   - Response: 400 Bad Request
   - Action: Return error details

2. Invalid Ride Type
   - Response: 400 Bad Request
   - Action: List valid ride types

3. Promo Code Not Found
   - Response: 404 Not Found
   - Action: Continue without promo

4. Promo Code Expired
   - Response: 400 Bad Request
   - Action: Return expiration details

5. Promo Code Usage Exceeded
   - Response: 400 Bad Request
   - Action: Return usage limit message

## Testing Strategy

### Unit Tests (Service Layer)

```go
Test_CalculateFare_BasicFormula()
Test_CalculateFare_WithSurge()
Test_CalculateFare_WithPromotion()
Test_GetSurgeMultiplier_LowDemand()
Test_GetSurgeMultiplier_HighDemand()
Test_GetCancellationFee_Before5Minutes()
Test_GetCancellationFee_AfterPickup()
```

### Integration Tests (Repository Layer)

```go
Test_GetPricingRules_ByRideType()
Test_CreatePricingRule()
Test_UpdatePricingRule()
```

### End-to-End Tests (Handler Layer)

```go
Test_GetFareEstimate_FullFlow()
Test_GetSurgeMultiplier_FullFlow()
```

## Database Schema

### Pricing Rules Table

```sql
CREATE TABLE pricing_rules (
    id VARCHAR(36) PRIMARY KEY,
    ride_type VARCHAR(50),
    base_fare DECIMAL(10, 2),
    distance_rate DECIMAL(10, 2),
    time_rate DECIMAL(10, 2),
    min_fare DECIMAL(10, 2),
    max_fare DECIMAL(10, 2),
    effective_from DATE,
    effective_to DATE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Surge Rules Table

```sql
CREATE TABLE surge_rules (
    id VARCHAR(36) PRIMARY KEY,
    zone_id VARCHAR(36),
    demand_multiplier DECIMAL(5, 2),
    supply_multiplier DECIMAL(5, 2),
    peak_hours_start TIME,
    peak_hours_end TIME,
    peak_multiplier DECIMAL(5, 2),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Pricing History Table

```sql
CREATE TABLE pricing_history (
    id VARCHAR(36) PRIMARY KEY,
    ride_id VARCHAR(36),
    base_fare DECIMAL(10, 2),
    distance_charge DECIMAL(10, 2),
    time_charge DECIMAL(10, 2),
    surge_multiplier DECIMAL(5, 2),
    final_fare DECIMAL(10, 2),
    created_at TIMESTAMP
);
```

## Performance Optimization

1. Cache pricing rules with short TTL
2. Use Redis for surge multiplier calculations
3. Batch distance matrix requests
4. Index on ride_type and effective dates
5. Pre-calculate peak hours
6. Cache promo code validations

## Integration Points

1. Rides Module - For fare calculations
2. Wallet Module - For payment processing
3. Promotions Module - For discount codes
4. Tracking Module - For distance/duration data
5. Drivers Module - For driver count in zones

## Configuration

Typical pricing configuration:

```yaml
pricing:
  base_rates:
    economy:
      base_fare: 2.00
      distance_rate: 0.75
      time_rate: 0.25
    premium:
      base_fare: 3.50
      distance_rate: 1.25
      time_rate: 0.45
  surge:
    max_multiplier: 3.0
    update_interval: 60s
  cancellation:
    no_charge_window: 300s
    percentage_fee: 0.10
```

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture
- See RIDES-MODULE.md for fare integration
- See PROMOTIONS-MODULE.md for discount handling
- See TRACKING-MODULE.md for distance calculations

## Common Pitfalls

1. Not caching pricing rules
2. Inefficient surge calculations
3. Race conditions in price updates
4. Not validating promotion codes early
5. Incorrect formula implementation
6. Missing minimum/maximum fare caps
7. Not handling timezone-aware peak hours

## Future Enhancements

1. Machine learning-based demand prediction
2. Personalized pricing based on user history
3. Time-of-day adjustments
4. Weather-based surge multipliers
5. Competitor price matching
6. Dynamic pricing by zone
7. Subscription/membership pricing
8. Referral rewards integration
