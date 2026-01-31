-- Drop the old UNIQUE constraint on referral_code
-- This constraint fails when multiple rows have NULL or empty values
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_referral_code_key;

-- Recreate the UNIQUE constraint with NULLS NOT DISTINCT
-- This allows multiple NULL values but ensures non-NULL values are unique
ALTER TABLE users ADD CONSTRAINT users_referral_code_unique UNIQUE (referral_code) WHERE referral_code IS NOT NULL;
