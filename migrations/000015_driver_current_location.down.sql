-- +goose Down
-- Revert current_location to NOT NULL (fallback POINT(0 0))

UPDATE driver_profiles
SET current_location = ST_GeomFromText('POINT(0 0)', 4326)
WHERE current_location IS NULL;

ALTER TABLE driver_profiles
    ALTER COLUMN current_location SET NOT NULL;

ALTER TABLE driver_profiles
    ALTER COLUMN current_location TYPE geometry(Point,4326)
        USING current_location::geometry;
