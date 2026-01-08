# Surge Pricing Configuration Guide

## Overview

Surge pricing in Supr can be managed through:
1. **API Endpoints** - for dynamic rule creation by admins
2. **Database Inserts** - for initial setup or bulk operations
3. **Admin Dashboard** - future UI for managing rules

---

## Who Creates Surge Rules?

### 1. **Admin Users** (Primary)
- Via API endpoint: `POST /api/v1/pricing/surge-rules`
- Can create, update, and manage surge rules in real-time
- Requires authentication and admin role

### 2. **Database Administrators**
- Direct database inserts for initial setup
- Batch creation for multiple rules
- Direct SQL for complex migrations

### 3. **System** (Automated)
- Demand tracking system records real-time demand metrics
- Used by demand-based surge calculation
- Automatic recording during ride requests

---

## Creating Surge Rules via API

### Endpoint
```
POST /api/v1/pricing/surge-rules
```

### Request Body
```json
{
  "name": "Peak Hours Surge - Economy",
  "description": "Extra surge during peak commute hours",
  "vehicleTypeId": "uuid-of-economy-vehicle",
  "dayOfWeek": 1,  // 0 = Sunday, 1 = Monday, etc. -1 = All days
  "startTime": "08:00",  // HH:MM format
  "endTime": "10:00",
  "baseMultiplier": 1.5,
  "minMultiplier": 1.0,
  "maxMultiplier": 2.0,
  "enableDemandBasedSurge": true,
  "demandThreshold": 10,  // Number of pending requests before surge kicks in
  "demandMultiplierPerRequest": 0.05  // +5% per pending request above threshold
}
```

### Response
```json
{
  "success": true,
  "data": {
    "id": "rule-uuid",
    "name": "Peak Hours Surge - Economy",
    "baseMultiplier": 1.5,
    "isActive": true,
    "createdAt": "2026-01-07T10:00:00Z"
  },
  "message": "Surge pricing rule created successfully"
}
```

---

## Creating Surge Rules via Database

### SQL Insert Examples

#### 1. Time-Based Surge (Peak Hours)
```sql
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Morning Peak - All Vehicles',
  'Surge during 7-10 AM on weekdays',
  NULL,  -- NULL means applies to all vehicle types
  1,     -- Monday (0=Sunday, 1=Monday, ..., 6=Saturday)
  '07:00',
  '10:00',
  1.5,   -- 1.5x surge
  1.0,   -- Min 1.0x
  2.0,   -- Max 2.0x
  false, -- Disable demand-based for this rule
  0,
  true,
  NOW(),
  NOW()
);
```

#### 2. Weekend Surge
```sql
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Evening Weekend Surge',
  'Extra surge on Saturday/Sunday evenings',
  NULL,
  6,     -- Saturday
  '18:00',
  '23:59',
  1.75,
  1.0,
  2.5,
  true,  -- Enable demand-based
  15,    -- Surge kicks in at 15 pending requests
  true,
  NOW(),
  NOW()
);
```

#### 3. Vehicle Type Specific Surge
```sql
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Premium Vehicle Late Night',
  'Premium vehicles have higher surge at night',
  (SELECT id FROM vehicle_types WHERE name = 'premium' LIMIT 1),
  -1,    -- All days
  '22:00',
  '06:00',
  2.0,   -- 2.0x surge
  1.5,
  3.0,
  true,
  5,
  true,
  NOW(),
  NOW()
);
```

#### 4. All-Days, All-Vehicle Rule
```sql
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Demand-Based Dynamic Surge',
  'Pure demand-based surge, no fixed time window',
  NULL,
  -1,    -- All days
  '00:00',
  '23:59',
  1.0,   -- Base is 1.0 (only demand affects)
  1.0,
  2.5,
  true,  -- Enable demand-based
  10,    -- Surge at 10+ pending requests
  true,
  NOW(),
  NOW()
);
```

---

## How Surge Rules are Applied During Ride Creation

### Flow:
1. User requests a ride via `POST /api/v1/rides`
2. System calculates basic fare using `GetFareEstimate()`
3. System calls `CalculateCombinedSurge()` which:
   - **Evaluates time-based rules**: Checks current time against `start_time` and `end_time`
   - **Evaluates day-based rules**: Checks `day_of_week` against current day
   - **Evaluates demand-based rules**: Gets pending requests vs available drivers
   - **Returns combined multiplier**: Max of all applicable rules
4. Fare is adjusted with the final surge multiplier
5. Surge is recorded in `surge_history` table for audit trail

### Example Calculation:

```
Current Time: Wednesday 9:00 AM
Pending Requests: 25
Available Drivers: 20

Applicable Rules:
- Time-based (8-10 AM): 1.5x multiplier
- Demand-based (25 > 10 threshold): 1.0 + (25-10)*0.05 = 1.75x
- Vehicle-specific premium: 1.2x

Final Surge = max(1.5, 1.75, 1.2) = 1.75x

Base Fare: $10.00
Surge Amount: $10.00 * (1.75 - 1.0) = $7.50
Total Fare: $17.50
```

---

## Managing Surge Rules

### View All Active Rules
```bash
curl -X GET "http://localhost:8080/api/v1/pricing/surge-rules" \
  -H "Content-Type: application/json"
```

### Deactivate a Rule (via DB)
```sql
UPDATE surge_pricing_rules 
SET is_active = false, updated_at = NOW()
WHERE id = 'rule-uuid';
```

### Update a Rule (via DB)
```sql
UPDATE surge_pricing_rules
SET base_multiplier = 1.8, updated_at = NOW()
WHERE name = 'Peak Hours Surge - Economy';
```

