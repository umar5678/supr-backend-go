CREATE TABLE surge_pricing_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    area_name VARCHAR(255) NOT NULL,
    area_geohash VARCHAR(12) NOT NULL,
    center_lat DECIMAL(10,8) NOT NULL,
    center_lon DECIMAL(11,8) NOT NULL,
    radius_km DECIMAL(6,2) NOT NULL,
    multiplier DECIMAL(3,2) NOT NULL DEFAULT 1.0 CHECK (multiplier >= 1.0 AND multiplier <= 5.0),
    active_from TIMESTAMP NOT NULL,
    active_until TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_surge_zones_geohash ON surge_pricing_zones(area_geohash);
CREATE INDEX idx_surge_zones_active ON surge_pricing_zones(is_active, active_from, active_until);
CREATE INDEX idx_surge_zones_deleted_at ON surge_pricing_zones(deleted_at);

-- Seed some demo surge zones
INSERT INTO surge_pricing_zones (area_name, area_geohash, center_lat, center_lon, radius_km, multiplier, active_from, active_until, is_active)
VALUES
('Downtown Peak Hours', 'ttdhkp9x', 30.3753, 69.3451, 3.0, 1.5, NOW(), NOW() + INTERVAL '365 days', true),
('Airport Zone', 'ttdhkp8y', 30.3850, 69.3550, 2.0, 1.8, NOW(), NOW() + INTERVAL '365 days', true);
