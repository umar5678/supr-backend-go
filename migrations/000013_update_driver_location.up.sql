-- 1. Drop old table (if exists)
DROP TABLE IF EXISTS driver_locations_history;

-- 2. Create new driver_locations table
CREATE TABLE driver_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES driver_profiles(id) ON DELETE CASCADE,
    location geometry(Point,4326) NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    heading INTEGER DEFAULT 0,
    speed DECIMAL(6,2) DEFAULT 0,
    accuracy DECIMAL(6,2) DEFAULT 0,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_driver_locations_driver_id ON driver_locations(driver_id);
CREATE INDEX idx_driver_locations_timestamp ON driver_locations(timestamp);
CREATE INDEX idx_driver_locations_location ON driver_locations USING GIST (location);
