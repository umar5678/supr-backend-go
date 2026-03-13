-- Rollback: Revert driver profile is_verified default to true
ALTER TABLE driver_profiles ALTER COLUMN is_verified SET DEFAULT true;
