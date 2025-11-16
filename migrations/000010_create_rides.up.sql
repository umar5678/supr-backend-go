-- Create rides table
CREATE TABLE rides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    driver_id UUID REFERENCES driver_profiles(id) ON DELETE SET NULL,
    vehicle_type_id UUID NOT NULL REFERENCES vehicle_types(id),
    status VARCHAR(50) NOT NULL CHECK (status IN ('searching', 'accepted', 'arrived', 'started', 'completed', 'cancelled')),
    
    -- Locations
    pickup_location geometry(Point, 4326) NOT NULL,
    pickup_lat DECIMAL(10,8) NOT NULL,
    pickup_lon DECIMAL(11,8) NOT NULL,
    pickup_address TEXT,
    
    dropoff_location geometry(Point, 4326) NOT NULL,
    dropoff_lat DECIMAL(10,8) NOT NULL,
    dropoff_lon DECIMAL(11,8) NOT NULL,
    dropoff_address TEXT,
    
    -- Estimates
    estimated_distance DECIMAL(10,2),
    estimated_duration INTEGER,
    estimated_fare DECIMAL(10,2),
    
    -- Actuals
    actual_distance DECIMAL(10,2),
    actual_duration INTEGER,
    actual_fare DECIMAL(10,2),
    
    -- Pricing
    surge_multiplier DECIMAL(3,2) DEFAULT 1.0,
    
    -- Wallet
    wallet_hold_id UUID,
    
    -- Notes
    rider_notes TEXT,
    cancellation_reason TEXT,
    cancelled_by VARCHAR(50),
    
    -- Timestamps
    requested_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    arrived_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create ride_requests table
CREATE TABLE ride_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL REFERENCES driver_profiles(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'expired', 'rejected', 'accepted')),
    sent_at TIMESTAMP NOT NULL,
    responded_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    rejection_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for rides
CREATE INDEX idx_rides_rider_id ON rides(rider_id);
CREATE INDEX idx_rides_driver_id ON rides(driver_id);
CREATE INDEX idx_rides_vehicle_type_id ON rides(vehicle_type_id);
CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_requested_at ON rides(requested_at DESC);
CREATE INDEX idx_rides_deleted_at ON rides(deleted_at);
CREATE INDEX idx_rides_pickup_location ON rides USING GIST (pickup_location);
CREATE INDEX idx_rides_dropoff_location ON rides USING GIST (dropoff_location);

-- Composite indexes
CREATE INDEX idx_rides_rider_status ON rides(rider_id, status);
CREATE INDEX idx_rides_driver_status ON rides(driver_id, status) WHERE driver_id IS NOT NULL;

-- Indexes for ride_requests
CREATE INDEX idx_ride_requests_ride_id ON ride_requests(ride_id);
CREATE INDEX idx_ride_requests_driver_id ON ride_requests(driver_id);
CREATE INDEX idx_ride_requests_status ON ride_requests(status);
CREATE INDEX idx_ride_requests_expires_at ON ride_requests(expires_at);
CREATE INDEX idx_ride_requests_deleted_at ON ride_requests(deleted_at);
