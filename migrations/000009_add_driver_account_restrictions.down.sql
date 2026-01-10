-- Rollback for driver account restrictions migration

-- Drop audit table
DROP TABLE IF EXISTS driver_balance_audit;

-- Drop indexes
DROP INDEX IF EXISTS idx_driver_profiles_account_status;
DROP INDEX IF EXISTS idx_driver_profiles_is_restricted;
DROP INDEX IF EXISTS idx_driver_profiles_restriction_reason;
DROP INDEX IF EXISTS idx_balance_audit_driver_id;
DROP INDEX IF EXISTS idx_balance_audit_user_id;
DROP INDEX IF EXISTS idx_balance_audit_action;
DROP INDEX IF EXISTS idx_balance_audit_created_at;
DROP INDEX IF EXISTS idx_balance_audit_triggered_restriction;

-- Remove added columns
ALTER TABLE driver_profiles
DROP COLUMN IF EXISTS account_status,
DROP COLUMN IF EXISTS is_restricted,
DROP COLUMN IF EXISTS restricted_at,
DROP COLUMN IF EXISTS restriction_reason,
DROP COLUMN IF EXISTS min_balance_threshold;
