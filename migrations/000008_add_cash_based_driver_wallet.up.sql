-- =====================================================
-- CASH-BASED WALLET MODEL FOR DRIVERS
-- =====================================================
-- This migration adds support for driver wallet operations in cash-based model:
-- 1. Commission tracking (deducted separately from earnings)
-- 2. Penalty tracking (cancellations, rule violations)
-- 3. Subscription fees (future: premium features, insurance)
-- 4. Payment method field for better transaction tracking

-- Add payment_method to wallet_transactions if not exists
ALTER TABLE wallet_transactions
ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT 'wallet',
ADD COLUMN IF NOT EXISTS payment_method_created BOOLEAN DEFAULT FALSE;

-- Mark this column as already existing in wallet_transactions
UPDATE wallet_transactions 
SET payment_method_created = TRUE 
WHERE payment_method_created = FALSE;

-- Create index for transaction tracking by reference_type
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_reference_type 
ON wallet_transactions(reference_type);

-- Create index for easy commission tracking
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_commission 
ON wallet_transactions(reference_type) 
WHERE reference_type = 'ride_commission';

-- Create index for easy penalty tracking
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_penalty 
ON wallet_transactions(reference_type) 
WHERE reference_type = 'driver_penalty';

-- Create index for easy subscription tracking
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_subscription 
ON wallet_transactions(reference_type) 
WHERE reference_type = 'subscription_fee';

-- Add comments for clarity
COMMENT ON COLUMN wallet_transactions.reference_type IS 'Type of transaction: ride_earnings, ride_commission, driver_penalty, subscription_fee, cash_settlement, refund, etc.';
COMMENT ON COLUMN wallet_transactions.payment_method IS 'Payment method used: wallet (virtual), cash, card, etc.';
COMMENT ON COLUMN wallets.balance IS 'Current balance in driver wallet. For drivers: tracks cash + commissions + penalties';
