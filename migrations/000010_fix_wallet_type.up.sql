-- Fix wallet_type for existing wallets
-- For users without a wallet, set them as 'rider' (most common case)
-- This migration assumes most users in the system are riders

-- Update NULL or empty wallet_type values to 'rider' (default)
UPDATE wallets 
SET wallet_type = 'rider'::wallet_type
WHERE wallet_type IS NULL OR wallet_type = '';

-- Add NOT NULL constraint if not already present
ALTER TABLE wallets 
ALTER COLUMN wallet_type SET NOT NULL,
ALTER COLUMN wallet_type SET DEFAULT 'rider'::wallet_type;

-- Add unique constraint to prevent duplicate wallets per user per type
ALTER TABLE wallets 
ADD CONSTRAINT unique_user_wallet_type UNIQUE (user_id, wallet_type);

-- Create index for faster lookups
CREATE INDEX idx_wallets_user_type ON wallets(user_id, wallet_type);
