-- Rollback wallet_type fixes
DROP INDEX IF EXISTS idx_wallets_user_type;

ALTER TABLE wallets 
DROP CONSTRAINT IF EXISTS unique_user_wallet_type;

ALTER TABLE wallets 
ALTER COLUMN wallet_type DROP NOT NULL,
ALTER COLUMN wallet_type DROP DEFAULT;
