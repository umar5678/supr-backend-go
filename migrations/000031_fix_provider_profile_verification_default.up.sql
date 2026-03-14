-- Fix service provider profile is_verified default to false
-- Service providers should not be auto-verified on creation, only after document verification
ALTER TABLE service_provider_profiles ALTER COLUMN is_verified SET DEFAULT false;

-- Update existing service providers that were auto-verified to not verified
-- This ensures all existing service providers without documents are re-verified
UPDATE service_provider_profiles 
SET is_verified = false 
WHERE is_verified = true AND created_at > NOW() - INTERVAL '30 days';
