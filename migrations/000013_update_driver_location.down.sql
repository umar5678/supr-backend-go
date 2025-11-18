-- Reverse: drop the new table and recreate the old one
DROP TABLE IF EXISTS driver_locations;

CREATE TABLE driver_locations_history (
    id BIGSERIAL PRIMARY KEY,
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
CREATE INDEX idx_driver_locations_history_driver_id ON driver_locations_history(driver_id);
CREATE INDEX idx_driver_locations_history_timestamp ON driver_locations_history(timestamp);
CREATE INDEX idx_driver_locations_history_location ON driver_locations_history USING GIST (location);
