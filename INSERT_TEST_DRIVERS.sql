-- Insert test users for drivers
INSERT INTO users (id, email, phone, first_name, last_name, user_type, status, created_at, updated_at)
VALUES 
  ('test-driver-1', 'driver1@test.com', '+923001111111', 'Driver', 'One', 'driver', 'active', NOW(), NOW()),
  ('test-driver-2', 'driver2@test.com', '+923001111112', 'Driver', 'Two', 'driver', 'active', NOW(), NOW()),
  ('test-driver-3', 'driver3@test.com', '+923001111113', 'Driver', 'Three', 'driver', 'active', NOW(), NOW())
ON CONFLICT(id) DO NOTHING;

-- Insert vehicle types (if not exists)
INSERT INTO vehicle_types (id, name, base_fare, per_km_rate, per_minute_rate, min_fare, cancellation_fee, is_active, created_at, updated_at)
VALUES 
  ('economy-type', 'Economy', 50, 20, 0.5, 100, 100, true, NOW(), NOW()),
  ('comfort-type', 'Comfort', 75, 25, 0.75, 150, 150, true, NOW(), NOW())
ON CONFLICT(id) DO NOTHING;

-- Insert driver profiles with online status and verified
INSERT INTO driver_profiles (id, user_id, license_number, status, is_verified, rating, total_rides, current_location, created_at, updated_at)
VALUES 
  ('driver-profile-1', 'test-driver-1', 'DL001', 'online', true, 4.8, 150, ST_GeomFromText('POINT(67.0011 24.8607)', 4326), NOW(), NOW()),
  ('driver-profile-2', 'test-driver-2', 'DL002', 'online', true, 4.6, 120, ST_GeomFromText('POINT(67.0050 24.8650)', 4326), NOW(), NOW()),
  ('driver-profile-3', 'test-driver-3', 'DL003', 'online', true, 4.9, 200, ST_GeomFromText('POINT(67.0080 24.8700)', 4326), NOW(), NOW())
ON CONFLICT(id) DO NOTHING;

-- Insert vehicles for drivers
INSERT INTO vehicles (id, driver_id, vehicle_type_id, license_plate, registration_number, color, make_model, year, status, created_at, updated_at)
VALUES 
  ('vehicle-1', 'driver-profile-1', 'economy-type', 'KHI-001', 'REG-001', 'White', 'Honda Civic', 2020, 'active', NOW(), NOW()),
  ('vehicle-2', 'driver-profile-2', 'economy-type', 'KHI-002', 'REG-002', 'Silver', 'Toyota Corolla', 2021, 'active', NOW(), NOW()),
  ('vehicle-3', 'driver-profile-3', 'comfort-type', 'KHI-003', 'REG-003', 'Black', 'Honda Accord', 2022, 'active', NOW(), NOW())
ON CONFLICT(id) DO NOTHING;

SELECT 'Test drivers inserted successfully' AS status;