### View Surge History for a Ride
```sql
SELECT * FROM surge_history 
WHERE ride_id = 'ride-uuid'
ORDER BY created_at DESC;
```

---

## Demand Tracking

### How Demand is Recorded
During ride creation, after finding drivers, the system calls:
```
POST /api/v1/pricing/record-demand
```

### Recording Demand (Internal - Called by System)
```json
{
  "zoneId": "zone-uuid",
  "geohash": "39.74_-104.99",
  "pendingRequests": 25,
  "availableDrivers": 20
}
```

### Retrieving Current Demand
```bash
curl -X GET "http://localhost:8080/api/v1/pricing/demand?geohash=39.74_-104.99"
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "demand-uuid",
    "zoneId": "zone-uuid",
    "pendingRequests": 25,
    "availableDrivers": 20,
    "demandSupplyRatio": 1.25,
    "surgeMultiplier": 1.75,
    "recordedAt": "2026-01-07T10:15:30Z"
  }
}
```

---

## ETA Calculation

### Request ETA
```bash
curl -X POST "http://localhost:8080/api/v1/pricing/calculate-eta" \
  -H "Content-Type: application/json" \
  -d '{
    "pickupLat": 40.7128,
    "pickupLon": -74.0060,
    "dropoffLat": 40.7580,
    "dropoffLon": -73.9855
  }'
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "eta-uuid",
    "distanceKm": 6.5,
    "durationSeconds": 585,
    "estimatedPickupETA": 300,
    "estimatedDropoffETA": 885,
    "trafficCondition": "normal",
    "trafficMultiplier": 1.0,
    "source": "calculated",
    "createdAt": "2026-01-07T10:20:00Z"
  }
}
```

---

## Best Practices for Surge Rule Configuration

### 1. **Avoid Rule Conflicts**
- Don't have overlapping time windows for the same vehicle type
- Use inclusive start/end times

### 2. **Balance Multipliers**
- Keep `min_multiplier` lower for fairness
- Set `max_multiplier` to prevent extreme prices
- Recommended: Min 1.0, Max 2.5

### 3. **Demand Threshold**
- Set based on driver supply in your city
- Small cities: 5-10 pending requests
- Medium cities: 10-20 pending requests
- Large cities: 20-50 pending requests

### 4. **Testing**
- Create test rules with limited time windows
- Monitor `surge_history` table for effectiveness
- Adjust based on actual ride demand

### 5. **Communication**
- Notify riders of surge pricing in app
- Show surge breakdown in fare estimate
- Be transparent about calculations

---

## Example: Complete Surge Setup

```sql
-- Weekend Peak Surge
INSERT INTO surge_pricing_rules (name, description, day_of_week, start_time, end_time, base_multiplier, min_multiplier, max_multiplier, enable_demand_based_surge, is_active, created_at, updated_at) 
VALUES ('Weekend Peak', 'Sat/Sun 6-11 PM', 6, '18:00', '23:59', 1.75, 1.0, 2.5, true, true, NOW(), NOW());

INSERT INTO surge_pricing_rules (name, description, day_of_week, start_time, end_time, base_multiplier, min_multiplier, max_multiplier, enable_demand_based_surge, is_active, created_at, updated_at) 
VALUES ('Weekend Peak', 'Sat/Sun 6-11 PM', 0, '18:00', '23:59', 1.75, 1.0, 2.5, true, true, NOW(), NOW());

-- Weekday Morning Commute
INSERT INTO surge_pricing_rules (name, description, day_of_week, start_time, end_time, base_multiplier, min_multiplier, max_multiplier, enable_demand_based_surge, is_active, created_at, updated_at) 
VALUES ('Morning Commute', 'Weekdays 7-10 AM', 1, '07:00', '10:00', 1.5, 1.0, 2.0, false, true, NOW(), NOW());

-- Evening Commute
INSERT INTO surge_pricing_rules (name, description, day_of_week, start_time, end_time, base_multiplier, min_multiplier, max_multiplier, enable_demand_based_surge, is_active, created_at, updated_at) 
VALUES ('Evening Commute', 'Weekdays 5-8 PM', 1, '17:00', '20:00', 1.6, 1.0, 2.2, true, true, NOW(), NOW());

-- Late Night Premium
INSERT INTO surge_pricing_rules (name, description, day_of_week, start_time, end_time, base_multiplier, min_multiplier, max_multiplier, enable_demand_based_surge, is_active, created_at, updated_at) 
VALUES ('Late Night', 'All days after 10 PM', -1, '22:00', '05:59', 1.8, 1.0, 2.5, true, true, NOW(), NOW());
```

---

## Monitoring & Analytics

### View All Surge Recordings
```sql
SELECT 
  sr.ride_id,
  sr.applied_multiplier,
  sr.base_amount,
  sr.surge_amount,
  sr.reason,
  sr.time_based_multiplier,
  sr.demand_based_multiplier,
  sr.created_at
FROM surge_history sr
WHERE DATE(sr.created_at) = CURRENT_DATE
ORDER BY sr.created_at DESC;
```

### Surge Impact Analysis
```sql
SELECT 
  DATE(created_at) as date,
  ROUND(AVG(applied_multiplier), 2) as avg_surge,
  COUNT(*) as rides_with_surge,
  SUM(surge_amount) as total_surge_revenue
FROM surge_history
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

---

## Next Steps

1. **Create initial surge rules** using the SQL examples above
2. **Test ride creation** with different times/demand levels
3. **Monitor `surge_history`** to validate calculations
4. **Adjust rules** based on actual demand patterns
5. **Build admin UI** for non-technical operations team

