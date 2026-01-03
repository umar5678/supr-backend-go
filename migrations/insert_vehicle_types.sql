-- Insert Vehicle Types
-- This file contains INSERT statements for common vehicle types

-- Bike/Motorcycle
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'bike',
    'Bike',
    1.50,
    0.50,
    0.10,
    0.25,
    1,
    'Fast and affordable bike service',
    TRUE,
    '/icons/bike.png'
) ON CONFLICT (name) DO NOTHING;

-- Economy Car
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'economy',
    'Economy',
    2.50,
    0.75,
    0.15,
    0.50,
    4,
    'Budget-friendly ride with comfort',
    TRUE,
    '/icons/economy.png'
) ON CONFLICT (name) DO NOTHING;

-- Premium/Comfort Car
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'premium',
    'Premium',
    4.00,
    1.25,
    0.25,
    1.00,
    4,
    'Premium comfort ride with air conditioning',
    TRUE,
    '/icons/premium.png'
) ON CONFLICT (name) DO NOTHING;

-- SUV
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'suv',
    'SUV',
    5.00,
    1.50,
    0.30,
    1.50,
    6,
    'Spacious SUV for larger groups',
    TRUE,
    '/icons/suv.png'
) ON CONFLICT (name) DO NOTHING;

-- Luxury
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'luxury',
    'Luxury',
    8.00,
    2.00,
    0.50,
    2.00,
    4,
    'Luxury ride experience with premium amenities',
    TRUE,
    '/icons/luxury.png'
) ON CONFLICT (name) DO NOTHING;

-- Van
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'van',
    'Van',
    6.00,
    1.75,
    0.35,
    1.75,
    8,
    'Large van for groups and luggage',
    TRUE,
    '/icons/van.png'
) ON CONFLICT (name) DO NOTHING;

-- Delivery Van
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'delivery_van',
    'Delivery Van',
    3.00,
    0.90,
    0.20,
    0.75,
    2,
    'Cargo delivery service',
    TRUE,
    '/icons/delivery_van.png'
) ON CONFLICT (name) DO NOTHING;

-- Scooter
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'scooter',
    'Scooter',
    1.00,
    0.40,
    0.08,
    0.20,
    1,
    'Quick and easy scooter rides',
    TRUE,
    '/icons/scooter.png'
) ON CONFLICT (name) DO NOTHING;

-- Rickshaw (Auto)
INSERT INTO vehicle_types (
    name,
    display_name,
    base_fare,
    per_km_rate,
    per_minute_rate,
    booking_fee,
    capacity,
    description,
    is_active,
    icon_url
) VALUES (
    'auto',
    'Auto/Rickshaw',
    1.80,
    0.60,
    0.12,
    0.30,
    3,
    'Local auto-rickshaw service',
    TRUE,
    '/icons/auto.png'
) ON CONFLICT (name) DO NOTHING;

-- Print confirmation
SELECT COUNT(*) as total_vehicle_types, string_agg(name, ', ') as vehicle_names 
FROM vehicle_types 
WHERE name IN ('bike', 'economy', 'premium', 'suv', 'luxury', 'van', 'delivery_van', 'scooter', 'auto');
