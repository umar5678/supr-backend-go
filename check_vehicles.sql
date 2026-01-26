SELECT 
  d.id as driver_id,
  d.status,
  d.is_verified,
  v.id as vehicle_id,
  v.vehicle_type_id,
  vt.id as vtype_id,
  vt.name as vehicle_type_name,
  u.id as user_id,
  u.name as user_name
FROM driver_profiles d
LEFT JOIN vehicles v ON d.id = v.driver_id
LEFT JOIN vehicle_types vt ON v.vehicle_type_id = vt.id
LEFT JOIN users u ON d.user_id = u.id
WHERE d.status = 'online' AND d.is_verified = true
SET d.status = 'offline'
LIMIT 10;
