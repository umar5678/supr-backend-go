-- Diagnostic: Check all drivers in the system
SELECT 
  dp.id,
  dp.user_id,
  u.first_name,
  dp.status,
  dp.is_verified,
  dp.rating,
  dp.total_rides,
  ST_Y(dp.current_location) as lat,
  ST_X(dp.current_location) as lng,
  CASE WHEN dp.current_location IS NOT NULL THEN 'HAS LOCATION' ELSE 'NO LOCATION' END as location_status,
  v.license_plate,
  vt.name as vehicle_type
FROM driver_profiles dp
LEFT JOIN users u ON u.id = dp.user_id
LEFT JOIN vehicles v ON v.driver_id = dp.id
LEFT JOIN vehicle_types vt ON vt.id = v.vehicle_type_id
ORDER BY dp.created_at DESC
LIMIT 20;

-- Check online+verified drivers specifically
SELECT COUNT(*) as online_verified_drivers
FROM driver_profiles
WHERE status = 'online' AND is_verified = true;

-- Check for any drivers in a general area (broader search)
SELECT 
  id,
  status,
  is_verified,
  ST_Y(current_location) as lat,
  ST_X(current_location) as lng
FROM driver_profiles
WHERE current_location IS NOT NULL
LIMIT 10;
