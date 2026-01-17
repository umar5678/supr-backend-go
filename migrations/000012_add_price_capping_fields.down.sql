-- Remove driver_fare and rider_fare columns (rollback price capping)
ALTER TABLE rides
DROP COLUMN IF EXISTS driver_fare,
DROP COLUMN IF EXISTS rider_fare;
