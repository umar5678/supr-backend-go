-- CREATE EXTENSION IF NOT EXISTS postgis; , dont include it


-- Create driver_locations_history table
CREATE TABLE driver_locations_history (
    id BIGSERIAL PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES driver_profiles(id) ON DELETE CASCADE,
    location geometry(Point, 4326) NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    heading INTEGER DEFAULT 0 CHECK (heading >= 0 AND heading <= 360),
    speed DECIMAL(6,2) DEFAULT 0 CHECK (speed >= 0),
    accuracy DECIMAL(6,2) DEFAULT 0 CHECK (accuracy >= 0),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_driver_locations_driver_id ON driver_locations_history(driver_id);
CREATE INDEX idx_driver_locations_timestamp ON driver_locations_history(timestamp DESC);
CREATE INDEX idx_driver_locations_location ON driver_locations_history USING GIST (location);
CREATE INDEX idx_driver_locations_driver_timestamp ON driver_locations_history(driver_id, timestamp DESC);

-- Composite index for common queries
CREATE INDEX idx_driver_locations_composite ON driver_locations_history(driver_id, timestamp DESC, location);

-- Comment
COMMENT ON TABLE driver_locations_history IS 'Stores historical location data for drivers';
COMMENT ON COLUMN driver_locations_history.location IS 'PostGIS geometry point (longitude, latitude)';
COMMENT ON COLUMN driver_locations_history.speed IS 'Speed in km/h';
COMMENT ON COLUMN driver_locations_history.accuracy IS 'GPS accuracy in meters';
