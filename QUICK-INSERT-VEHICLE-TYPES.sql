-- Quick Vehicle Types Insert Script
-- Copy and paste these into your PostgreSQL database

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active) 
VALUES ('bike', 'Bike', 1.50, 0.50, 0.10, 0.25, 1, 'Fast and affordable bike service', TRUE);

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active) 
VALUES ('economy', 'Economy', 2.50, 0.75, 0.15, 0.50, 4, 'Budget-friendly ride', TRUE);

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active) 
VALUES ('premium', 'Premium', 4.00, 1.25, 0.25, 1.00, 4, 'Premium comfort ride', TRUE);

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active) 
VALUES ('suv', 'SUV', 5.00, 1.50, 0.30, 1.50, 6, 'Spacious SUV', TRUE);

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active) 
VALUES ('luxury', 'Luxury', 8.00, 2.00, 0.50, 2.00, 4, 'Luxury experience', TRUE);

-- Verify
SELECT name, display_name, base_fare, per_km_rate, capacity FROM vehicle_types WHERE name IN ('bike', 'economy', 'premium', 'suv', 'luxury');
