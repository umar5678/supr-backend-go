-- Rollback for cash-based driver wallet migration

-- Drop new indexes
DROP INDEX IF EXISTS idx_wallet_transactions_reference_type;
DROP INDEX IF EXISTS idx_wallet_transactions_commission;
DROP INDEX IF EXISTS idx_wallet_transactions_penalty;
DROP INDEX IF EXISTS idx_wallet_transactions_subscription;

-- Remove added columns (if they were new)
ALTER TABLE wallet_transactions
DROP COLUMN IF EXISTS payment_method,
DROP COLUMN IF EXISTS payment_method_created;
