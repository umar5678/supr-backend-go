-- Rollback service provider profile is_verified default
ALTER TABLE service_provider_profiles ALTER COLUMN is_verified SET DEFAULT true;

-- Revert service providers back to verified
UPDATE service_provider_profiles 
SET is_verified = true 
WHERE is_verified = false AND created_at > NOW() - INTERVAL '30 days';
