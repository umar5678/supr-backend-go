-- Drop the old UNIQUE constraint on referral_code
-- This constraint fails when multiple rows have NULL or empty values
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_referral_code_key;

-- Create a partial UNIQUE index that allows multiple NULL values
-- but ensures non-NULL values are unique
CREATE UNIQUE INDEX idx_users_referral_code_unique ON users (referral_code) WHERE referral_code IS NOT NULL;

