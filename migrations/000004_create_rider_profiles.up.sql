-- Create rider_profiles table
CREATE TABLE IF NOT EXISTS rider_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    home_address JSONB,
    work_address JSONB,
    preferred_vehicle_type VARCHAR(50),
    rating DECIMAL(3,2) NOT NULL DEFAULT 5.0,
    total_rides INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT rating_range CHECK (rating >= 0 AND rating <= 5),
    CONSTRAINT total_rides_non_negative CHECK (total_rides >= 0)
);

-- Create indexes
CREATE INDEX idx_rider_profiles_user_id ON rider_profiles(user_id);
CREATE INDEX idx_rider_profiles_rating ON rider_profiles(rating DESC);

-- Comments
COMMENT ON TABLE rider_profiles IS 'Rider profile information';
COMMENT ON COLUMN rider_profiles.home_address IS 'JSON: {lat, lng, address}';
COMMENT ON COLUMN rider_profiles.work_address IS 'JSON: {lat, lng, address}';
COMMENT ON COLUMN rider_profiles.preferred_vehicle_type IS 'Preferred vehicle type for rides';

-- Trigger for updated_at
CREATE TRIGGER update_rider_profiles_updated_at
    BEFORE UPDATE ON rider_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();