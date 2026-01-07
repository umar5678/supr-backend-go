-- Migration: Add surge pricing rules, demand tracking, and ETA estimates
-- Version: 000006

-- =====================================================
-- SURGE PRICING RULES TABLE
-- =====================================================
CREATE TABLE surge_pricing_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    vehicle_type_id UUID,
    
    -- Time-based surge
    day_of_week INTEGER DEFAULT -1, -- -1 = all days, 0 = Sunday, etc.
    start_time TIME,
    end_time TIME,
    
    -- Surge multipliers
    base_multiplier DECIMAL(3,2) NOT NULL DEFAULT 1.0,
    min_multiplier DECIMAL(3,2) NOT NULL DEFAULT 1.0,
    max_multiplier DECIMAL(3,2) NOT NULL DEFAULT 2.0,
    
    -- Demand-based surge
    enable_demand_based_surge BOOLEAN DEFAULT FALSE,
    demand_threshold INTEGER DEFAULT 10,
    demand_multiplier_per_request DECIMAL(3,2) DEFAULT 0.05,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_surge_pricing_rules_vehicle_type ON surge_pricing_rules(vehicle_type_id);
CREATE INDEX idx_surge_pricing_rules_day_time ON surge_pricing_rules(day_of_week, start_time, end_time);
CREATE INDEX idx_surge_pricing_rules_is_active ON surge_pricing_rules(is_active);
CREATE INDEX idx_surge_pricing_rules_deleted_at ON surge_pricing_rules(deleted_at);

-- =====================================================
-- DEMAND TRACKING TABLE
-- =====================================================
CREATE TABLE demand_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    zone_id UUID NOT NULL,
    zone_geohash VARCHAR(12),
    
    -- Demand metrics
    pending_requests INTEGER DEFAULT 0,
    available_drivers INTEGER DEFAULT 0,
    completed_rides INTEGER DEFAULT 0,
    average_wait_time INTEGER DEFAULT 0, -- in seconds
    
    -- Calculated metrics
    demand_supply_ratio DECIMAL(5,2),
    surge_multiplier DECIMAL(3,2) DEFAULT 1.0,
    
    -- Timestamps
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_demand_tracking_zone FOREIGN KEY (zone_id) REFERENCES surge_pricing_zones(id) ON DELETE CASCADE
);

CREATE INDEX idx_demand_tracking_zone ON demand_tracking(zone_id);
CREATE INDEX idx_demand_tracking_geohash ON demand_tracking(zone_geohash);
CREATE INDEX idx_demand_tracking_recorded_at ON demand_tracking(recorded_at);
CREATE INDEX idx_demand_tracking_expires_at ON demand_tracking(expires_at);
CREATE INDEX idx_demand_tracking_deleted_at ON demand_tracking(deleted_at);

-- =====================================================
-- ETA ESTIMATES TABLE
-- =====================================================
CREATE TABLE eta_estimates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID,
    
    -- Route information
    pickup_lat DECIMAL(10,8) NOT NULL,
    pickup_lon DECIMAL(11,8) NOT NULL,
    dropoff_lat DECIMAL(10,8) NOT NULL,
    dropoff_lon DECIMAL(11,8) NOT NULL,
    
    -- Distance and duration
    distance_km DECIMAL(10,2) NOT NULL,
    duration_seconds INTEGER NOT NULL,
    
    -- ETA
    estimated_pickup_eta INTEGER NOT NULL,     -- seconds
    estimated_dropoff_eta INTEGER NOT NULL,   -- seconds
    
    -- Traffic conditions
    traffic_condition VARCHAR(50) DEFAULT 'normal',
    traffic_multiplier DECIMAL(3,2) DEFAULT 1.0,
    
    -- Source
    source VARCHAR(50), -- 'google_maps', 'osrm', 'calculated'
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_eta_estimates_ride ON eta_estimates(ride_id);
CREATE INDEX idx_eta_estimates_created_at ON eta_estimates(created_at);
CREATE INDEX idx_eta_estimates_deleted_at ON eta_estimates(deleted_at);

-- =====================================================
-- SURGE HISTORY TABLE
-- =====================================================
CREATE TABLE surge_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID,
    zone_id UUID,
    
    -- Surge details
    applied_multiplier DECIMAL(3,2) NOT NULL,
    base_amount DECIMAL(10,2) NOT NULL,
    surge_amount DECIMAL(10,2) NOT NULL,
    
    -- Reason for surge
    reason VARCHAR(255),
    
    -- Contributing factors
    time_based_multiplier DECIMAL(3,2),
    demand_based_multiplier DECIMAL(3,2),
    pending_requests INTEGER,
    available_drivers INTEGER,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_surge_history_ride ON surge_history(ride_id);
CREATE INDEX idx_surge_history_zone ON surge_history(zone_id);
CREATE INDEX idx_surge_history_created_at ON surge_history(created_at);
CREATE INDEX idx_surge_history_deleted_at ON surge_history(deleted_at);

-- =====================================================
-- TRIGGERS
-- =====================================================
CREATE TRIGGER update_surge_pricing_rules_updated_at BEFORE UPDATE ON surge_pricing_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_demand_tracking_updated_at BEFORE UPDATE ON demand_tracking
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_eta_estimates_updated_at BEFORE UPDATE ON eta_estimates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_surge_history_updated_at BEFORE UPDATE ON surge_history
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
