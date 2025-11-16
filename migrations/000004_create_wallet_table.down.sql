-- Drop triggers
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;

-- Drop tables
DROP TABLE IF EXISTS wallet_holds;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS wallets;

-- Drop types
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS wallet_type;