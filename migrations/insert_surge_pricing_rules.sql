-- Initial Surge Pricing Rules Setup
-- This script creates standard surge pricing rules for a typical ride-sharing service
-- Execute this after running main migrations

-- =====================================================
-- PEAK HOURS SURGE - WEEKDAY MORNINGS
-- =====================================================
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Morning Surge',
  'Extra surge Monday-Friday during morning commute (7-10 AM)',
  NULL,
  1,  -- Monday
  '07:00',
  '10:00',
  1.5,
  1.0,
  2.0,
  false,
  10,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Morning Surge',
  'Extra surge Monday-Friday during morning commute (7-10 AM)',
  NULL,
  2,  -- Tuesday
  '07:00',
  '10:00',
  1.5,
  1.0,
  2.0,
  false,
  10,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Morning Surge',
  'Extra surge Monday-Friday during morning commute (7-10 AM)',
  NULL,
  3,  -- Wednesday
  '07:00',
  '10:00',
  1.5,
  1.0,
  2.0,
  false,
  10,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Morning Surge',
  'Extra surge Monday-Friday during morning commute (7-10 AM)',
  NULL,
  4,  -- Thursday
  '07:00',
  '10:00',
  1.5,
  1.0,
  2.0,
  false,
  10,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Morning Surge',
  'Extra surge Monday-Friday during morning commute (7-10 AM)',
  NULL,
  5,  -- Friday
  '07:00',
  '10:00',
  1.5,
  1.0,
  2.0,
  false,
  10,
  0.05,
  true,
  NOW(),
  NOW()
);

-- =====================================================
-- PEAK HOURS SURGE - WEEKDAY EVENINGS
-- =====================================================
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Evening Surge',
  'Extra surge Monday-Friday during evening commute (5-8 PM)',
  NULL,
  1,  -- Monday
  '17:00',
  '20:00',
  1.6,
  1.0,
  2.2,
  true,
  15,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Evening Surge',
  'Extra surge Monday-Friday during evening commute (5-8 PM)',
  NULL,
  2,  -- Tuesday
  '17:00',
  '20:00',
  1.6,
  1.0,
  2.2,
  true,
  15,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Evening Surge',
  'Extra surge Monday-Friday during evening commute (5-8 PM)',
  NULL,
  3,  -- Wednesday
  '17:00',
  '20:00',
  1.6,
  1.0,
  2.2,
  true,
  15,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Evening Surge',
  'Extra surge Monday-Friday during evening commute (5-8 PM)',
  NULL,
  4,  -- Thursday
  '17:00',
  '20:00',
  1.6,
  1.0,
  2.2,
  true,
  15,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekday Evening Surge',
  'Extra surge Monday-Friday during evening commute (5-8 PM)',
  NULL,
  5,  -- Friday
  '17:00',
  '20:00',
  1.6,
  1.0,
  2.2,
  true,
  15,
  0.05,
  true,
  NOW(),
  NOW()
);

-- =====================================================
-- WEEKEND EVENING SURGE
-- =====================================================
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekend Evening Surge',
  'Extra surge Saturday-Sunday evenings (6 PM - Midnight)',
  NULL,
  6,  -- Saturday
  '18:00',
  '23:59',
  1.75,
  1.0,
  2.5,
  true,
  12,
  0.05,
  true,
  NOW(),
  NOW()
);

INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Weekend Evening Surge',
  'Extra surge Saturday-Sunday evenings (6 PM - Midnight)',
  NULL,
  0,  -- Sunday
  '18:00',
  '23:59',
  1.75,
  1.0,
  2.5,
  true,
  12,
  0.05,
  true,
  NOW(),
  NOW()
);

-- =====================================================
-- LATE NIGHT SURGE (ALL DAYS)
-- =====================================================
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Late Night Surge',
  'Extra surge every night from 10 PM to 6 AM',
  NULL,
  -1,  -- All days
  '22:00',
  '05:59',
  1.8,
  1.0,
  2.5,
  true,
  8,
  0.08,
  true,
  NOW(),
  NOW()
);

-- =====================================================
-- PURE DEMAND-BASED SURGE (NO FIXED TIME)
-- =====================================================
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Dynamic Demand Surge',
  'Pure demand-based surge: activates when pending requests exceed available drivers',
  NULL,
  -1,  -- All days
  '00:00',
  '23:59',
  1.0,  -- Base is 1.0 (only demand affects pricing)
  1.0,
  2.5,
  true,
  10,  -- Surge kicks in at 10+ pending requests
  0.05,  -- +5% per request above threshold
  true,
  NOW(),
  NOW()
);

-- =====================================================
-- PREMIUM VEHICLE LATE NIGHT SURGE
-- =====================================================
INSERT INTO surge_pricing_rules (
  id, name, description, vehicle_type_id,
  day_of_week, start_time, end_time,
  base_multiplier, min_multiplier, max_multiplier,
  enable_demand_based_surge, demand_threshold, demand_multiplier_per_request,
  is_active, created_at, updated_at
) VALUES (
  uuid_generate_v4(),
  'Premium Late Night',
  'Higher surge for premium vehicles during night hours',
  (SELECT id FROM vehicle_types WHERE name = 'premium' LIMIT 1),
  -1,  -- All days
  '21:00',
  '06:59',
  2.0,
  1.5,
  3.0,
  true,
  5,
  0.1,
  true,
  NOW(),
  NOW()
);

-- =====================================================
-- VERIFY SETUP
-- =====================================================
SELECT COUNT(*) as total_rules FROM surge_pricing_rules;
SELECT name, base_multiplier, day_of_week, start_time, end_time FROM surge_pricing_rules WHERE is_active = true ORDER BY day_of_week, start_time;
