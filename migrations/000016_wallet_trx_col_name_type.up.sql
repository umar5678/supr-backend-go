-- Fix column name mismatch in wallet_transactions table
-- Rename 'transaction_type' to 'type' to match GORM model

ALTER TABLE wallet_transactions 
RENAME COLUMN transaction_type TO type;

-- Update index name for consistency
DROP INDEX IF EXISTS idx_wallet_transactions_type;
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(type);

-- Add comment for clarity
COMMENT ON COLUMN wallet_transactions.type IS 'Transaction type: credit, debit, refund, hold, release, transfer';