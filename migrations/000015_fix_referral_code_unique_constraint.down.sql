-- Rollback the migration
-- Recreate the original UNIQUE constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_referral_code_unique;

-- Recreate with the original definition
ALTER TABLE users ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);
