-- Rollback wallet_type fixes
-- This removes the constraints and indexes we added

DROP INDEX IF EXISTS idx_wallets_user_type;

ALTER TABLE wallets 
DROP CONSTRAINT IF EXISTS unique_user_wallet_type;

-- Note: We cannot rollback the deletion of invalid wallets
-- Those records are gone, which is intentional (they were broken data)
-- To properly rollback, you would need to restore from backup
