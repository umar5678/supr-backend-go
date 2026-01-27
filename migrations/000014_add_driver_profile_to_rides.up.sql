-- Add driver_profile relationship support to rides table
-- This migration documents the GORM relationship between rides and driver_profiles
-- No schema changes needed as the relationship is defined through existing columns:
-- - rides.driver_id -> driver_profiles.user_id

-- This allows efficient querying of driver profiles when fetching rides
-- The relationship is: Ride.DriverProfile -> DriverProfile (using DriverID as foreign key to UserID)

-- No actual SQL changes required - this is a model-level relationship in GORM
