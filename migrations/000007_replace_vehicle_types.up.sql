-- Delete dependent records first (vehicles and rides reference vehicle_types)
DELETE FROM vehicles WHERE vehicle_type_id IS NOT NULL;
DELETE FROM rides WHERE vehicle_type_id IS NOT NULL;
DELETE FROM vehicle_types;

-- Insert the new vehicle types with exact names and base fares from the UI
INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active)
VALUES
  ('bike', 'Bike', 87, 5.0, 1.0, 0, 1, 'Fast and affordable bike service', TRUE),
  ('auto', 'Auto', 152.5, 8.0, 1.5, 0, 3, 'Hassle-free Auto rides for 3 passengers', TRUE),
  ('cab_economy', 'Cab Economy', 185, 15.0, 2.0, 0, 4, 'Budget-friendly cab rides', TRUE),
  ('cab_sedan', 'Cab Sedan', 218, 18.0, 2.5, 0, 4, 'Comfortable sedan cab service', TRUE),
  ('cab_premium', 'Cab Premium', 269, 22.0, 3.0, 0, 4, 'Premium sedan with enhanced comfort', TRUE),
  ('cab_xl', 'Cab XL', 348, 28.0, 3.5, 0, 6, 'Spacious cab for larger groups', TRUE),
  ('scooty', 'Scooty', 95, 4.0, 0.8, 0, 1, 'Quick and eco-friendly scooter ride', TRUE);
