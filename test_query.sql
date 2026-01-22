SELECT COUNT(*) as total_drivers FROM driver_profiles;
SELECT COUNT(*) as online_drivers FROM driver_profiles WHERE status = 'online';
SELECT COUNT(*) as verified_drivers FROM driver_profiles WHERE is_verified = true;
SELECT COUNT(*) as with_vehicles FROM driver_profiles WHERE id IN (SELECT driver_id FROM vehicles);
SELECT COUNT(*) as with_location FROM driver_profiles WHERE current_location IS NOT NULL;
SELECT COUNT(*) as with_user FROM driver_profiles WHERE user_id IS NOT NULL;
SELECT d.id, d.status, d.is_verified, v.id as vehicle_id, vt.id as vehicle_type_id, u.id as user_id 
FROM driver_profiles d 
LEFT JOIN vehicles v ON d.id = v.driver_id 
LEFT JOIN vehicle_types vt ON v.vehicle_type_id = vt.id
LEFT JOIN users u ON d.user_id = u.id
LIMIT 5;
