CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


-- Create wallet type enum
CREATE TYPE wallet_type AS ENUM ('rider', 'driver', 'platform');

-- Create transaction type enum
CREATE TYPE transaction_type AS ENUM (
    'credit',           -- Money added to wallet
    'debit',            -- Money deducted from wallet
    'refund',           -- Money refunded
    'hold',             -- Money held (pending transaction)
    'release',          -- Hold released
    'transfer'          -- Transfer between wallets
);

-- Create transaction status enum
CREATE TYPE transaction_status AS ENUM (
    'pending',
    'completed',
    'failed',
    'cancelled',
    'held',
    'released'
);

-- Create wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_type wallet_type NOT NULL,
    balance DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    held_balance DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT balance_non_negative CHECK (balance >= 0),
    CONSTRAINT held_balance_non_negative CHECK (held_balance >= 0),
    CONSTRAINT unique_user_wallet_type UNIQUE(user_id, wallet_type)
);

-- Create wallet transactions table
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    transaction_type transaction_type NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    balance_before DECIMAL(12, 2) NOT NULL,
    balance_after DECIMAL(12, 2) NOT NULL,
    status transaction_status NOT NULL DEFAULT 'pending',
    reference_type VARCHAR(50),  -- 'ride', 'topup', 'withdrawal', 'commission'
    reference_id UUID,            -- ID of the referenced entity
    description TEXT,
    metadata JSONB,               -- Additional data: {rideId, commission, etc.}
    processed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT amount_positive CHECK (amount > 0)
);

-- Create wallet holds table (for pending transactions)
CREATE TABLE IF NOT EXISTS wallet_holds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    amount DECIMAL(12, 2) NOT NULL,
    reference_type VARCHAR(50) NOT NULL,  -- 'ride', etc.
    reference_id UUID NOT NULL,
    status transaction_status NOT NULL DEFAULT 'held',
    expires_at TIMESTAMP NOT NULL,
    released_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT hold_amount_positive CHECK (amount > 0)
);

-- Create indexes
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_type ON wallets(wallet_type);
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(transaction_type);
CREATE INDEX idx_wallet_transactions_status ON wallet_transactions(status);
CREATE INDEX idx_wallet_transactions_reference ON wallet_transactions(reference_type, reference_id);
CREATE INDEX idx_wallet_transactions_created_at ON wallet_transactions(created_at DESC);
CREATE INDEX idx_wallet_holds_wallet_id ON wallet_holds(wallet_id);
CREATE INDEX idx_wallet_holds_reference ON wallet_holds(reference_type, reference_id);
CREATE INDEX idx_wallet_holds_status ON wallet_holds(status);
CREATE INDEX idx_wallet_holds_expires_at ON wallet_holds(expires_at);

-- Comments
COMMENT ON TABLE wallets IS 'User wallets for different roles';
COMMENT ON TABLE wallet_transactions IS 'All wallet transactions with complete audit trail';
COMMENT ON TABLE wallet_holds IS 'Temporary holds on wallet balance';
COMMENT ON COLUMN wallets.balance IS 'Available balance';
COMMENT ON COLUMN wallets.held_balance IS 'Balance held for pending transactions';
COMMENT ON COLUMN wallet_transactions.metadata IS 'Additional transaction data in JSON format';

-- Trigger for wallets updated_at
CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();