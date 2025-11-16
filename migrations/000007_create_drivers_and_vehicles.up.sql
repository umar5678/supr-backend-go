CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable PostGIS for location tracking
-- CREATE EXTENSION IF NOT EXISTS postgis;

-- Create driver_profiles table
CREATE TABLE driver_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    license_number VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'offline' CHECK (status IN ('offline', 'online', 'busy', 'on_trip')),
    current_location geometry(Point, 4326),
    heading INTEGER DEFAULT 0 CHECK (heading >= 0 AND heading <= 360),
    rating DECIMAL(3,2) DEFAULT 5.0 CHECK (rating >= 0 AND rating <= 5),
    total_trips INTEGER DEFAULT 0,
    total_earnings DECIMAL(10,2) DEFAULT 0,
    acceptance_rate DECIMAL(5,2) DEFAULT 100.0,
    cancellation_rate DECIMAL(5,2) DEFAULT 0.0,
    is_verified BOOLEAN DEFAULT true,
    wallet_balance DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create vehicles table
CREATE TABLE vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID UNIQUE NOT NULL REFERENCES driver_profiles(id) ON DELETE CASCADE,
    vehicle_type_id UUID NOT NULL REFERENCES vehicle_types(id),
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL CHECK (year >= 1900 AND year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1),
    color VARCHAR(50) NOT NULL,
    license_plate VARCHAR(50) UNIQUE NOT NULL,
    capacity INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for driver_profiles
CREATE INDEX idx_driver_profiles_user_id ON driver_profiles(user_id);
CREATE INDEX idx_driver_profiles_status ON driver_profiles(status);
CREATE INDEX idx_driver_profiles_location ON driver_profiles USING GIST (current_location);
CREATE INDEX idx_driver_profiles_is_verified ON driver_profiles(is_verified);
CREATE INDEX idx_driver_profiles_deleted_at ON driver_profiles(deleted_at);

-- Create indexes for vehicles
CREATE INDEX idx_vehicles_driver_id ON vehicles(driver_id);
CREATE INDEX idx_vehicles_vehicle_type_id ON vehicles(vehicle_type_id);
CREATE INDEX idx_vehicles_is_active ON vehicles(is_active);
CREATE INDEX idx_vehicles_deleted_at ON vehicles(deleted_at);