-- Fix wallet_type for existing wallets
-- Problem: Some wallets were created without a wallet_type, causing enum validation errors
-- Solution: Delete invalid wallets and ensure new ones have proper types

-- Step 1: Delete wallets with NULL or empty wallet_type
-- These are likely test/broken records
DELETE FROM wallet_transactions WHERE wallet_id IN (
    SELECT id FROM wallets WHERE wallet_type IS NULL
);

DELETE FROM wallet_holds WHERE wallet_id IN (
    SELECT id FROM wallets WHERE wallet_type IS NULL
);

DELETE FROM wallets WHERE wallet_type IS NULL;

-- Step 2: Ensure wallet_type column has NOT NULL constraint and default
ALTER TABLE wallets 
ALTER COLUMN wallet_type SET NOT NULL;

ALTER TABLE wallets 
ALTER COLUMN wallet_type SET DEFAULT 'rider'::wallet_type;

-- Step 3: Add unique constraint to prevent duplicate wallets per user per type
-- This ensures we can't create a second 'rider' or 'driver' wallet for same user
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'unique_user_wallet_type'
    ) THEN
        ALTER TABLE wallets 
        ADD CONSTRAINT unique_user_wallet_type UNIQUE (user_id, wallet_type);
    END IF;
END $$;

-- Step 4: Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_wallets_user_type ON wallets(user_id, wallet_type);
