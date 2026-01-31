-- Rollback the migration
-- Drop the partial unique index
DROP INDEX IF EXISTS idx_users_referral_code_unique;

-- Recreate the original UNIQUE constraint
ALTER TABLE users ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);

