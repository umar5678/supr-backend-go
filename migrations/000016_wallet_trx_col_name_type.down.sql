-- Revert column name change
ALTER TABLE wallet_transactions 
RENAME COLUMN type TO transaction_type;

-- Revert index
DROP INDEX IF EXISTS idx_wallet_transactions_type;
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(transaction_type);