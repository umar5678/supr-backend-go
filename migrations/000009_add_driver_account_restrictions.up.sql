-- =====================================================
-- DRIVER ACCOUNT RESTRICTIONS & BALANCE MANAGEMENT
-- =====================================================
-- This migration adds driver account restrictions based on wallet balance
-- Features:
-- 1. Driver account status: active, suspended, disabled
-- 2. Auto-disable when balance exceeds negative threshold
-- 3. Requires balance top-up before resuming service
-- 4. Tracks restricted_at timestamp for auditing

-- Add fields to driver_profiles for account restriction
ALTER TABLE driver_profiles
ADD COLUMN IF NOT EXISTS account_status VARCHAR(50) DEFAULT 'active',
ADD COLUMN IF NOT EXISTS is_restricted BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS restricted_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS restriction_reason VARCHAR(255),
ADD COLUMN IF NOT EXISTS min_balance_threshold DECIMAL(10,2) DEFAULT -50.00;

-- Create index for quick restriction checks
CREATE INDEX IF NOT EXISTS idx_driver_profiles_account_status 
ON driver_profiles(account_status);

CREATE INDEX IF NOT EXISTS idx_driver_profiles_is_restricted 
ON driver_profiles(is_restricted);

CREATE INDEX IF NOT EXISTS idx_driver_profiles_restriction_reason
ON driver_profiles(restriction_reason)
WHERE is_restricted = TRUE;

-- Add comments for clarity
COMMENT ON COLUMN driver_profiles.account_status IS 'Driver account status: active, suspended, disabled';
COMMENT ON COLUMN driver_profiles.is_restricted IS 'Flag indicating if driver account is restricted due to negative balance';
COMMENT ON COLUMN driver_profiles.restricted_at IS 'Timestamp when driver account was restricted';
COMMENT ON COLUMN driver_profiles.restriction_reason IS 'Reason for account restriction (e.g., balance_negative, manual_suspension)';
COMMENT ON COLUMN driver_profiles.min_balance_threshold IS 'Minimum balance threshold (usually negative) - if balance goes below this, account gets disabled';

-- Create table for balance audit trail
CREATE TABLE IF NOT EXISTS driver_balance_audit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL,
    user_id UUID NOT NULL,
    previous_balance DECIMAL(10,2) NOT NULL,
    new_balance DECIMAL(10,2) NOT NULL,
    change_amount DECIMAL(10,2) NOT NULL,
    action VARCHAR(100) NOT NULL,
    reason VARCHAR(255),
    triggered_restriction BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_balance_audit_driver FOREIGN KEY (driver_id) REFERENCES driver_profiles(id) ON DELETE CASCADE,
    CONSTRAINT fk_balance_audit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_balance_audit_driver_id 
ON driver_balance_audit(driver_id);

CREATE INDEX IF NOT EXISTS idx_balance_audit_user_id 
ON driver_balance_audit(user_id);

CREATE INDEX IF NOT EXISTS idx_balance_audit_action 
ON driver_balance_audit(action);

CREATE INDEX IF NOT EXISTS idx_balance_audit_created_at 
ON driver_balance_audit(created_at);

CREATE INDEX IF NOT EXISTS idx_balance_audit_triggered_restriction
ON driver_balance_audit(triggered_restriction)
WHERE triggered_restriction = TRUE;

COMMENT ON TABLE driver_balance_audit IS 'Audit trail for driver wallet balance changes - tracks all deductions and credits';
