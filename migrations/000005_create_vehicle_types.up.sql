CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


-- Create vehicle_types table
CREATE TABLE vehicle_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    base_fare DECIMAL(10,2) NOT NULL,
    per_km_rate DECIMAL(10,2) NOT NULL,
    per_minute_rate DECIMAL(10,2) NOT NULL,
    booking_fee DECIMAL(10,2) NOT NULL DEFAULT 0.50,
    capacity INTEGER NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    icon_url VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_vehicle_types_name ON vehicle_types(name);
CREATE INDEX idx_vehicle_types_is_active ON vehicle_types(is_active);
CREATE INDEX idx_vehicle_types_deleted_at ON vehicle_types(deleted_at);

-- Seed default vehicle types
INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description) VALUES
('economy', 'Economy', 2.50, 1.20, 0.30, 0.50, 4, 'Affordable rides for everyday travel'),
('comfort', 'Comfort', 3.50, 1.50, 0.40, 0.75, 4, 'More spacious and comfortable rides'),
('premium', 'Premium', 5.00, 2.00, 0.60, 1.00, 4, 'Luxury vehicles for a premium experience'),
('xl', 'XL', 4.00, 1.80, 0.50, 0.75, 6, 'Extra space for groups or luggage'),
('bike', 'Bike', 1.50, 0.80, 0.20, 0.25, 1, 'Quick and affordable motorcycle rides');