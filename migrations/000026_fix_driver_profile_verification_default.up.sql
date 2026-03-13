-- Fix driver profile is_verified default to false
-- Drivers should not be auto-verified on creation, only after document verification
ALTER TABLE driver_profiles ALTER COLUMN is_verified SET DEFAULT false;

-- Update existing drivers that were auto-verified to not verified
-- This ensures all existing drivers without documents are re-verified
UPDATE driver_profiles 
SET is_verified = false 
WHERE is_verified = true AND created_at > NOW() - INTERVAL '30 days';
